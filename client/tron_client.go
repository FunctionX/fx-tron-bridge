package client

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"math/big"
	"net/url"
	"time"

	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core/contract"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/google"
)

type TronClient struct {
	*client.GrpcClient
}

func NewTronGrpcClient(grpcUrl string) (*TronClient, error) {
	parseUrl, err := url.Parse(grpcUrl)
	if err != nil {
		return nil, err
	}
	host := parseUrl.Host
	cli := client.NewGrpcClient(host)

	var opts []grpc.DialOption
	if parseUrl.Scheme == "https" {
		opts = append(opts, grpc.WithCredentialsBundle(google.NewDefaultCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err := cli.Start(opts...); err != nil {
		return nil, err
	}
	return &TronClient{GrpcClient: cli}, nil
}

func (c *TronClient) BlockNumber(_ context.Context) (uint64, error) {
	block, err := c.GetNowBlock()
	if err != nil {
		return 0, err
	}
	return uint64(block.GetBlockHeader().RawData.Number), nil
}

func (c *TronClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	parameters, err := c.Client.GetChainParameters(ctx, &api.EmptyMessage{})
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
	tx := &contract.TriggerSmartContract{
		OwnerAddress:    from,
		ContractAddress: to,
		Data:            data,
	}
	transactionExtention, err := c.Client.TriggerConstantContract(context.Background(), tx)
	if err != nil {
		return 0, err
	}
	if transactionExtention.Transaction == nil {
		return 0, fmt.Errorf("transaction is nil")
	}
	rets := transactionExtention.Transaction.Ret
	for _, ret := range rets {

		//
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

func (c *TronClient) TriggerContract(ct *contract.TriggerSmartContract, feeLimit int64) (*api.TransactionExtention, error) {
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
		if err := c.UpdateHash(tx); err != nil {
			return nil, err
		}
	}
	return tx, err
}

func (c *TronClient) BroadcastTx(tx *api.TransactionExtention) (*api.Return, error) {
	result, err := c.Broadcast(tx.Transaction)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("bad transaction: %v", string(result.GetMessage()))
	}
	return result, nil
}

func (c *TronClient) WithMint(txId []byte, timeOut time.Duration) (info *core.TransactionInfo, err error) {
	transactionId := new(api.BytesMessage)
	transactionId.Value = txId
	timeout, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	for {
		info, err = c.Client.GetTransactionInfoById(timeout, transactionId)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(info.Id, txId) {
			return info, nil
		}
		time.Sleep(3 * time.Second)
	}
}

func GetLimit(gasPrice *big.Int, energy uint64) int64 {
	limit := int64(energy * gasPrice.Uint64())
	return limit*2/10 + limit
}
