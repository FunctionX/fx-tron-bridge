package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	sdkContract "github.com/fbsobreira/gotron-sdk/pkg/proto/core/contract"
	"github.com/functionx/fx-tron-bridge/contract"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"math/big"
	"testing"
	"time"
)

const grpcUrl = "http://3.225.171.164:50051"
const testTronPrivKey = ""

func TestNewTronClient(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}
	decimals, err := cli.TRC20GetDecimals("TBswVtM9kcgwq35RGLC3xEH4ec8LxavfsX")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decimals)
}

func TestGetLastBlockNumber(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}
	lastBlockNumber, err := cli.BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("lastBlockNumber:", lastBlockNumber)
}

func TestWithMint(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}
	mint, err := cli.WithMint([]byte("acc35d1cfc53f21f5a134e517dc9681362242b9afb56ad193b62ec982f8b4c27"), time.Second*30)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(mint)
}

func TestAllowance(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}
	allowance, err := cli.Allowance("TDzLXdsScB8abJdkV3vuaHGqRMYUMyap5y", "TL5w1JJNU68G821xKNck92qMyK6UJ9dTRs", "TRoP1bXv5L19QWN6jDdDfUQ97ediXUYQNw")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("allowance:", allowance)
}

func TestTransaction(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := cli.GetTransactionByID("acc35d1cfc53f21f5a134e517dc9681362242b9afb56ad193b62ec982f8b4c27")
	if err != nil {
		t.Fatal()
	}
	txData, _ := json.Marshal(tx)
	t.Log("tx", string(txData))
	tx.Ret = nil
	t.Log("Bandwidth Point:", proto.Size(tx)+64)
}

func TestGasPrice(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice: %v", gasPrice.Uint64())
}

