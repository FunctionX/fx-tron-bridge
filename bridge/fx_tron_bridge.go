package bridge

import (
	"context"
	"crypto/ecdsa"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/functionx/fx-tron-bridge/client"
	"github.com/functionx/fx-tron-bridge/fxchain"
	"github.com/functionx/fx-tron-bridge/utils/logger"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"time"
)

type FxTronBridge struct {
	TronClient       *client.TronClient
	CrossChainClient *fxchain.CrossChainClient
	BridgeAddr       string
	OrcAddr          string
	OrcPrivKey       *secp256k1.PrivKey
	TronAddr         string
	TronPrivKey      *ecdsa.PrivateKey
}

func NewFxTronBridge(bridgeAddr, tronGrpc, tronJsonRpc, fxGrpc string, orcPrivKey *secp256k1.PrivKey, tronPrivateKey *ecdsa.PrivateKey) (*FxTronBridge, error) {
	logger.Infof("NewFxTronBridge, bridgeAddr: %s, tronGrpc: %v, tronJsonRpc: %v, fxGrpc: %v", bridgeAddr, tronGrpc, tronJsonRpc, fxGrpc)

	var tronClient *client.TronClient
	var err error
	if tronGrpc != "" && tronJsonRpc != "" {
		tronClient, err = client.NewTronGrpcJsonRpcClient(tronGrpc, tronJsonRpc)
	} else if tronGrpc != "" {
		tronClient, err = client.NewTronGrpcClient(tronGrpc)
	}
	if err != nil {
		return nil, err
	}

	crossChainClient, err := fxchain.NewCrossChainClient(fxGrpc)
	if err != nil {
		return nil, err
	}

	var orcAddr string
	if orcPrivKey != nil {
		orcAddr = sdk.AccAddress(orcPrivKey.PubKey().Address()).String()
		logger.Infof("orc address: %s", orcAddr)
	}
	var tronAddr string
	if tronPrivateKey != nil {
		tronAddr = address.PubkeyToAddress(tronPrivateKey.PublicKey).String()
		logger.Infof("tron address: %s", tronAddr)
	}
	return &FxTronBridge{
		BridgeAddr:       bridgeAddr,
		OrcAddr:          orcAddr,
		OrcPrivKey:       orcPrivKey,
		TronAddr:         tronAddr,
		TronPrivKey:      tronPrivateKey,
		TronClient:       tronClient,
		CrossChainClient: crossChainClient,
	}, nil
}

func (f *FxTronBridge) WaitNewBlock() {
	retryTime := 3 * time.Second
	for {
		tronLastBlockNumber, tronErr := f.TronClient.GetLastBlockNumber()
		block, fxErr := f.CrossChainClient.GetLatestBlock()
		if tronErr == nil && fxErr == nil {
			logger.Infof("withSyncBlock tronLastBlockNumber: %v, fxLastBlockNumber: %v", tronLastBlockNumber, block.GetHeader().Height)

			if f.TronClient.JsonRpcClient != nil {
				jsonRpcBlockNumber, err := f.TronClient.JsonRpcClient.BlockNumber(context.Background())
				if err != nil {
					logger.Warnf("could not reach tron json rpc! err: %v", err)
					time.Sleep(retryTime)
					continue
				}
				logger.Infof("withSyncBlock tron json rpc block number: %v", jsonRpcBlockNumber)
			}

			return
		}
		if tronErr == nil && fxErr != nil {
			logger.Warn("could not contact Fx grpc, trying again fxErr:", fxErr)
			time.Sleep(retryTime)
			continue
		}
		if tronErr != nil && fxErr == nil {
			logger.Warn("could not contact tron grpc, trying again tronErr:", tronErr)
			time.Sleep(retryTime)
			continue
		}
		if tronErr != nil && fxErr != nil {
			logger.Warnf("could not reach tron and FxChain grpc! tronErr: %v, fxErr: %v", tronErr, fxErr)
			time.Sleep(retryTime)
			continue
		}
	}
}
