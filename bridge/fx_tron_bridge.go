package bridge

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/client"
	"github.com/functionx/fx-tron-bridge/fxchain"
	"github.com/functionx/fx-tron-bridge/internal/logger"
	"time"
)

type FxTronBridge struct {
	TronClient       *client.TronClient
	CrossChainClient *fxchain.CrossChainClient
	BridgeAddr       string
	OrcPrivKey       *secp256k1.PrivKey
	TronPrivKey      *ecdsa.PrivateKey
}

func NewFxTronBridge(bridgeAddr, tronGrpc, fxGrpc string, orcPrivKey *secp256k1.PrivKey, tronPrivateKey *ecdsa.PrivateKey) (*FxTronBridge, error) {
	logger.Infof("NewFxTronBridge, bridgeAddr: %s, tronGrpc: %s, fxGrpc: %s", bridgeAddr, tronGrpc, fxGrpc)

	tronClient, err := client.NewTronGrpcClient(tronGrpc)
	if err != nil {
		return nil, err
	}

	crossChainClient, err := fxchain.NewCrossChainClient(fxGrpc)
	if err != nil {
		return nil, err
	}

	return &FxTronBridge{
		BridgeAddr:       bridgeAddr,
		OrcPrivKey:       orcPrivKey,
		TronPrivKey:      tronPrivateKey,
		TronClient:       tronClient,
		CrossChainClient: crossChainClient,
	}, nil
}

func (f *FxTronBridge) GetOrcAddr() sdk.AccAddress {
	return f.OrcPrivKey.PubKey().Address().Bytes()
}

func (f *FxTronBridge) GetTronAddr() address.Address {
	return address.PubkeyToAddress(f.TronPrivKey.PublicKey)
}

func (f *FxTronBridge) setFxKeyBalanceMetrics(fees string) {
	balance, err := f.CrossChainClient.QueryBalance(f.GetOrcAddr(), fees)
	if err != nil {
		logger.Errorf("query balance fail fees: %s, err: %s", fees, err.Error())
		return
	}
	fxtronbridge.FxKeyBalanceProm.Set(float64(balance.Amount.Quo(sdk.NewInt(1e18)).Uint64()))
}

func (f *FxTronBridge) WaitNewBlock() error {
	retryTime := fxtronbridge.FxAvgBlockMillisecond

	var lastTronBlockNumber uint64 = 0
	var lastFxBlockNumber int64 = 0

	for {
		tronBlockNumber, err := f.TronClient.BlockNumber(context.Background())
		if err != nil {
			return err
		}
		if lastTronBlockNumber <= 0 && tronBlockNumber > 0 {
			lastTronBlockNumber = tronBlockNumber
		}
		fxBlock, err := f.CrossChainClient.GetLatestBlock()
		if err != nil {
			return err
		}
		if lastFxBlockNumber <= 0 && fxBlock.Header.Height > 0 {
			lastFxBlockNumber = fxBlock.Header.Height
		}
		if fxBlock.Header.Height-lastFxBlockNumber > 0 && lastTronBlockNumber-tronBlockNumber > 0 {
			logger.Infof("starting external block number: %d, fxCore block height: %d", tronBlockNumber, fxBlock.Header.Height)
			break
		}
		time.Sleep(retryTime)
	}
	return nil
}

func (f *FxTronBridge) BatchSendMsg(msgs []sdk.Msg, batchNumber int) error {
	if len(msgs) <= 0 {
		return nil
	}
	batchCount := len(msgs) / batchNumber
	endIndex := 0
	for i := 0; i < batchCount; i++ {
		startIndex := i * batchNumber
		endIndex = startIndex + batchNumber
		if err := f.sendMsg(msgs[startIndex:endIndex]); err != nil {
			return err
		}
	}
	if len(msgs) > endIndex {
		if err := f.sendMsg(msgs[endIndex:]); err != nil {
			return err
		}
	}
	return nil
}

func (f *FxTronBridge) sendMsg(msgs []sdk.Msg) error {
	txRaw, err := f.CrossChainClient.BuildTx(f.OrcPrivKey, msgs)
	if err != nil {
		logger.Errorf("build tx fail messages len: %d, err: %s", len(msgs), err.Error())
		return err
	}
	txResp, err := f.CrossChainClient.BroadcastTx(txRaw)
	if err != nil {
		logger.Errorf("broadcast tx fail messages len: %d, err: %s", len(msgs), err.Error())
		return err
	}
	if txResp.Code != 0 {
		return fmt.Errorf("send msg fail Height: %d, Hash: %s, code: %d", txResp.Height, txResp.TxHash, txResp.Code)
	} else {
		logger.Infof("send msg success Height: %d, Hash: %s", txResp.Height, txResp.TxHash)
	}
	return nil
}
