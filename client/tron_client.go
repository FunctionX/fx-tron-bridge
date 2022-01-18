package client

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	sdkCommon "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	sdkContract "github.com/fbsobreira/gotron-sdk/pkg/proto/core/contract"
	gravitytypes "github.com/functionx/fx-core/x/crosschain/types"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/google"
	"math/big"
	"net/url"
	"time"
)

const signaturePrefix = "\x19TRON Signed Message:\n32"

type TronClient struct {
	*client.GrpcClient
}

func NewTronGrpcClient(grpcUrl string) (*TronClient, error) {
	parseU, err := url.Parse(grpcUrl)
	if err != nil {
		return nil, err
	}
	host := parseU.Host
	cli := client.NewGrpcClient(host)

	var opts []grpc.DialOption
	if parseU.Scheme == "https" {
		opts = append(opts, grpc.WithCredentialsBundle(google.NewDefaultCredentials()))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if err := cli.Start(opts...); err != nil {
		return nil, err
	}
	return &TronClient{GrpcClient: cli}, nil
}

func (c TronClient) SuggestGasPrice() (*big.Int, error) {
	parameters, err := c.Client.GetChainParameters(context.Background(), &api.EmptyMessage{})
	if err != nil {
		return nil, err
	}

	for _, parameter := range parameters.GetChainParameter() {
		if parameter.GetKey() == "getEnergyFee" {
			return big.NewInt(parameter.GetValue()), nil
		}
	}
	return nil, fmt.Errorf("not gasPrice")
}

func (c *TronClient) EstimateGas(from, to, data []byte) (uint64, error) {
	tx := &sdkContract.TriggerSmartContract{
		OwnerAddress:    from,
		ContractAddress: to,
		Data:            data,
	}
	transactionExtention, err := c.Client.TriggerConstantContract(context.Background(), tx)
	if err != nil {
		return 0, err
	}

	rets := transactionExtention.Transaction.Ret
	for _, ret := range rets {
		if ret.Ret != core.Transaction_Result_SUCESS {
			var constantResults []byte
			for _, by := range transactionExtention.ConstantResult[0] {
				if by != 0 {
					constantResults = append(constantResults, by)
				}
			}
			return 0, fmt.Errorf("EstimateGas error txRet: %v, contractRet: %v, txError: %v, contractError: %v", ret.Ret, ret.ContractRet.String(), string(transactionExtention.Result.Message), string(constantResults))
		}
	}
	return uint64(transactionExtention.EnergyUsed), nil
}

func (c TronClient) GetFeeLimit(gasPrice *big.Int, energy uint64) int64 {
	feeLimit := int64(energy * gasPrice.Uint64())
	return feeLimit*2/10 + feeLimit
}

func (c *TronClient) GetLastBlockNumber() (uint64, error) {
	block, err := c.GetNowBlock()
	if err != nil {
		return 0, err
	}
	return uint64(block.GetBlockHeader().RawData.Number), nil
}

func (c *TronClient) TriggerContract(ct *sdkContract.TriggerSmartContract, feeLimit int64) (*api.TransactionExtention, error) {
	tx, err := c.Client.TriggerContract(context.Background(), ct)
	if err != nil {
		return nil, err
	}

	if tx.Result.Code > 0 {
		return nil, fmt.Errorf("%s", string(tx.Result.Message))
	}
	if feeLimit > 0 {
		tx.Transaction.RawData.FeeLimit = feeLimit
		// update hash
		err := c.UpdateHash(tx)
		if err != nil {
			return nil, err
		}
	}
	return tx, err
}

func SignTx(privateKey *ecdsa.PrivateKey, tx *api.TransactionExtention) (*api.TransactionExtention, error) {
	rawData, err := proto.Marshal(tx.Transaction.RawData)
	if err != nil {
		return nil, err
	}
	h256 := sha256.New()
	h256.Write(rawData)
	hash := h256.Sum(nil)

	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, err
	}
	tx.Transaction.Signature = append(tx.Transaction.Signature, signature)
	return tx, nil
}

func (c *TronClient) BroadcastTransaction(tx *api.TransactionExtention) (*api.Return, error) {
	result, err := c.Broadcast(tx.Transaction)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("bad transaction: %v", string(result.GetMessage()))
	}
	return result, nil
}

func (c *TronClient) WithMint(txId []byte) (info *core.TransactionInfo, err error) {
	transactionId := new(api.BytesMessage)
	transactionId.Value = txId
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		info, err = c.Client.GetTransactionInfoById(timeout, transactionId)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(info.Id, txId) {
			return info, nil
		}
		time.Sleep(1 * time.Second)
	}
}

func EncodeOracleSetConfirmHash(gravityId string, oracle gravitytypes.OracleSet) ([]byte, error) {
	addresses := make([]string, len(oracle.Members))
	powers := make([]*big.Int, len(oracle.Members))
	for i, member := range oracle.Members {
		addresses[i] = member.ExternalAddress
		powers[i] = big.NewInt(int64(member.Power))
	}
	params := []abi.Param{
		{"bytes32": FixedBytes(gravityId)},
		{"bytes32": FixedBytes("checkpoint")},
		{"uint256": big.NewInt(int64(oracle.Nonce))},
		{"address[]": addresses},
		{"uint256[]": powers},
	}
	encode, err := abi.GetPaddedParam(params)
	if err != nil {
		return nil, err
	}
	encodeBys := crypto.Keccak256Hash(encode).Bytes()
	protectedHash := crypto.Keccak256Hash(append([]uint8(signaturePrefix), encodeBys...))
	return protectedHash.Bytes(), nil
}

func EncodeConfirmBatchHash(gravityId string, txBatch gravitytypes.OutgoingTxBatch) ([]byte, error) {
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
		{"bytes32": FixedBytes(gravityId)},
		{"bytes32": FixedBytes("transactionBatch")},
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
	protectedHash := crypto.Keccak256Hash(append([]uint8(signaturePrefix), encodeBys...))
	return protectedHash.Bytes(), nil
}

func FixedBytes(value string) [32]byte {
	slice := []byte(value)
	var arr [32]byte
	copy(arr[:], slice[:])
	return arr
}

func AddressToString(addr common.Address) string {
	addressByte := append([]byte{}, address.TronBytePrefix)
	addressByte = append(addressByte, addr.Bytes()...)
	return sdkCommon.EncodeCheck(addressByte)
}

func HexByte32ToTargetIbc(bytes [32]byte) string {
	for i := len(bytes) - 1; i >= 0; i-- {
		if bytes[i] != 0 {
			return hex.EncodeToString(bytes[:i+1])
		}
	}
	return ""
}
