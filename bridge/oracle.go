package bridge

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/functionx/fx-core/x/crosschain/types"
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/client"
	"github.com/functionx/fx-tron-bridge/contract"
	"github.com/functionx/fx-tron-bridge/utils"
	"github.com/functionx/fx-tron-bridge/utils/logger"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
)

type Oracle struct {
	*FxTronBridge
	lastEventNonce   uint64
	startBlockNumber uint64
}

func NewOracle(fxBridge *FxTronBridge, startBlockNumber uint64) (*Oracle, error) {
	lastBlockNumber, err := fxBridge.CrossChainClient.LastEventBlockHeightByAddr(fxBridge.OrcAddr, fxtronbridge.Tron)
	logger.Infof("New oracle start block number: %v, fx core last block number: %v", startBlockNumber, lastBlockNumber)
	if err != nil {
		logger.Errorf("Get fx core last block number fail orcAddr: %v, err: %v", fxBridge.OrcAddr, err)
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
		logger.Infof("Read cache last block number: %v, last block number: %v", cacheBlockNumber, lastBlockNumber)
	}

	return &Oracle{
		startBlockNumber: lastBlockNumber,
		FxTronBridge:     fxBridge,
	}, nil
}

func getLastBlockNumber(bridgeAddr string, tronClient *client.TronClient) (uint64, error) {
	latestBlockNumber, err := tronClient.GetLastBlockNumber()
	if err != nil {
		logger.Error("Tron get last block number fail err:", err)
		return 0, err
	}

	minBlockNumber := latestBlockNumber - 1000
	for i := latestBlockNumber; i > minBlockNumber; i-- {
		logger.Infof("Get tron last block number current blockNumber: %v", i)
		oracleSetUpdatedEvents, err := tronClient.QueryOracleSetUpdatedEvent(bridgeAddr, i)
		if err != nil {
			logger.Errorf("Tron query oracle set updated event fail bridgeAddr: %v, blockNumber: %v, err: %v", bridgeAddr, i, err)
			return 0, err
		}
		if len(oracleSetUpdatedEvents) > 0 {
			return i - 1, nil
		}
	}
	return 0, fmt.Errorf("get last block number does not exist oracle set updated events latestBlockNumber: %v, minBlockNumber: %v", latestBlockNumber, minBlockNumber)
}

func (o *Oracle) HandleEvent() error {
	orchestrator, err := o.CrossChainClient.GetOracleByOrchestrator(o.FxTronBridge.OrcAddr, fxtronbridge.Tron)
	if err != nil {
		logger.Warnf("Fx core get oracle by orchestrator fail orcAddr: %v, err: %v", o.FxTronBridge.OrcAddr, err)
		return err
	}
	if orchestrator.Jailed {
		logger.Warn("Get orchestrator oracle status is not active orchestrator:", orchestrator)
		return nil
	}

	lastEventNonce, err := o.CrossChainClient.LastEventNonceByAddr(o.OrcAddr, fxtronbridge.Tron)
	if err != nil {
		logger.Warnf("Get fx core last event nonce by addr fail orcAddr: %v, err: %v", o.OrcAddr, err)
		return err
	}
	o.lastEventNonce = lastEventNonce

	latestBlockNumber, err := o.TronClient.GetLastBlockNumber()
	if err != nil {
		logger.Warn("Tron get last block number fail err:", err)
		return err
	}
	endBlockNumber := latestBlockNumber - fxtronbridge.TronBlockDelay

	logger.Infof("Oracle handle event startBlockNumber: %v, endBlockNumber: %v, lastEventNonce: %v", o.startBlockNumber, endBlockNumber, o.lastEventNonce)

	fxtronbridge.BlockHeightProm.Set(float64(o.startBlockNumber))
	if o.startBlockNumber >= endBlockNumber {
		return nil
	}

	blockNumberInterval := endBlockNumber - o.startBlockNumber
	fxtronbridge.BlockIntervalProm.Set(float64(blockNumberInterval))

	if blockNumberInterval > fxtronbridge.TronDelayBlockWarn {
		logger.Warnf("Bridge behind too much block number startBlockNumber: %v, endBlockNumber: %v", o.startBlockNumber, endBlockNumber)
	}

	msgs := make([]Msg, 0)
	batchBlockNumber := 0
	for blockNumber := o.startBlockNumber + 1; blockNumber <= endBlockNumber; blockNumber++ {

		sendToFxEvents, transactionBatchExecutedEvents, addBridgeTokenEvents, oracleSetUpdatedEvents, err := o.TronClient.QueryBlockEvent(o.BridgeAddr, blockNumber)
		if err != nil {
			logger.Warnf("Tron query block event fail bridgeAddr: %v, blockNumber: %v, err: %v", o.BridgeAddr, blockNumber, err)
			return err
		}

		msgsCount := len(msgs)

		msgs = append(msgs, o.setMsgOracleSetUpdatedClaim(oracleSetUpdatedEvents, blockNumber)...)

		msgs = append(msgs, o.setMsgBridgeTokenClaim(addBridgeTokenEvents, blockNumber)...)

		msgs = append(msgs, o.setMsgSendToFxClaim(sendToFxEvents, blockNumber)...)

		msgs = append(msgs, o.setMsgSendToExternalClaim(transactionBatchExecutedEvents, blockNumber)...)

		increaseCount := len(msgs) - msgsCount
		if increaseCount > 0 {
			logger.Infof("Handle event currentBlockNumber: %v, msgsLen: %v", blockNumber, increaseCount)
		}

		if len(msgs) > fxtronbridge.BatchSendMsgCount || batchBlockNumber > 100 || blockNumber == endBlockNumber {

			if err = o.batchSendMsg(msgs, fxtronbridge.BatchSendMsgCount); err != nil {
				return err
			}
			o.startBlockNumber = blockNumber
			_ = saveLastBlockNumber(o.startBlockNumber)
			batchBlockNumber = 0
			msgs = make([]Msg, 0)
		}
		batchBlockNumber++
	}

	return nil
}

