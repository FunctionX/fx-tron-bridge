package bridge

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/functionx/fx-core/app"
	"github.com/functionx/fx-core/x/crosschain/types"
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"testing"
)

func TestRun(t *testing.T) {
	testClient, err := GetTestClient()
	if err != nil {
		t.Error(err)
	}
	err = Run(testClient, 21138373-1, "FX")
	if err != nil {
		t.Error(err)
	}
}

func GetTestClient() (*FxTronBridge, error) {
	orcPrivKey := app.NewPrivKeyFromMnemonic(testOrcMnemonic)
	tronPrivateKey, err := crypto.HexToECDSA(testTronPrivKey)
	if err != nil {
		return nil, err
	}
	fxBridge, err := NewFxTronBridge("TGKeDpMMbgqDfT7CNsWZ5Xzvh4Ch9KyETr", "http://127.0.0.1:50051", "http://127.0.0.1:9090", orcPrivKey, tronPrivateKey)
	if err != nil {
		return nil, err
	}
	return fxBridge, nil
}

func TestMsgsToJson(t *testing.T) {
	msgs := make([]Msg, 0)
	msgs = append(msgs, Msg{
		Height:     100,
		EventNonce: 23,
		Msg: &types.MsgSendToFxClaim{
			ChainName: fxtronbridge.Tron,
		},
	})
	msgs = append(msgs, Msg{
		Height:     200,
		EventNonce: 43,
		Msg: &types.MsgSendToFxClaim{
			ChainName: "test",
		},
	})
	t.Log(MsgsToJson(msgs))
}
