package client

import (
	"context"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	sdkContract "github.com/fbsobreira/gotron-sdk/pkg/proto/core/contract"
	gravitytypes "github.com/functionx/fx-core/x/crosschain/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"math/big"
	"testing"
)

const grpcUrl = "http://127.0.0.1:50051"
const rpcUrl = "http://127.0.0.1:50545/jsonrpc"
const testTronPrivKey =  ""

func TestNewTronClient(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}
	//t.Log("cli:", cli)
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
	lastBlockNumber, err := cli.GetLastBlockNumber()
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
	mint, err := cli.WithMint([]byte("sdfdf"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(mint)
}

func TestSendToFx(t *testing.T) {
	cli, err := NewTronGrpcJsonRpcClient(grpcUrl, rpcUrl)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := cli.SuggestGasPrice()
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
	feeLimit := cli.GetFeeLimit(gasPrice, energy)
	t.Log("feeLimit:", feeLimit)

	tx, err := cli.TriggerContract(&sdkContract.TriggerSmartContract{
		OwnerAddress:    tronAddress.Bytes(),
		ContractAddress: contractDesc.Bytes(),
		Data:            data,
	}, feeLimit)
	if err != nil {
		t.Fatal(err)
	}

	signTx, err := SignTx(tronPrivateKey, tx)
	if err != nil {
		t.Fatal(err)
	}
	apiReturn, err := cli.BroadcastTransaction(signTx)
	if err != nil {
		t.Fatal(err)
	}

	info, err := cli.WithMint(tx.Txid)
	if err != nil {
		t.Fatal(err)
	}
	receipt := info.Receipt

	t.Logf("EnergyUsage: %v, EnergyUsageTotal: %v, NetUsage: %v, RawData: %v, Signature:%v, data:%v", receipt.EnergyUsage, receipt.EnergyUsageTotal, receipt.NetUsage, len(signTx.Transaction.RawData.Data), len(signTx.Transaction.Signature), len(data))
	t.Logf("apiReturn: %v, txHash: %v, code: %v", apiReturn, hex.EncodeToString(tx.Txid), info.Receipt.Result)
}

func TestSize(t *testing.T) {
	cli, err := NewTronGrpcClient(grpcUrl)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := cli.GetTransactionByID("9eaa02d1d5a2dde9ff5f6141f3fd4b94f14167189250c6a9b0a8a549c30bcb7c")
	if err != nil {
		t.Fatal()
	}
	tx.Ret = nil
	t.Log("Bandwidth Point:", proto.Size(tx)+64)
}

func TestGasPrice(t *testing.T) {
	cli, err := NewTronJsonRpcClient(rpcUrl)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := cli.SuggestGasPrice()
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

func TestEstimateGas(t *testing.T) {
	cli, err := NewTronGrpcJsonRpcClient(grpcUrl, rpcUrl)
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

func TestGetFeeLimit(t *testing.T) {
	cli := TronClient{}
	feeLimit := cli.GetFeeLimit(big.NewInt(280), 700)
	t.Log("feeLimit:", feeLimit)
}

func TestEnergy(t *testing.T) {
	cli, err := NewTronJsonRpcClient(rpcUrl)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := cli.SuggestGasPrice()
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
	data, err := getTestData()
	if err != nil {
		t.Fatal(err)
	}

	msg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &toAddress,
		GasPrice: gasPrice,
		Data:     data,
	}
	estimateGas, err := cli.JsonRpcClient.EstimateGas(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("energy: %v", estimateGas)
}

func getTestData() ([]byte, error) {
	_, destination, err := bech32.DecodeAndConvert("fx1zgpzdf2uqla7hkx85wnn4p2r3duwqzd8xst6v2")
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return data, nil
}

func TestPrivateKeyToAddress(t *testing.T) {
	tronPrivateKey, err := crypto.HexToECDSA(testTronPrivKey)
	if err != nil {
		t.Fatal(err)
	}
	tronAddress := address.PubkeyToAddress(tronPrivateKey.PublicKey)
	t.Log("tronAddress:", tronAddress.String())
}

func TestAddressToString(t *testing.T) {
	ethAddress := common.HexToAddress("0x8a21bcef7269bd328bf843207bfe0d84dc3b68e9")
	tronAddress := AddressToString(ethAddress)
	t.Log("tronAddress:", tronAddress)
}

func TestFixedBytes(t *testing.T) {
	bytes := FixedBytes("fx/transfer")
	t.Log("bytes:", bytes)
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

func TestEncodeOracleSetConfirmHash(t *testing.T) {
	var oracle = gravitytypes.OracleSet{
		Nonce: 10,
		Members: []*gravitytypes.BridgeValidator{
			{
				Power:           1000,
				ExternalAddress: "TFysCB929XGezbnyumoFScyevjDggu3BPq",
			}, {
				Power:           3000,
				ExternalAddress: "TYp74R387xLBoCGp2imSwkzWkfcMyr7FeP",
			},
		},
		Height: 100000,
	}
	hash, err := EncodeOracleSetConfirmHash("tron", oracle)
	if err != nil {
		t.Fatal(err)
	}
	println("hash:", hex.EncodeToString(hash))
}

func TestEncodeConfirmBatchHash(t *testing.T) {
	txBatch := gravitytypes.OutgoingTxBatch{
		BatchNonce:   4,
		BatchTimeout: 1000,
		Transactions: []*gravitytypes.OutgoingTransferTx{
			{
				Id:          1,
				Sender:      "fx1zgpzdf2uqla7hkx85wnn4p2r3duwqzd8xst6v2",
				DestAddress: "TFysCB929XGezbnyumoFScyevjDggu3BPq",
				Token: &gravitytypes.ExternalToken{
					Contract: "TFETBkg3wrgEEDPXEPZCem4tegsaTw2fwR",
					Amount:   sdk.NewInt(2000000000),
				},
				Fee: &gravitytypes.ExternalToken{
					Contract: "TFETBkg3wrgEEDPXEPZCem4tegsaTw2fwR",
					Amount:   sdk.NewInt(10000000),
				},
			},
		},
		TokenContract: "TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh",
		Block:         9000,
		FeeReceive:    "TWF75HQEiMJpKZbX1CE6iwXfwU7ZZm7T7f",
	}

	hash, err := EncodeConfirmBatchHash("tron", txBatch)
	if err != nil {
		t.Fatal(err)
	}
	println("hash:", hex.EncodeToString(hash))
}