func (o *Oracle) batchSendMsg(msgs []Msg, batchNumber int) error {
	if len(msgs) <= 0 {
		return nil
	}

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].EventNonce < msgs[j].EventNonce
	})

	fxtronbridge.MsgPendingLenProm.Add(float64(len(msgs)))

	if msgs[0].EventNonce != o.lastEventNonce+1 {
		panic("Oracle bridge event nonce out of the current state startBlockNumber:" + strconv.FormatUint(o.startBlockNumber, 10))
	}

	batchCount := len(msgs) / batchNumber
	var endIndex int
	for i := 0; i < batchCount; i++ {
		startIndex := i * batchNumber
		endIndex = startIndex + batchNumber
		err := o.sendMsg(msgs[startIndex:endIndex])
		if err != nil {
			return err
		}
	}

	err := o.sendMsg(msgs[endIndex:])
	if err != nil {
		return err
	}
	o.lastEventNonce += uint64(len(msgs))
	return nil
}

func (o *Oracle) sendMsg(msgs []Msg) error {
	if len(msgs) <= 0 {
		return nil
	}
	sdkMsgs := make([]sdk.Msg, len(msgs))
	for i, msg := range msgs {
		sdkMsgs[i] = msg.Msg
	}

	txRaw, err := o.CrossChainClient.BuildTx(o.OrcPrivKey, sdkMsgs)
	if err != nil {
		logger.Warnf("Oracle build tx fail msgsLen: %v, err: %v", len(sdkMsgs), err)
		return err
	}
	txResp, err := o.CrossChainClient.BroadcastTx(txRaw)
	if err != nil {
		logger.Warnf("Oracle broadcast tx fail msgsLen: %v, err: %v", len(sdkMsgs), err)
		return err
	}

	if txResp.Code != 0 {
		logger.Warnf("Oracle send msg fail fxcoreHeight: %v, fxcoreHash: %v, resp code: %v, tronEventInfo: %v", txResp.Height, txResp.TxHash, txResp.Code, MsgsToJson(msgs))
		return fxtronbridge.ErrSendTx
	}
	logger.Infof("Oracle send msg success fxcoreHeight: %v, fxcoreHash: %v, tronEventInfo: %v", txResp.Height, txResp.TxHash, MsgsToJson(msgs))

	return nil
}

func (o *Oracle) setMsgSendToExternalClaim(transactionBatchExecutedEvents []*contract.FxBridgeTronTransactionBatchExecutedEvent, blockNumber uint64) []Msg {
	msgs := make([]Msg, 0)
	for _, transactionBatchExecuted := range transactionBatchExecutedEvents {
		eventNonce := transactionBatchExecuted.EventNonce.Uint64()
		if eventNonce <= o.lastEventNonce {
			continue
		}

		msgs = append(msgs, Msg{
			Height:          blockNumber,
			EventNonce:      eventNonce,
			TransactionHash: transactionBatchExecuted.Raw.TxHash.String(),
			EventName:       "TransactionBatchExecutedEvent",
			Msg: &types.MsgSendToExternalClaim{
				EventNonce:    transactionBatchExecuted.EventNonce.Uint64(),
				BlockHeight:   blockNumber,
				BatchNonce:    transactionBatchExecuted.BatchNonce.Uint64(),
				TokenContract: client.AddressToString(transactionBatchExecuted.Token),
				Orchestrator:  o.OrcAddr,
				ChainName:     fxtronbridge.Tron,
			},
		})
	}
	return msgs
}

