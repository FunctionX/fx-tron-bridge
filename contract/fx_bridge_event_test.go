package contract

import (
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	troncommon "github.com/fbsobreira/gotron-sdk/pkg/common"
)

func TestAddressToString(t *testing.T) {
	ethAddress := ethCommon.HexToAddress("0x8a21bcef7269bd328bf843207bfe0d84dc3b68e9")
	tronAddress := AddressToString(ethAddress)
	t.Log("tronAddress:", tronAddress)
}

func TestFixedBytes(t *testing.T) {
	bytes := fixedBytes("fx/transfer")
	t.Log("bytes:", bytes)
}

func TestTronAddress(t *testing.T) {
	contract, err := troncommon.DecodeCheck("TTjQHsuUru5HRWDmePBgjHNFUJ2nhFTffd")
	if err != nil {
		return
	}
	contractStr := troncommon.BytesToHexString(contract)
	t.Log(contractStr)

	token, err := troncommon.DecodeCheck("TXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj")
	if err != nil {
		return
	}
	tokenStr := troncommon.BytesToHexString(token)
	t.Log(tokenStr)
}
