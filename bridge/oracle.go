package bridge

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/client"
	"github.com/functionx/fx-tron-bridge/internal/logger"
	"github.com/functionx/fx-tron-bridge/internal/utils"
)

type Oracle struct {
	*FxTronBridge
	lastEventNonce   uint64
	startBlockNumber uint64
}

func NewOracle(fxBridge *FxTronBridge, startBlockNumber uint64) (*Oracle, error) {
	lastBlockNumber, err := fxBridge.CrossChainClient.LastEventBlockHeightByAddr(fxBridge.GetBridgerAddr().String(), fxtronbridge.Tron)
	logger.Infof("new oracle start block number: %d, fx core last block number: %d", startBlockNumber, lastBlockNumber)
	if err != nil {
		logger.Errorf("get last block number fail bridger address: %s, err: %s", fxBridge.GetBridgerAddr().String(), err.Error())
		return nil, err
	}
	cacheBlockNumber, err := readLastBlockNumber()
	if err != nil {
		return nil, err
	}

	if startBlockNumber > lastBlockNumber {
		lastBlockNumber = startBlockNumber
	}

	if lastBlockNumber <= 0 {
		lastBlockNumber, err = getLastBlockNumber(fxBridge.BridgeAddr, fxBridge.TronClient)
		if err != nil {
			return nil, err
		}
	} else if cacheBlockNumber > fxtronbridge.TronRestartDelayBlock {
		cacheBlockNumber = cacheBlockNumber - fxtronbridge.TronRestartDelayBlock
		if cacheBlockNumber > lastBlockNumber {
			lastBlockNumber = cacheBlockNumber
		}
		logger.Infof("read cache last block number: %d, last block number: %d", cacheBlockNumber, lastBlockNumber)
	}

	return &Oracle{
		startBlockNumber: lastBlockNumber,
		FxTronBridge:     fxBridge,
	}, nil
}

func getLastBlockNumber(bridgeAddr string, tronClient *client.TronClient) (uint64, error) {
	latestBlockNumber, err := tronClient.BlockNumber(context.Background())
	if err != nil {
		logger.Errorf("get tron last block number fail err: %s", err.Error())
		return 0, err
	}

	minBlockNumber := latestBlockNumber - 1000
	for i := latestBlockNumber; i > minBlockNumber; i-- {
		logger.Infof("get tron last block number current blockNumber: %d", i)
		oracleSetUpdatedEvents, err := tronClient.QueryOracleSetUpdatedEvent(bridgeAddr, i)
		if err != nil {
			logger.Errorf("query oracle set updated event fail bridgeAddr: %s, blockNumber: %d, err: %s", bridgeAddr, i, err.Error())
			return 0, err
		}
		if len(oracleSetUpdatedEvents) > 0 {
			return i - 1, nil
		}
	}
	return 0, fmt.Errorf("get last block number does not exist oracle set updated events latestBlockNumber: %d, minBlockNumber: %d", latestBlockNumber, minBlockNumber)
}

func (o *Oracle) bridgeEvent() error {
	bridger, err := o.CrossChainClient.GetOracleByBridgerAddr(o.GetBridgerAddr().String(), fxtronbridge.Tron)
	if err != nil {
		logger.Errorf("get oracle by bridger fail bridger: %s, err: %s", o.GetBridgerAddr().String(), err.Error())
		return err
	}
	if bridger.Online {
		logger.Warn("get oracle status is not active bridger: %v", bridger)
		return nil
	}
	lastEventNonce, err := o.CrossChainClient.LastEventNonceByAddr(o.GetBridgerAddr().String(), fxtronbridge.Tron)
	if err != nil {
		logger.Errorf("get last event nonce by addr fail orcAddr: %s, err: %s", o.GetBridgerAddr().String(), err.Error())
		return err
	}
	o.lastEventNonce = lastEventNonce
	latestBlockNumber, err := o.TronClient.BlockNumber(context.Background())
	if err != nil {
		logger.Errorf("get last block number fail err: %s", err.Error())
		return err
	}
	endBlockNumber := latestBlockNumber - fxtronbridge.TronBlockDelay
	logger.Infof("oracle handle event startBlockNumber: %d, endBlockNumber: %d, lastEventNonce: %d", o.startBlockNumber, endBlockNumber, o.lastEventNonce)

	fxtronbridge.BlockHeightProm.Set(float64(o.startBlockNumber))
	if o.startBlockNumber >= endBlockNumber {
		return nil
	}

	blockNumberInterval := endBlockNumber - o.startBlockNumber
	fxtronbridge.BlockIntervalProm.Set(float64(blockNumberInterval))

	if blockNumberInterval > fxtronbridge.TronDelayBlockWarn {
		logger.Warnf("bridge behind too much block number startBlockNumber: %d, endBlockNumber: %d", o.startBlockNumber, endBlockNumber)
	}

	msgs := make([]sdk.Msg, 0)
	batchBlockNumber := 0
	for blockNumber := o.startBlockNumber + 1; blockNumber <= endBlockNumber; blockNumber++ {
		events, err := o.TronClient.QueryBlockEvent(o.BridgeAddr, blockNumber)
		if err != nil {
			logger.Errorf("query block event fail bridgeAddr: %s, blockNumber: %d, err: %s", o.BridgeAddr, blockNumber, err.Error())
			return err
		}
		sort.Slice(events, func(i, j int) bool {
			return events[i].GetEventNonce() < events[j].GetEventNonce()
		})
		for _, event := range events {
			if event.GetEventNonce() <= lastEventNonce {
				continue
			}
			msgs = append(msgs, event.ToMsg(blockNumber, o.GetBridgerAddr().String()))
		}

		if len(msgs) > fxtronbridge.BatchSendMsgCount || batchBlockNumber > 100 || blockNumber == endBlockNumber {
			if err = o.BatchSendMsg(msgs, fxtronbridge.BatchSendMsgCount); err != nil {
				return err
			}
			o.startBlockNumber = blockNumber
			_ = saveLastBlockNumber(o.startBlockNumber)
			batchBlockNumber = 0
			msgs = make([]sdk.Msg, 0)
		}
		batchBlockNumber++
	}

	return nil
}

func readLastBlockNumber() (uint64, error) {
	fileName := path.Join(os.ExpandEnv(fxtronbridge.TronHome), "lastBlockNumber.info")
	isFile, err := utils.PathExists(fileName)
	if err != nil {
		return 0, err
	}
	if !isFile {
		return 0, nil
	}
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return 0, err
	}
	str := string(bytes)
	if len(str) <= 0 {
		return 0, nil
	}
	return strconv.ParseUint(str, 10, 64)
}

func saveLastBlockNumber(lastBlockNumber uint64) error {
	filePath := os.ExpandEnv(fxtronbridge.TronHome)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return err
	}
	fileName := path.Join(filePath, "lastBlockNumber.info")
	return os.WriteFile(fileName, []byte(strconv.FormatUint(lastBlockNumber, 10)), os.ModePerm)
}
