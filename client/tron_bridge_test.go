package client

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	sdkCommon "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	crosschaintypes "github.com/functionx/fx-core/v2/x/crosschain/types"

	"github.com/functionx/fx-tron-bridge/contract"
)

func TestStateLastOracleSetNonce(t *testing.T) {
	tronClient := NewTestTronClient(t)
	nonce, err := tronClient.StateLastOracleSetNonce("TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("nonce:", nonce)
}

func TestStateFxBridgeId(t *testing.T) {
	tronClient := NewTestTronClient(t)
	fxBridgeId, err := tronClient.StateFxBridgeId("TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("fxBridgeId:", fxBridgeId)
}

func TestStateLastOracleSetHeight(t *testing.T) {
	tronClient := NewTestTronClient(t)
	lastOracleSetHeight, err := tronClient.StateLastOracleSetHeight("TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("lastOracleSetHeight:", lastOracleSetHeight)
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

				info, err := tronClient.WithMint(transactionInfo.Id, time.Second*30)
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
			//t.Log("log data:{}", hex.EncodeToString(log.Data))
		}
	}
}

func TestEncodeOracleSetConfirmHash(t *testing.T) {
	var oracle = crosschaintypes.OracleSet{
		Nonce: 10,
		Members: []crosschaintypes.BridgeValidator{
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
	hash, err := contract.EncodeOracleSetConfirmHash("tron", oracle)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("hash:", hex.EncodeToString(hash))
}

func TestEncodeConfirmBatchHash(t *testing.T) {
	txBatch := crosschaintypes.OutgoingTxBatch{
		BatchNonce:   4,
		BatchTimeout: 1000,
		Transactions: []*crosschaintypes.OutgoingTransferTx{
			{
				Id:          1,
				Sender:      "fx1zgpzdf2uqla7hkx85wnn4p2r3duwqzd8xst6v2",
				DestAddress: "TFysCB929XGezbnyumoFScyevjDggu3BPq",
				Token: crosschaintypes.ERC20Token{
					Contract: "TFETBkg3wrgEEDPXEPZCem4tegsaTw2fwR",
					Amount:   sdk.NewInt(2000000000),
				},
				Fee: crosschaintypes.ERC20Token{
					Contract: "TFETBkg3wrgEEDPXEPZCem4tegsaTw2fwR",
					Amount:   sdk.NewInt(10000000),
				},
			},
		},
		TokenContract: "TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh",
		Block:         9000,
		FeeReceive:    "TWF75HQEiMJpKZbX1CE6iwXfwU7ZZm7T7f",
	}

	hash, err := contract.EncodeConfirmBatchHash("tron", txBatch)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("hash:", hex.EncodeToString(hash))
}

func TestPrivateKeyToAddress(t *testing.T) {
	tronPrivateKey, err := crypto.HexToECDSA(testTronPrivKey)
	if err != nil {
		t.Fatal(err)
	}
	tronAddress := address.PubkeyToAddress(tronPrivateKey.PublicKey)
	t.Log("tronAddress:", tronAddress.String())
}
