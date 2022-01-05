package fxchain

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/functionx/fx-tron-bridge/utils/logger"
	"net/url"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	otherTypes "github.com/functionx/fx-core/x/other/types"
	"github.com/gogo/protobuf/proto"
	tenderminttypes "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/google"
)

const DefGasLimit = 200000

type Client struct {
	chainId string
	context.Context
	TmClient      tmservice.ServiceClient
	TxClient      tx.ServiceClient
	AuthClient    authtypes.QueryClient
	BankClient    banktypes.QueryClient
	StakingClient stakingtypes.QueryClient
	MintClient    minttypes.QueryClient
	OtherClient   otherTypes.QueryClient
}

func NewGrpcConn(rawUrl string) (*grpc.ClientConn, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	_url := u.Host
	if u.Port() == "" {
		if u.Scheme == "http" {
			_url = u.Host + ":80"
		} else {
			_url = u.Host + ":443"
		}
	}
	var opts []grpc.DialOption
	if u.Scheme == "https" {
		opts = append(opts, grpc.WithCredentialsBundle(google.NewDefaultCredentials()))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return grpc.Dial(_url, opts...)
}

func NewGRPCClient(rawUrl string) (*Client, error) {
	grpcConn, err := NewGrpcConn(rawUrl)
	if err != nil {
		return nil, err
	}
	return &Client{
		Context:       context.Background(),
		TmClient:      tmservice.NewServiceClient(grpcConn),
		TxClient:      tx.NewServiceClient(grpcConn),
		AuthClient:    authtypes.NewQueryClient(grpcConn),
		BankClient:    banktypes.NewQueryClient(grpcConn),
		OtherClient:   otherTypes.NewQueryClient(grpcConn),
		StakingClient: stakingtypes.NewQueryClient(grpcConn),
		MintClient:    minttypes.NewQueryClient(grpcConn),
	}, nil
}

func (cli *Client) SetContext(ctx context.Context) {
	cli.Context = ctx
}

func (cli *Client) QueryAccount(fxAddress sdk.AccAddress) (authtypes.AccountI, error) {
	return cli.QueryAccountByFxAddress(fxAddress.String())
}

func (cli *Client) QueryAccountByFxAddress(fxAddress string) (authtypes.AccountI, error) {
	response, err := cli.AuthClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: fxAddress})
	if err != nil {
		return nil, err
	}
	var account authtypes.AccountI
	interfaceRegistry := types.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)
	if err = interfaceRegistry.UnpackAny(response.GetAccount(), &account); err != nil {
		return nil, err
	}
	return account, err
}

func (cli *Client) QueryBalance(fxAddress sdk.AccAddress, denom string) (*sdk.Coin, error) {
	response, err := cli.BankClient.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: fxAddress.String(),
		Denom:   denom,
	})
	if err != nil {
		return nil, err
	}
	return response.Balance, nil
}

func (cli *Client) QueryBalances(fxAddress sdk.AccAddress) (sdk.Coins, error) {
	response, err := cli.BankClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: fxAddress.String(),
	})
	if err != nil {
		return nil, err
	}
	return response.Balances, nil
}

func (cli *Client) GetMintDenom() (string, error) {
	response, err := cli.StakingClient.Params(context.Background(), &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return "", err
	}
	return response.Params.BondDenom, nil
}

func (cli *Client) GetBlockHeight() (int64, error) {
	response, err := cli.TmClient.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return 0, err
	}
	return response.Block.Header.Height, nil
}

func (cli *Client) GetChainId() (string, error) {
	response, err := cli.TmClient.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return "", err
	}
	return response.Block.Header.ChainID, nil
}

func (cli *Client) GetBlockTimeInterval() (time.Duration, error) {
	response1, err := cli.TmClient.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return 0, err
	}
	if response1.Block.Header.Height <= 1 {
		return 0, fmt.Errorf("please try again later, the current block height is less than 1")
	}
	response2, err := cli.TmClient.GetBlockByHeight(context.Background(), &tmservice.GetBlockByHeightRequest{
		Height: response1.Block.Header.Height - 1,
	})
	if err != nil {
		return 0, err
	}
	return response1.Block.Header.Time.Sub(response2.Block.Header.Time), nil
}

