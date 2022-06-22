package contract

import (
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	troncommon "github.com/fbsobreira/gotron-sdk/pkg/common"
	crosschaintypes "github.com/functionx/fx-core/x/crosschain/types"
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"math/big"
)

const TronSignaturePrefix = "\x19TRON Signed Message:\n32"

type IEvent interface {
	ToMsg(blockHeight uint64, orchestrator string) sdk.Msg
	GetEventNonce() uint64
}

func (event *FxBridgeTronTransactionBatchExecutedEvent) ToMsg(blockHeight uint64, bridgerAddress string) sdk.Msg {
	return &crosschaintypes.MsgSendToExternalClaim{
		EventNonce:     event.EventNonce.Uint64(),
		BlockHeight:    blockHeight,
		BridgerAddress: bridgerAddress,
		BatchNonce:     event.BatchNonce.Uint64(),
		TokenContract:  AddressToString(event.Token),
		ChainName:      fxtronbridge.Tron,
	}
}

func (event *FxBridgeTronTransactionBatchExecutedEvent) GetEventNonce() uint64 {
	return event.EventNonce.Uint64()
}

func (event *FxBridgeTronOracleSetUpdatedEvent) ToMsg(blockHeight uint64, bridgerAddress string) sdk.Msg {
	members := make([]crosschaintypes.BridgeValidator, len(event.Oracles))
	for i, oracleAddress := range event.Oracles {
		members[i] = crosschaintypes.BridgeValidator{
			Power:           event.Powers[i].Uint64(),
			ExternalAddress: AddressToString(oracleAddress),
		}
	}
	return &crosschaintypes.MsgOracleSetUpdatedClaim{
		EventNonce:     event.EventNonce.Uint64(),
		BlockHeight:    blockHeight,
		BridgerAddress: bridgerAddress,
		OracleSetNonce: event.NewOracleSetNonce.Uint64(),
		Members:        members,
		ChainName:      fxtronbridge.Tron,
	}
}

func (event *FxBridgeTronOracleSetUpdatedEvent) GetEventNonce() uint64 {
	return event.EventNonce.Uint64()
}

func (event *FxBridgeTronAddBridgeTokenEvent) ToMsg(blockHeight uint64, bridgerAddress string) sdk.Msg {
	return &crosschaintypes.MsgBridgeTokenClaim{
		EventNonce:     event.EventNonce.Uint64(),
		TokenContract:  AddressToString(event.TokenContract),
		BlockHeight:    blockHeight,
		BridgerAddress: bridgerAddress,
		Name:           event.Name,
		Symbol:         event.Symbol,
		Decimals:       uint64(event.Decimals),
		ChannelIbc:     hexByte32ToTargetIbc(event.ChannelIBC),
		ChainName:      fxtronbridge.Tron,
	}
}

func (event *FxBridgeTronAddBridgeTokenEvent) GetEventNonce() uint64 {
	return event.EventNonce.Uint64()
}

func (event *FxBridgeTronSendToFxEvent) ToMsg(blockHeight uint64, bridgerAddress string) sdk.Msg {
	return &crosschaintypes.MsgSendToFxClaim{
		EventNonce:     event.EventNonce.Uint64(),
		BlockHeight:    blockHeight,
		BridgerAddress: bridgerAddress,
		TokenContract:  AddressToString(event.TokenContract),
		Amount:         sdk.NewIntFromBigInt(event.Amount),
		Sender:         AddressToString(event.Sender),
		Receiver:       sdk.AccAddress(event.Destination[12:]).String(),
		TargetIbc:      hexByte32ToTargetIbc(event.TargetIBC),
		ChainName:      fxtronbridge.Tron,
	}
}

func (event *FxBridgeTronSendToFxEvent) GetEventNonce() uint64 {
	return event.EventNonce.Uint64()
}

func UnpackLog(abi ethabi.ABI, out interface{}, event string, log ethtypes.Log) error {
	if log.Topics[0] != abi.Events[event].ID {
		return fmt.Errorf("event signature mismatch")
	}
	if len(log.Data) > 0 {
		if err := abi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed ethabi.Arguments
	for _, arg := range abi.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return ethabi.ParseTopics(out, indexed, log.Topics[1:])
}

func EncodeOracleSetConfirmHash(gravityId string, oracle crosschaintypes.OracleSet) ([]byte, error) {
	addresses := make([]string, len(oracle.Members))
	powers := make([]*big.Int, len(oracle.Members))
	for i, member := range oracle.Members {
		addresses[i] = member.ExternalAddress
		powers[i] = big.NewInt(int64(member.Power))
	}
	params := []abi.Param{
		{"bytes32": fixedBytes(gravityId)},
		{"bytes32": fixedBytes("checkpoint")},
		{"uint256": big.NewInt(int64(oracle.Nonce))},
		{"address[]": addresses},
		{"uint256[]": powers},
	}
	encode, err := abi.GetPaddedParam(params)
	if err != nil {
		return nil, err
	}
	encodeBys := crypto.Keccak256Hash(encode).Bytes()
	protectedHash := crypto.Keccak256Hash(append([]uint8(TronSignaturePrefix), encodeBys...))
	return protectedHash.Bytes(), nil
}

func EncodeConfirmBatchHash(gravityId string, txBatch crosschaintypes.OutgoingTxBatch) ([]byte, error) {
	txCount := len(txBatch.Transactions)
	amounts := make([]*big.Int, txCount)
	destinations := make([]string, txCount)
	fees := make([]*big.Int, txCount)
	for i, transferTx := range txBatch.Transactions {
		amounts[i] = transferTx.Token.Amount.BigInt()
		destinations[i] = transferTx.DestAddress
		fees[i] = transferTx.Fee.Amount.BigInt()
	}

	params := []abi.Param{
		{"bytes32": fixedBytes(gravityId)},
		{"bytes32": fixedBytes("transactionBatch")},
		{"uint256[]": amounts},
		{"address[]": destinations},
		{"uint256[]": fees},
		{"uint256": big.NewInt(int64(txBatch.BatchNonce))},
		{"address": txBatch.TokenContract},
		{"uint256": big.NewInt(int64(txBatch.BatchTimeout))},
		{"address": txBatch.FeeReceive},
	}
	encode, err := abi.GetPaddedParam(params)
	if err != nil {
		return nil, err
	}
	encodeBys := crypto.Keccak256Hash(encode).Bytes()
	protectedHash := crypto.Keccak256Hash(append([]uint8(TronSignaturePrefix), encodeBys...))
	return protectedHash.Bytes(), nil
}

func fixedBytes(value string) [32]byte {
	slice := []byte(value)
	var arr [32]byte
	copy(arr[:], slice[:])
	return arr
}

func AddressToString(addr ethcommon.Address) string {
	addressByte := append([]byte{}, address.TronBytePrefix)
	addressByte = append(addressByte, addr.Bytes()...)
	return troncommon.EncodeCheck(addressByte)
}

func hexByte32ToTargetIbc(bytes [32]byte) string {
	for i := len(bytes) - 1; i >= 0; i-- {
		if bytes[i] != 0 {
			return hex.EncodeToString(bytes[:i+1])
		}
	}
	return ""
}