func (o *Oracle) setMsgOracleSetUpdatedClaim(oracleSetUpdatedEvents []*contract.FxBridgeTronOracleSetUpdatedEvent, blockNumber uint64) []Msg {
	msgs := make([]Msg, 0)
	for _, oracleSetUpdated := range oracleSetUpdatedEvents {
		eventNonce := oracleSetUpdated.EventNonce.Uint64()
		if eventNonce <= o.lastEventNonce {
			continue
		}

		members := make([]*types.BridgeValidator, len(oracleSetUpdated.Oracles))
		for i, oracleAddress := range oracleSetUpdated.Oracles {
			members[i] = &types.BridgeValidator{
				Power:           oracleSetUpdated.Powers[i].Uint64(),
				ExternalAddress: client.AddressToString(oracleAddress),
			}
		}

		msgs = append(msgs, Msg{
			Height:          blockNumber,
			EventNonce:      eventNonce,
			TransactionHash: oracleSetUpdated.Raw.TxHash.String(),
			EventName:       "OracleSetUpdatedEvent",
			Msg: &types.MsgOracleSetUpdatedClaim{
				EventNonce:     oracleSetUpdated.EventNonce.Uint64(),
				BlockHeight:    blockNumber,
				OracleSetNonce: oracleSetUpdated.NewOracleSetNonce.Uint64(),
				Members:        members,
				Orchestrator:   o.OrcAddr,
				ChainName:      fxtronbridge.Tron,
			},
		})
	}
	return msgs
}

func (o *Oracle) setMsgBridgeTokenClaim(addBridgeTokenEvents []*contract.FxBridgeTronAddBridgeTokenEvent, blockNumber uint64) []Msg {
	msgs := make([]Msg, 0)
	for _, addBridgeToken := range addBridgeTokenEvents {
		eventNonce := addBridgeToken.EventNonce.Uint64()
		if eventNonce <= o.lastEventNonce {
			continue
		}

		msgs = append(msgs, Msg{
			Height:          blockNumber,
			EventNonce:      eventNonce,
			TransactionHash: addBridgeToken.Raw.TxHash.String(),
			EventName:       "AddBridgeTokenEvent",
			Msg: &types.MsgBridgeTokenClaim{
				EventNonce:    eventNonce,
				BlockHeight:   blockNumber,
				TokenContract: client.AddressToString(addBridgeToken.TokenContract),
				Name:          addBridgeToken.Name,
				Symbol:        addBridgeToken.Symbol,
				Decimals:      uint64(addBridgeToken.Decimals),
				Orchestrator:  o.OrcAddr,
				ChannelIbc:    client.HexByte32ToTargetIbc(addBridgeToken.ChannelIBC),
				ChainName:     fxtronbridge.Tron,
			},
		})
	}
	return msgs
}

func (o *Oracle) setMsgSendToFxClaim(sendToFxEvents []*contract.FxBridgeTronSendToFxEvent, blockNumber uint64) []Msg {
	msgs := make([]Msg, 0)
	for _, sendToFx := range sendToFxEvents {
		eventNonce := sendToFx.EventNonce.Uint64()
		if eventNonce <= o.lastEventNonce {
			continue
		}

		msgs = append(msgs, Msg{
			Height:          blockNumber,
			EventNonce:      eventNonce,
			TransactionHash: sendToFx.Raw.TxHash.String(),
			EventName:       "SendToFxEvent",
			Msg: &types.MsgSendToFxClaim{
				EventNonce:    eventNonce,
				BlockHeight:   blockNumber,
				TokenContract: client.AddressToString(sendToFx.TokenContract),
				Amount:        sdk.NewIntFromBigInt(sendToFx.Amount),
				Sender:        client.AddressToString(sendToFx.Sender),
				Receiver:      sdk.AccAddress(sendToFx.Destination[12:]).String(),
				TargetIbc:     client.HexByte32ToTargetIbc(sendToFx.TargetIBC),
				Orchestrator:  o.OrcAddr,
				ChainName:     fxtronbridge.Tron,
			},
		})
	}
	return msgs
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
	bytes, err := ioutil.ReadFile(fileName)
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
	return ioutil.WriteFile(fileName, []byte(strconv.FormatUint(lastBlockNumber, 10)), os.ModePerm)
}

type Msg struct {
	Height          uint64
	EventNonce      uint64
	TransactionHash string
	EventName       string
	sdk.Msg         `json:"-"`
}

func MsgsToJson(msgs []Msg) string {
	data, err := json.Marshal(msgs)
	if err != nil {
		return "json marshal err"
	}
	return string(data)
}
