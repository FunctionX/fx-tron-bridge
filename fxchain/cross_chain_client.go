package fxchain

import (
	"context"

	"github.com/functionx/fx-core/v2/client/grpc"
	crosschaintypes "github.com/functionx/fx-core/v2/x/crosschain/types"
)

/* ======================================> Cross Chain gravity grpc <====================================== */

type CrossChainClient struct {
	*grpc.Client
	ctx              context.Context
	crossChainClient crosschaintypes.QueryClient
}

func NewCrossChainClient(ctx context.Context, grpcUrl string) (*CrossChainClient, error) {
	client, err := grpc.NewClient(grpcUrl)
	if err != nil {
		return nil, err
	}
	client.WithContext(ctx)
	cli := &CrossChainClient{
		ctx:              ctx,
		Client:           client,
		crossChainClient: crosschaintypes.NewQueryClient(client),
	}
	return cli, nil
}

func (cli *CrossChainClient) CurrentOracleSet(chainName string) (*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.CurrentOracleSet(cli.ctx, &crosschaintypes.QueryCurrentOracleSetRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSet, nil
}

func (cli *CrossChainClient) TokenToDenom(token, chainName string) (*crosschaintypes.QueryTokenToDenomResponse, error) {
	response, err := cli.crossChainClient.TokenToDenom(cli.ctx, &crosschaintypes.QueryTokenToDenomRequest{Token: token, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (cli *CrossChainClient) BatchFees(chainName string) ([]*crosschaintypes.BatchFees, error) {
	response, err := cli.crossChainClient.BatchFees(cli.ctx, &crosschaintypes.QueryBatchFeeRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.BatchFees, nil
}

func (cli *CrossChainClient) BatchConfirms(nonce uint64, tokenContract, chainName string) ([]*crosschaintypes.MsgConfirmBatch, error) {
	response, err := cli.crossChainClient.BatchConfirms(cli.ctx, &crosschaintypes.QueryBatchConfirmsRequest{Nonce: nonce, TokenContract: tokenContract, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Confirms, nil
}

func (cli *CrossChainClient) OutgoingTxBatches(chainName string) ([]*crosschaintypes.OutgoingTxBatch, error) {
	response, err := cli.crossChainClient.OutgoingTxBatches(cli.ctx, &crosschaintypes.QueryOutgoingTxBatchesRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Batches, nil
}

func (cli *CrossChainClient) OracleSetConfirmsByNonce(nonce uint64, chainName string) ([]*crosschaintypes.MsgOracleSetConfirm, error) {
	response, err := cli.crossChainClient.OracleSetConfirmsByNonce(cli.ctx, &crosschaintypes.QueryOracleSetConfirmsByNonceRequest{Nonce: nonce, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Confirms, nil
}

func (cli *CrossChainClient) LastOracleSetRequests(chainName string) ([]*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.LastOracleSetRequests(cli.ctx, &crosschaintypes.QueryLastOracleSetRequestsRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSets, nil
}

func (cli *CrossChainClient) OracleSetRequest(nonce uint64, chainName string) (*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.OracleSetRequest(cli.ctx, &crosschaintypes.QueryOracleSetRequestRequest{Nonce: nonce, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSet, nil
}

func (cli *CrossChainClient) LastPendingBatchRequestByAddr(bridgerAddress string, chainName string) (*crosschaintypes.OutgoingTxBatch, error) {
	response, err := cli.crossChainClient.LastPendingBatchRequestByAddr(cli.ctx, &crosschaintypes.QueryLastPendingBatchRequestByAddrRequest{BridgerAddress: bridgerAddress, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Batch, nil
}

func (cli *CrossChainClient) Params(chainName string) (*crosschaintypes.Params, error) {
	response, err := cli.crossChainClient.Params(cli.ctx, &crosschaintypes.QueryParamsRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return &response.Params, nil
}

func (cli *CrossChainClient) LastPendingOracleSetRequestByAddr(bridgerAddress string, chainName string) ([]*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.LastPendingOracleSetRequestByAddr(cli.ctx, &crosschaintypes.QueryLastPendingOracleSetRequestByAddrRequest{BridgerAddress: bridgerAddress, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSets, nil
}

func (cli *CrossChainClient) GetOracleByBridgerAddr(bridgerAddress string, chainName string) (*crosschaintypes.Oracle, error) {
	response, err := cli.crossChainClient.GetOracleByBridgerAddr(cli.ctx, &crosschaintypes.QueryOracleByBridgerAddrRequest{BridgerAddress: bridgerAddress, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Oracle, nil
}

func (cli *CrossChainClient) LastEventNonceByAddr(bridgerAddress string, chainName string) (uint64, error) {
	response, err := cli.crossChainClient.LastEventNonceByAddr(cli.ctx, &crosschaintypes.QueryLastEventNonceByAddrRequest{BridgerAddress: bridgerAddress, ChainName: chainName})
	if err != nil {
		return 0, err
	}
	return response.EventNonce, nil
}

func (cli *CrossChainClient) LastEventBlockHeightByAddr(bridgerAddress string, chainName string) (uint64, error) {
	response, err := cli.crossChainClient.LastEventBlockHeightByAddr(cli.ctx, &crosschaintypes.QueryLastEventBlockHeightByAddrRequest{BridgerAddress: bridgerAddress, ChainName: chainName})
	if err != nil {
		return 0, err
	}
	return response.BlockHeight, nil
}

func (cli *CrossChainClient) GetGravityId(chainName string) (string, error) {
	response, err := cli.crossChainClient.Params(cli.ctx, &crosschaintypes.QueryParamsRequest{
		ChainName: chainName,
	})
	if err != nil {
		return "", err
	}
	return response.Params.GravityId, nil
}

func (cli *CrossChainClient) GetCurrentOracleSet(chainName string) (*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.CurrentOracleSet(cli.ctx, &crosschaintypes.QueryCurrentOracleSetRequest{
		ChainName: chainName,
	})
	if err != nil {
		return nil, err
	}
	return response.OracleSet, nil
}

func (cli *CrossChainClient) GetOracleSetRequest(chainName string, nonce uint64) (*crosschaintypes.OracleSet, error) {
	request, err := cli.crossChainClient.OracleSetRequest(cli.ctx, &crosschaintypes.QueryOracleSetRequestRequest{
		ChainName: chainName,
		Nonce:     nonce,
	})
	if err != nil {
		return nil, err
	}
	return request.OracleSet, nil
}

func (cli *CrossChainClient) GetLastOracleSetRequest(chainName string) ([]*crosschaintypes.OracleSet, error) {
	request, err := cli.crossChainClient.LastOracleSetRequests(cli.ctx, &crosschaintypes.QueryLastOracleSetRequestsRequest{
		ChainName: chainName,
	})
	if err != nil {
		return nil, err
	}
	return request.OracleSets, nil
}

func (cli *CrossChainClient) GetOracleSetConfirmsByNonce(chainName string, nonce uint64) ([]*crosschaintypes.MsgOracleSetConfirm, error) {
	request, err := cli.crossChainClient.OracleSetConfirmsByNonce(cli.ctx, &crosschaintypes.QueryOracleSetConfirmsByNonceRequest{
		ChainName: chainName,
		Nonce:     nonce,
	})
	if err != nil {
		return nil, err
	}
	return request.Confirms, nil
}
