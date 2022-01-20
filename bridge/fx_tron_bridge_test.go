package bridge

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	sdkContract "github.com/fbsobreira/gotron-sdk/pkg/proto/core/contract"
	"github.com/functionx/fx-core/app"
	"github.com/functionx/fx-tron-bridge/client"
	"github.com/functionx/fx-tron-bridge/logger"
	"strconv"
	"testing"
)

const testOrcMnemonic = ""
const testTronPrivKey = ""

func TestConnect(t *testing.T) {
	fxBridge, err := GetTestFxBridge()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(fxBridge)
}

func TestWaitNewBlock(t *testing.T) {
	logger.Init("debug")
	fxBridge, err := GetTestFxBridge()
	if err != nil {
		t.Error(err)
	}
	fxBridge.WaitNewBlock()
	t.Log("end")
}

func GetTestFxBridge() (*FxTronBridge, error) {
	orcPrivKey := app.NewPrivKeyFromMnemonic(testOrcMnemonic)
	tronPrivateKey, err := crypto.HexToECDSA(testTronPrivKey)
	if err != nil {
		return nil, err
	}
	fxBridge, err := NewFxTronBridge("TDrfM9c4P2rSkg41C3bCPn8wW1MMR93gXq", "http://127.0.0.1:50051", "http://127.0.0.1:9090", orcPrivKey, tronPrivateKey)
	if err != nil {
		return nil, err
	}
	return fxBridge, nil
}

func TestNewFxBridge2(t *testing.T) {
	contractAddress := "TDzvk3dzcEThbAAJE9e4y4LsHZX8iXQq4N"
	fmt.Printf("contract: %s\n", contractAddress)
	tokenAddress := "TDzvk3dzcEThbAAJE9e4y4LsHZX8iXQq4N"
	fmt.Printf("token: %s\n", tokenAddress)
	contractDesc, err := address.Base58ToAddress(contractAddress)
	if err != nil {
		t.Fatal(err)
	}
	params := []abi.Param{
		{"address": tokenAddress},
	}
	data, err := abi.Pack("checkAssetStatus(address)", params)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("data: %s\n", hex.EncodeToString(data))
	tronClient, err := client.NewTronGrpcClient("http://127.0.0.1:50051")
	if err != nil {
		t.Fatal(err)
	}
	tx := &sdkContract.TriggerSmartContract{
		ContractAddress: contractDesc.Bytes(),
		Data:            data,
	}
	res, err := tronClient.Client.TriggerConstantContract(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(res.ConstantResult), res.ConstantResult)
	t.Log(res.String())
	fmt.Printf("ConstantResult: %s\n", hex.EncodeToString(res.ConstantResult[0]))

	atoi, err := strconv.Atoi(hex.EncodeToString(res.ConstantResult[0]))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(atoi)
}