func (cli *Client) GetLatestBlock() (*tenderminttypes.Block, error) {
	response, err := cli.TmClient.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return nil, err
	}
	return response.Block, nil
}

func (cli *Client) GetBlockByHeight(blockHeight int64) (*tenderminttypes.Block, error) {
	response, err := cli.TmClient.GetBlockByHeight(context.Background(), &tmservice.GetBlockByHeightRequest{
		Height: blockHeight,
	})
	if err != nil {
		return nil, err
	}
	return response.Block, nil
}

func (cli *Client) GetStatusByTx(txHash string) (*tx.GetTxResponse, error) {
	response, err := cli.TxClient.GetTx(context.Background(), &tx.GetTxRequest{
		Hash: txHash,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (cli *Client) GetGasPrices() (sdk.Coins, error) {
	response, err := cli.OtherClient.GasPrice(context.Background(), &otherTypes.GasPriceRequest{})
	if err != nil {
		return nil, err
	}
	return response.GasPrices, nil
}

func (cli *Client) GetAddressPrefix() (string, error) {
	response, err := cli.TmClient.GetValidatorSetByHeight(context.Background(), &tmservice.GetValidatorSetByHeightRequest{Height: 1})
	if err != nil {
		return "", err
	}
	if len(response.Validators) <= 0 {
		return "", errors.New("no found validator")
	}
	prefix, _, err := bech32.DecodeAndConvert(response.Validators[0].Address)
	if err != nil {
		return "", err
	}
	valConPrefix := sdk.PrefixValidator + sdk.PrefixConsensus
	if strings.HasSuffix(prefix, valConPrefix) {
		return prefix[:len(prefix)-len(valConPrefix)], nil
	}
	return "", errors.New("no found address prefix")
}

func (cli *Client) EstimatingGas(txBody *tx.TxBody, authInfo *tx.AuthInfo, sign []byte) (*sdk.GasInfo, error) {
	response, err := cli.TxClient.Simulate(context.Background(), &tx.SimulateRequest{Tx: &tx.Tx{
		Body:       txBody,
		AuthInfo:   authInfo,
		Signatures: [][]byte{sign},
	}})
	if err != nil {
		return nil, err
	}
	return response.GasInfo, nil
}

func (cli *Client) BuildTx(privKey *secp256k1.PrivKey, msgs []sdk.Msg) (*tx.TxRaw, error) {
	account, err := cli.QueryAccount(privKey.PubKey().Address().Bytes())
	if err != nil {
		return nil, err
	}
	if len(cli.chainId) <= 0 {
		chainId, err := cli.GetChainId()
		if err != nil {
			return nil, err
		}
		cli.chainId = chainId
	}

	prices, err := cli.GetGasPrices()
	if err != nil {
		return nil, err
	}

	txBodyMessage := make([]*types.Any, 0)
	for i := 0; i < len(msgs); i++ {
		msgAnyValue, err := types.NewAnyWithValue(msgs[i])
		if err != nil {
			return nil, err
		}
		txBodyMessage = append(txBodyMessage, msgAnyValue)
	}

	txBody := &tx.TxBody{
		Messages:                    txBodyMessage,
		Memo:                        "",
		TimeoutHeight:               0,
		ExtensionOptions:            nil,
		NonCriticalExtensionOptions: nil,
	}
	txBodyBytes, err := proto.Marshal(txBody)
	if err != nil {
		return nil, err
	}

	pubAny, err := types.NewAnyWithValue(privKey.PubKey())
	if err != nil {
		return nil, err
	}

	authInfo := &tx.AuthInfo{
		SignerInfos: []*tx.SignerInfo{
			{
				PublicKey: pubAny,
				ModeInfo: &tx.ModeInfo{
					Sum: &tx.ModeInfo_Single_{
						Single: &tx.ModeInfo_Single{Mode: signing.SignMode_SIGN_MODE_DIRECT},
					},
				},
				Sequence: account.GetSequence(),
			},
		},
		Fee: &tx.Fee{
			Amount:   nil,
			GasLimit: DefGasLimit,
			Payer:    "",
			Granter:  "",
		},
	}
	for _, price := range prices {
		authInfo.Fee.Amount = sdk.NewCoins(sdk.NewCoin(price.Denom, price.Amount.MulRaw(int64(authInfo.Fee.GasLimit))))
		continue
	}

	txAuthInfoBytes, err := proto.Marshal(authInfo)
	if err != nil {
		return nil, err
	}
	signDoc := &tx.SignDoc{
		BodyBytes:     txBodyBytes,
		AuthInfoBytes: txAuthInfoBytes,
		ChainId:       cli.chainId,
		AccountNumber: account.GetAccountNumber(),
	}
	signatures, err := proto.Marshal(signDoc)
	if err != nil {
		return nil, err
	}
	sign, err := privKey.Sign(signatures)
	if err != nil {
		return nil, err
	}
	logger.Debug("sign", hex.EncodeToString(sign))
	gasInfo, err := cli.EstimatingGas(txBody, authInfo, sign)
	if err != nil {
		return nil, err
	}
	logger.Debug("EstimatingGas GasUsed: ", gasInfo.GasUsed, " GasWanted: ", gasInfo.GasWanted)

	authInfo.Fee.GasLimit = gasInfo.GasUsed * 12 / 10
	for _, price := range prices {
		authInfo.Fee.Amount = sdk.NewCoins(sdk.NewCoin(price.Denom, price.Amount.MulRaw(int64(authInfo.Fee.GasLimit))))
		continue
	}
	logger.Debug("Tx fee amount: ", authInfo.Fee.Amount)

	signDoc.AuthInfoBytes, err = proto.Marshal(authInfo)
	if err != nil {
		return nil, err
	}
	signatures, err = proto.Marshal(signDoc)
	if err != nil {
		return nil, err
	}
	sign, err = privKey.Sign(signatures)
	if err != nil {
		return nil, err
	}
	return &tx.TxRaw{
		BodyBytes:     txBodyBytes,
		AuthInfoBytes: signDoc.AuthInfoBytes,
		Signatures:    [][]byte{sign},
	}, nil
}

func (cli *Client) BroadcastTx(txRaw *tx.TxRaw, mode ...tx.BroadcastMode) (*sdk.TxResponse, error) {
	txBytes, err := proto.Marshal(txRaw)
	if err != nil {
		return nil, err
	}
	defaultMode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	if len(mode) > 0 {
		defaultMode = mode[0]
	}

	_, err = proto.Marshal(&tx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    defaultMode,
	})
	if err != nil {
		return nil, err
	}
	//logger.Warnf("proto marshal tx BroadcastTxRequest: ", broadcastTxData)
	broadcastTxResponse, err := cli.TxClient.BroadcastTx(context.Background(), &tx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    defaultMode,
	})
	if err != nil {
		return nil, err
	}
	txResponse := broadcastTxResponse.TxResponse
	if txResponse.Code != 0 {
		return txResponse, fmt.Errorf(txResponse.RawLog)
	}
	return txResponse, nil
}

func BuildTxV1(chainId string, sequence, accountNumber uint64, privKey *secp256k1.PrivKey, gasPrice sdk.Coin, memo string, timeout uint64, msgs []sdk.Msg) (*tx.TxRaw, error) {

	txBodyMessage := make([]*types.Any, 0)
	for i := 0; i < len(msgs); i++ {
		msgAnyValue, err := types.NewAnyWithValue(msgs[i])
		if err != nil {
			return nil, err
		}
		txBodyMessage = append(txBodyMessage, msgAnyValue)
	}

	txBody := &tx.TxBody{
		Messages:                    txBodyMessage,
		Memo:                        memo,
		TimeoutHeight:               timeout,
		ExtensionOptions:            nil,
		NonCriticalExtensionOptions: nil,
	}
	txBodyBytes, err := proto.Marshal(txBody)
	if err != nil {
		return nil, err
	}

	pubAny, err := types.NewAnyWithValue(privKey.PubKey())
	if err != nil {
		return nil, err
	}

	authInfo := &tx.AuthInfo{
		SignerInfos: []*tx.SignerInfo{
			{
				PublicKey: pubAny,
				ModeInfo: &tx.ModeInfo{
					Sum: &tx.ModeInfo_Single_{
						Single: &tx.ModeInfo_Single{Mode: signing.SignMode_SIGN_MODE_DIRECT},
					},
				},
				Sequence: sequence,
			},
		},
		Fee: &tx.Fee{
			Amount:   sdk.NewCoins(sdk.NewCoin(gasPrice.Denom, gasPrice.Amount.MulRaw(DefGasLimit))),
			GasLimit: DefGasLimit,
			Payer:    "",
			Granter:  "",
		},
	}

	logger.Debugf("tx gas limit: %d, amount: %s", authInfo.Fee.GasLimit, gasPrice.Amount.MulRaw(DefGasLimit).Quo(sdk.NewInt(1000000000000000000)).String())

	txAuthInfoBytes, err := proto.Marshal(authInfo)
	if err != nil {
		return nil, err
	}
	signDoc := &tx.SignDoc{
		BodyBytes:     txBodyBytes,
		AuthInfoBytes: txAuthInfoBytes,
		ChainId:       chainId,
		AccountNumber: accountNumber,
	}
	signatures, err := proto.Marshal(signDoc)
	if err != nil {
		return nil, err
	}
	sign, err := privKey.Sign(signatures)
	if err != nil {
		return nil, err
	}
	return &tx.TxRaw{
		BodyBytes:     txBodyBytes,
		AuthInfoBytes: signDoc.AuthInfoBytes,
		Signatures:    [][]byte{sign},
	}, nil
}

func BuildTxV2(chainId string, sequence, accountNumber uint64, privKey *secp256k1.PrivKey, gasPrice sdk.Coin, msgs []sdk.Msg) (*tx.TxRaw, error) {
	txBodyMessage := make([]*types.Any, 0)
	for i := 0; i < len(msgs); i++ {
		//if err := msgs[i].ValidateBasic(); err != nil {
		//	return nil, err
		//}
		msgAnyValue, err := types.NewAnyWithValue(msgs[i])
		if err != nil {
			return nil, err
		}
		txBodyMessage = append(txBodyMessage, msgAnyValue)
	}

	txBody := &tx.TxBody{
		Messages:                    txBodyMessage,
		Memo:                        "",
		TimeoutHeight:               0,
		ExtensionOptions:            nil,
		NonCriticalExtensionOptions: nil,
	}
	txBodyBytes, err := proto.Marshal(txBody)
	if err != nil {
		return nil, err
	}

	pubAny, err := types.NewAnyWithValue(privKey.PubKey())
	if err != nil {
		return nil, err
	}

	authInfo := &tx.AuthInfo{
		SignerInfos: []*tx.SignerInfo{
			{
				PublicKey: pubAny,
				ModeInfo: &tx.ModeInfo{
					Sum: &tx.ModeInfo_Single_{
						Single: &tx.ModeInfo_Single{Mode: signing.SignMode_SIGN_MODE_DIRECT},
					},
				},
				Sequence: sequence,
			},
		},
		Fee: &tx.Fee{
			Amount:   sdk.NewCoins(sdk.NewCoin(gasPrice.Denom, gasPrice.Amount.MulRaw(DefGasLimit))),
			GasLimit: DefGasLimit,
			Payer:    "",
			Granter:  "",
		},
	}
	logger.Debugf("tx gas limit: %d, gasPrice: %s", authInfo.Fee.GasLimit, gasPrice.Amount.MulRaw(DefGasLimit).Quo(sdk.NewInt(1000000000000000000)).String())

	txAuthInfoBytes, err := proto.Marshal(authInfo)
	if err != nil {
		return nil, err
	}
	signDoc := &tx.SignDoc{
		BodyBytes:     txBodyBytes,
		AuthInfoBytes: txAuthInfoBytes,
		ChainId:       chainId,
		AccountNumber: accountNumber,
	}
	signatures, err := proto.Marshal(signDoc)
	if err != nil {
		return nil, err
	}
	sign, err := privKey.Sign(signatures)
	if err != nil {
		return nil, err
	}
	return &tx.TxRaw{
		BodyBytes:     txBodyBytes,
		AuthInfoBytes: signDoc.AuthInfoBytes,
		Signatures:    [][]byte{sign},
	}, nil
}
