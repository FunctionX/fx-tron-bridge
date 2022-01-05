package fxchain

import (
	"context"
	crosschaintypes "github.com/functionx/fx-core/x/crosschain/types"
)

/* ======================================> Cross Chain gravity grpc <====================================== */

type CrossChainClient struct {
	*Client
	crossChainClient crosschaintypes.QueryClient
}

func NewCrossChainClient(grpcUrl string, urls ...string) (*CrossChainClient, error) {
	grpcConn, err := NewGrpcConn(grpcUrl)
	if err != nil {
		return nil, err
	}
	client, err := NewGRPCClient(grpcUrl)
	if err != nil {
		return nil, err
	}
	cli := &CrossChainClient{
		Client:           client,
		crossChainClient: crosschaintypes.NewQueryClient(grpcConn),
	}
	return cli, nil
}

func (cli *CrossChainClient) CurrentOracleSet(chainName string) (*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.CurrentOracleSet(context.Background(), &crosschaintypes.QueryCurrentOracleSetRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSet, nil
}

func (cli *CrossChainClient) TokenToDenom(token, chainName string) (*crosschaintypes.QueryTokenToDenomResponse, error) {
	response, err := cli.crossChainClient.TokenToDenom(context.Background(), &crosschaintypes.QueryTokenToDenomRequest{Token: token, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (cli *CrossChainClient) BatchFees(chainName string) ([]*crosschaintypes.BatchFees, error) {
	response, err := cli.crossChainClient.BatchFees(context.Background(), &crosschaintypes.QueryBatchFeeRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.BatchFees, nil
}

func (cli *CrossChainClient) BatchConfirms(nonce uint64, tokenContract, chainName string) ([]*crosschaintypes.MsgConfirmBatch, error) {
	response, err := cli.crossChainClient.BatchConfirms(context.Background(), &crosschaintypes.QueryBatchConfirmsRequest{Nonce: nonce, TokenContract: tokenContract, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Confirms, nil
}

func (cli *CrossChainClient) OutgoingTxBatches(chainName string) ([]*crosschaintypes.OutgoingTxBatch, error) {
	response, err := cli.crossChainClient.OutgoingTxBatches(context.Background(), &crosschaintypes.QueryOutgoingTxBatchesRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Batches, nil
}

func (cli *CrossChainClient) OracleSetConfirmsByNonce(nonce uint64, chainName string) ([]*crosschaintypes.MsgOracleSetConfirm, error) {
	response, err := cli.crossChainClient.OracleSetConfirmsByNonce(context.Background(), &crosschaintypes.QueryOracleSetConfirmsByNonceRequest{Nonce: nonce, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Confirms, nil
}

func (cli *CrossChainClient) LastOracleSetRequests(chainName string) ([]*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.LastOracleSetRequests(context.Background(), &crosschaintypes.QueryLastOracleSetRequestsRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSets, nil
}

func (cli *CrossChainClient) OracleSetRequest(nonce uint64, chainName string) (*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.OracleSetRequest(context.Background(), &crosschaintypes.QueryOracleSetRequestRequest{Nonce: nonce, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSet, nil
}

func (cli *CrossChainClient) LastPendingBatchRequestByAddr(orchestratorAddress string, chainName string) (*crosschaintypes.OutgoingTxBatch, error) {
	response, err := cli.crossChainClient.LastPendingBatchRequestByAddr(context.Background(), &crosschaintypes.QueryLastPendingBatchRequestByAddrRequest{OrchestratorAddress: orchestratorAddress, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Batch, nil
}

func (cli *CrossChainClient) Params(chainName string) (*crosschaintypes.Params, error) {
	response, err := cli.crossChainClient.Params(context.Background(), &crosschaintypes.QueryParamsRequest{ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return &response.Params, nil
}

func (cli *CrossChainClient) LastPendingOracleSetRequestByAddr(orchestratorAddress string, chainName string) ([]*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.LastPendingOracleSetRequestByAddr(context.Background(), &crosschaintypes.QueryLastPendingOracleSetRequestByAddrRequest{OrchestratorAddress: orchestratorAddress, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.OracleSets, nil
}

func (cli *CrossChainClient) GetOracleByOrchestrator(orchestratorAddress string, chainName string) (*crosschaintypes.Oracle, error) {
	response, err := cli.crossChainClient.GetOracleByOrchestrator(context.Background(), &crosschaintypes.QueryOracleByOrchestratorRequest{OrchestratorAddress: orchestratorAddress, ChainName: chainName})
	if err != nil {
		return nil, err
	}
	return response.Oracle, nil
}

func (cli *CrossChainClient) LastEventNonceByAddr(orchestratorAddress string, chainName string) (uint64, error) {
	response, err := cli.crossChainClient.LastEventNonceByAddr(context.Background(), &crosschaintypes.QueryLastEventNonceByAddrRequest{OrchestratorAddress: orchestratorAddress, ChainName: chainName})
	if err != nil {
		return 0, err
	}
	return response.EventNonce, nil
}

func (cli *CrossChainClient) LastEventBlockHeightByAddr(orchestratorAddress string, chainName string) (uint64, error) {
	response, err := cli.crossChainClient.LastEventBlockHeightByAddr(context.Background(), &crosschaintypes.QueryLastEventBlockHeightByAddrRequest{OrchestratorAddress: orchestratorAddress, ChainName: chainName})
	if err != nil {
		return 0, err
	}
	return response.BlockHeight, nil
}

func (cli *CrossChainClient) GetGravityId(chainName string) (string, error) {
	response, err := cli.crossChainClient.Params(context.Background(), &crosschaintypes.QueryParamsRequest{
		ChainName: chainName,
	})
	if err != nil {
		return "", err
	}
	return response.Params.GravityId, nil
}

func (cli *CrossChainClient) GetCurrentOracleSet(chainName string) (*crosschaintypes.OracleSet, error) {
	response, err := cli.crossChainClient.CurrentOracleSet(context.Background(), &crosschaintypes.QueryCurrentOracleSetRequest{
		ChainName: chainName,
	})
	if err != nil {
		return nil, err
	}
	return response.OracleSet, nil
}

func (cli *CrossChainClient) GetOracleSetRequest(chainName string, nonce uint64) (*crosschaintypes.OracleSet, error) {
	request, err := cli.crossChainClient.OracleSetRequest(context.Background(), &crosschaintypes.QueryOracleSetRequestRequest{
		ChainName: chainName,
		Nonce:     nonce,
	})
	if err != nil {
		return nil, err
	}
	return request.OracleSet, nil
}

func (cli *CrossChainClient) GetLastOracleSetRequest(chainName string) ([]*crosschaintypes.OracleSet, error) {
	request, err := cli.crossChainClient.LastOracleSetRequests(context.Background(), &crosschaintypes.QueryLastOracleSetRequestsRequest{
		ChainName: chainName,
	})
	if err != nil {
		return nil, err
	}
	return request.OracleSets, nil
}

func (cli *CrossChainClient) GetOracleSetConfirmsByNonce(chainName string, nonce uint64) ([]*crosschaintypes.MsgOracleSetConfirm, error) {
	request, err := cli.crossChainClient.OracleSetConfirmsByNonce(context.Background(), &crosschaintypes.QueryOracleSetConfirmsByNonceRequest{
		ChainName: chainName,
		Nonce:     nonce,
	})
	if err != nil {
		return nil, err
	}
	return request.Confirms, nil
}
