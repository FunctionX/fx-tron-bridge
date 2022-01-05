package client

import (
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	sdkCommon "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"testing"
)

func TestStateLastOracleSetNonce(t *testing.T) {
	tronClient := NewTestTronClient(t)
	nonce, err := tronClient.StateLastOracleSetNonce("TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh")
	if err != nil {
		t.Fatal(err)
	}
	println("nonce:", nonce)
}

func TestStateFxBridgeId(t *testing.T) {
	tronClient := NewTestTronClient(t)
	fxBridgeId, err := tronClient.StateFxBridgeId("TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh")
	if err != nil {
		t.Fatal(err)
	}
	println("fxBridgeId:", fxBridgeId)
}

func TestStateLastOracleSetHeight(t *testing.T) {
	tronClient := NewTestTronClient(t)
	lastOracleSetHeight, err := tronClient.StateLastOracleSetHeight("TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh")
	if err != nil {
		t.Fatal(err)
	}
	println("lastOracleSetHeight:", lastOracleSetHeight)
}

func TestGetTokenStatus(t *testing.T) {
	tronClient := NewTestTronClient(t)
	a, b, c, err := tronClient.GetTokenStatus("TAXdkUAde6ztmJyQqANL4jwDYab8D9shQs", "TLBaRhANQoJFTqre9Nf1mjuwNWjCJeYqUL")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("tokenStatus:", a, b, c)
}

func TestGetBridgeTokenList(t *testing.T) {
	tronClient := NewTestTronClient(t)
	bridgeTokenList, err := tronClient.GetBridgeTokenList("TAXdkUAde6ztmJyQqANL4jwDYab8D9shQs")
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range bridgeTokenList {
		t.Log(token)
	}
}

func NewTestTronClient(t *testing.T) *TronClient {
	client, err := NewTronGrpcClient("http://127.0.0.1:50051") // testnet
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func Test_Fee(t *testing.T) {
	tronClient := NewTestTronClient(t)
	contractAddress := "TKEdTmLSocskqBUFL2fBHSUixhZbo9RG55"

	var i int64 = 20941200
	for {
		blockInfo, err := tronClient.GetBlockInfoByNum(i)
		if err != nil {
			t.Fatal(err)
		}

		for _, transactionInfo := range blockInfo.TransactionInfo {
			if core.Transaction_Result_SUCCESS != transactionInfo.Receipt.Result {
				continue
			}
			for _, sdkLog := range transactionInfo.Log {
				t.Logf("ContractAddress: %v, Address: %v", sdkCommon.EncodeCheck(transactionInfo.ContractAddress), sdkCommon.EncodeCheck(append([]byte{address.TronBytePrefix}, sdkLog.Address...)))
				if contractAddress != sdkCommon.EncodeCheck(transactionInfo.ContractAddress) && contractAddress != sdkCommon.EncodeCheck(append([]byte{address.TronBytePrefix}, sdkLog.Address...)) {
					continue
				}

				if len(sdkLog.Topics) <= 0 || ethCommon.BytesToHash(sdkLog.Topics[0]).Hex() != fxBridgeAbi.Events["TransactionBatchExecutedEvent"].ID.String() {
					continue
				}

				info, err := tronClient.WithMint(transactionInfo.Id)
				if err != nil {
					t.Fatal(err)
				}
				receipt := info.GetReceipt()
				t.Logf("EnergyUsage: %v, EnergyUsageTotal: %v, NetUsage: %v", receipt.EnergyUsage, receipt.EnergyUsageTotal, receipt.NetUsage)
			}
		}
		i--
	}
}

func Test_GetBlockInfoByNum(t *testing.T) {
	tronClient := NewTestTronClient(t)

	blockInfo, err := tronClient.GetBlockInfoByNum(20874558)
	if err != nil {
		t.Fatal(err)
	}

	for i, transactionInfo := range blockInfo.TransactionInfo {
		fmt.Printf("\nindex: %v, result: %v, transactionHash: %v, blockNumber: %v, contractAddress: %v \n", i, transactionInfo.Receipt.Result.String(),
			hex.EncodeToString(transactionInfo.Id), transactionInfo.BlockNumber, sdkCommon.EncodeCheck(transactionInfo.ContractAddress))
		for _, log := range transactionInfo.Log {
			if hex.EncodeToString(log.Topics[0]) != "034c5b22dd525a50d0a6b15549df0a6ac83b833a6c3da57ea16890832c72507c" {
				continue
			}
			for _, topic := range log.Topics {
				t.Log("==>", hex.EncodeToString(topic))
			}
			t.Log("==>", sdkCommon.EncodeCheck(append([]byte{address.TronBytePrefix}, log.Address...)))
			//fmt.Println("log data:{}", hex.EncodeToString(log.Data))
		}
	}
}

func Test_QueryBlockEvent(t *testing.T) {
	tronClient := NewTestTronClient(t)
	var i uint64 = 20874960
	for {
		sendToFxEvents, transactionBatchExecutedEvents, addBridgeTokenEvents, oracleSetUpdatedEvents, err := tronClient.QueryBlockEvent("TKEdTmLSocskqBUFL2fBHSUixhZbo9RG55", i)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("blockNumber:", i)
		for _, sendToFx := range sendToFxEvents {
			t.Logf("\nTokenContract: %v, Sender: %v, Amount: %v, EventNonce: %v, Destination: %v, TargetIBC: %v \n", AddressToString(sendToFx.TokenContract), AddressToString(sendToFx.Sender), sendToFx.Amount.String(),
				sendToFx.EventNonce.String(), sdk.AccAddress(sendToFx.Destination[12:]).String(), HexByte32ToTargetIbc(sendToFx.TargetIBC))
		}

		for _, transactionBatchExecuted := range transactionBatchExecutedEvents {
			t.Logf("\nToken: %v, BatchNonce: %v, EventNonce: %v \n", AddressToString(transactionBatchExecuted.Token), transactionBatchExecuted.BatchNonce.String(), transactionBatchExecuted.EventNonce.String())
		}

		for _, addBridgeToken := range addBridgeTokenEvents {
			t.Logf("\nTokenContract: %v, Name: %v, Symbol: %v, Decimals: %v, EventNonce: %v \n", AddressToString(addBridgeToken.TokenContract), addBridgeToken.Name, addBridgeToken.Symbol,
				addBridgeToken.Decimals, addBridgeToken.EventNonce.String())
		}

		for _, oracleSetUpdated := range oracleSetUpdatedEvents {
			validatorAddress := ""
			powers := ""
			for i, validator := range oracleSetUpdated.Oracles {
				validatorAddress += AddressToString(validator)
				powers += oracleSetUpdated.Powers[i].String()
				if i > 0 {
					validatorAddress += ","
					powers += ","
				}
			}

			t.Logf("\nvalidatorAddress: %v, NewValsetNonce: %v, powers: %v, EventNonce: %v \n", validatorAddress, oracleSetUpdated.NewOracleSetNonce.String(), powers, oracleSetUpdated.EventNonce)
		}
		i++
	}
}