func TestGetNowBlock(t *testing.T) {
	grpcClient := client.NewGrpcClient("http://127.0.0.1:50051")
	err := grpcClient.Start(grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	block, err := grpcClient.Client.GetNowBlock(context.Background(), &api.EmptyMessage{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(block.BlockHeader.RawData.Number)
}

func TestGetChainParams(t *testing.T) {
	grpcClient := client.NewGrpcClient("http://127.0.0.1:50051")
	err := grpcClient.Start(grpc.WithInsecure())
	require.NoError(t, err)

	parameters, err := grpcClient.Client.GetChainParameters(context.Background(), &api.EmptyMessage{})
	require.NoError(t, err)

	for i, parameter := range parameters.GetChainParameter() {
		t.Logf("index:[%v], param key:[%v], value:[%v]", i, parameter.GetKey(), parameter.GetValue())
	}
}

func TestGetLimit(t *testing.T) {
	feeLimit := GetLimit(big.NewInt(280), 111542)
	assert.Equal(t, 37478112, feeLimit)
}

func TestEstimateGas(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}

	_, destination, err := bech32.DecodeAndConvert("fx1zgpzdf2uqla7hkx85wnn4p2r3duwqzd8xst6v2")
	if err != nil {
		t.Log(err)
	}
	var destinationByte32 [32]byte
	copy(destinationByte32[12:], destination)
	var targetIBCByte32 [32]byte
	copy(targetIBCByte32[:], "transfer/channel-0")

	params := []abi.Param{
		{"address": "TBswVtM9kcgwq35RGLC3xEH4ec8LxavfsX"},
		{"bytes32": destinationByte32},
		{"bytes32": targetIBCByte32},
		{"uint256": big.NewInt(int64(80000001))},
	}

	data, err := abi.Pack("sendToFx(address,bytes32,bytes32,uint256)", params)
	if err != nil {
		t.Fatal(err)
	}

	contractDesc, err := address.Base58ToAddress("TXqLgpc9FnixpZMrAA6M2PYgZaPfqRUdU7")
	if err != nil {
		t.Fatal(err)
	}
	tronPrivateKey, err := crypto.HexToECDSA(testTronPrivKey)
	if err != nil {
		t.Fatal(err)
	}
	tronAddress := address.PubkeyToAddress(tronPrivateKey.PublicKey)
	t.Log("tronAddress:", tronAddress.String())

	energy, err := cli.EstimateGas(tronAddress.Bytes(), contractDesc.Bytes(), data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("energy:", energy)
}

func TestEnergy(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	contractDesc, err := address.Base58ToAddress("TXqLgpc9FnixpZMrAA6M2PYgZaPfqRUdU7")
	if err != nil {
		t.Fatal(err)
	}
	toAddress := common.BytesToAddress(contractDesc.Bytes())
	from, err := address.Base58ToAddress("TL5w1JJNU68G821xKNck92qMyK6UJ9dTRs")
	if err != nil {
		t.Fatal(err)
	}
	fromAddress := common.BytesToAddress(from.Bytes())
	_, destination, err := bech32.DecodeAndConvert("fx1zgpzdf2uqla7hkx85wnn4p2r3duwqzd8xst6v2")
	if err != nil {
		t.Fatal(err)
	}

	var destinationByte32 [32]byte
	copy(destinationByte32[12:], destination)
	var targetIBCByte32 [32]byte
	copy(targetIBCByte32[:], "transfer/channel-0")

	params := []abi.Param{
		{"address": "TBswVtM9kcgwq35RGLC3xEH4ec8LxavfsX"},
		{"bytes32": destinationByte32},
		{"bytes32": targetIBCByte32},
		{"uint256": big.NewInt(int64(80000001))},
	}

	data, err := abi.Pack("sendToFx(address,bytes32,bytes32,uint256)", params)
	if err != nil {
		t.Fatal(err)
	}

	msg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &toAddress,
		GasPrice: gasPrice,
		Data:     data,
	}
	estimateGas, err := cli.EstimateGas(msg.From.Bytes(), msg.To.Bytes(), msg.Data)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("energy: %v", estimateGas)
}

func TestSendToFx(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_, destination, err := bech32.DecodeAndConvert("fx1zgpzdf2uqla7hkx85wnn4p2r3duwqzd8xst6v2")
	if err != nil {
		t.Fatal(err)
	}
	var destinationByte32 [32]byte
	copy(destinationByte32[12:], destination)
	var targetIBCByte32 [32]byte
	copy(targetIBCByte32[:], "transfer/channel-0")

	params := []abi.Param{
		{"address": "TBswVtM9kcgwq35RGLC3xEH4ec8LxavfsX"},
		{"bytes32": destinationByte32},
		{"bytes32": targetIBCByte32},
		{"uint256": big.NewInt(int64(80000001))},
	}

	data, err := abi.Pack("sendToFx(address,bytes32,bytes32,uint256)", params)
	if err != nil {
		t.Fatal(err)
	}

	contractDesc, err := address.Base58ToAddress("TXqLgpc9FnixpZMrAA6M2PYgZaPfqRUdU7")
	if err != nil {
		t.Fatal(err)
	}
	tronPrivateKey, err := crypto.HexToECDSA(testTronPrivKey)
	if err != nil {
		t.Fatal(err)
	}
	tronAddress := address.PubkeyToAddress(tronPrivateKey.PublicKey)
	t.Log("tronAddress:", tronAddress.String())

	energy, err := cli.EstimateGas(tronAddress.Bytes(), contractDesc.Bytes(), data)
	if err != nil {
		t.Fatal(err)
	}
	feeLimit := GetLimit(gasPrice, energy)
	t.Log("feeLimit:", feeLimit)

	tx, err := cli.TriggerContract(&sdkContract.TriggerSmartContract{
		OwnerAddress:    tronAddress.Bytes(),
		ContractAddress: contractDesc.Bytes(),
		Data:            data,
	}, feeLimit)
	if err != nil {
		t.Fatal(err)
	}

	signTx, err := contract.SignTx(tronPrivateKey, tx)
	if err != nil {
		t.Fatal(err)
	}
	apiReturn, err := cli.BroadcastTx(signTx)
	if err != nil {
		t.Fatal(err)
	}

	info, err := cli.WithMint(tx.Txid, time.Second*30)
	if err != nil {
		t.Fatal(err)
	}
	receipt := info.Receipt

	t.Logf("EnergyUsage: %v, EnergyUsageTotal: %v, NetUsage: %v, RawData: %v, Signature:%v, data:%v", receipt.EnergyUsage, receipt.EnergyUsageTotal, receipt.NetUsage, len(signTx.Transaction.RawData.Data), len(signTx.Transaction.Signature), len(data))
	t.Logf("apiReturn: %v, txHash: %v, code: %v", apiReturn, hex.EncodeToString(tx.Txid), info.Receipt.Result)
}
