package bridge

import (
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	gravitytypes "github.com/functionx/fx-core/x/crosschain/types"
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/client"
	"github.com/functionx/fx-tron-bridge/utils/logger"
	"sort"
	"strconv"
)

type Singer struct {
	*FxTronBridge
	GravityId string
	fees      string
}

func NewSinger(fxBridge *FxTronBridge, fees string) (*Singer, error) {
	params, err := fxBridge.CrossChainClient.Params(fxtronbridge.Tron)
	if err != nil {
		logger.Error("Fx core get gravityId fail err: ", err)
		return nil, err
	}
	return &Singer{
		FxTronBridge: fxBridge,
		GravityId:    params.GravityId,
		fees:         fees,
	}, nil
}

func (s Singer) signer() error {
	orchestrator, err := s.CrossChainClient.GetOracleByOrchestrator(s.OrcAddr, fxtronbridge.Tron)
	if err != nil {
		return err
	}
	if orchestrator.Jailed {
		logger.Warn("Get orchestrator oracle status is not active orchestrator:", orchestrator)
		return nil
	}
	if orchestrator.ExternalAddress != s.TronAddr {
		panic("invalid tron private key, expect " + orchestrator.ExternalAddress)
	}

	logger.Debug("Signer orchestratorAddress: ", orchestrator.OrchestratorAddress)

	if err = s.singerOracleSetConfirm(); err != nil {
		return err
	}

	if err = s.singerConfirmBatch(); err != nil {
		return err
	}

	return s.setFxKeyBalanceMetrics()
}

func (s Singer) setFxKeyBalanceMetrics() error {
	balance, err := s.CrossChainClient.QueryBalance(sdk.AccAddress(s.OrcPrivKey.PubKey().Address()), s.fees)
	if err != nil {
		logger.Warnf("Set fx key balance metrics query balance fail fees: %v, err: %v", s.fees, err)
		return err
	}

	balanceFloat, err := strconv.ParseFloat(balance.Amount.Quo(sdk.NewInt(1e18)).String(), 64)
	if err != nil {
		return err
	}

	fxtronbridge.FxKeyBalanceProm.Set(balanceFloat)
	return nil
}

func (s *Singer) singerConfirmBatch() error {
	txBatch, err := s.CrossChainClient.LastPendingBatchRequestByAddr(s.OrcAddr, fxtronbridge.Tron)
	if err != nil {
		logger.Warnf("Get fx core last pending batch request by addr fail orcAddr: %v, err: %v", s.OrcAddr, err)
		return err
	}
	if txBatch == nil {
		return nil
	}
	logger.Infof("Singer confirm batch tokenContract: %v, batchNonce: %v, block: %v", txBatch.TokenContract, txBatch.BatchNonce, txBatch.Block)

	confirmBatchHash, err := client.EncodeConfirmBatchHash(s.GravityId, *txBatch)
	if err != nil {
		logger.Warnf("Singer confirm batch encodeConfirmBatchHash fail txBatch: %v, err: %v", txBatch, err)
		return err
	}
	sign, err := crypto.Sign(confirmBatchHash, s.TronPrivKey)
	if err != nil {
		logger.Warn("Singer confirm batch sign fail err:", err)
		return err
	}

	txRaw, err := s.CrossChainClient.BuildTx(s.OrcPrivKey, []sdk.Msg{
		&gravitytypes.MsgConfirmBatch{
			Nonce:               txBatch.BatchNonce,
			TokenContract:       txBatch.TokenContract,
			OrchestratorAddress: s.OrcAddr,
			ExternalAddress:     s.TronAddr,
			Signature:           hex.EncodeToString(sign),
			ChainName:           fxtronbridge.Tron,
		}})
	if err != nil {
		logger.Warn("Singer confirm batch build tx fail err:", err)
		return err
	}
	txResp, err := s.CrossChainClient.BroadcastTx(txRaw)
	if err != nil {
		logger.Warn("Singer confirm batch broadcast tx fail err:", err)
		return err
	}
	if txResp.Code != 0 {
		logger.Warnf("Singer confirm batch send msg fail batchNonce: %v, fxcoreHeight: %v, fxcoreHash: %v, code: %v, rawLog: %v", txBatch.BatchNonce, txResp.Height, txResp.TxHash, txResp.Code, txResp.RawLog)
		return fxtronbridge.ErrSendTx
	}

	fxtronbridge.FxSubmitBatchSignProm.Inc()
	logger.Infof("Singer confirm batch send msg success batchNonce: %v, fxcoreHeight: %v, fxcoreHash: %v", txBatch.BatchNonce, txResp.Height, txResp.TxHash)
	return nil
}

func (s *Singer) singerOracleSetConfirm() error {
	oracleSet, err := s.CrossChainClient.LastPendingOracleSetRequestByAddr(s.OrcAddr, fxtronbridge.Tron)
	if err != nil {
		logger.Warnf("Get fx core last pending oracle set request by addr fail orcAddr: %v, err: %v", s.OrcAddr, err)
		return err
	}
	if len(oracleSet) <= 0 {
		return nil
	}
	logger.Infof("Singer oracle set confirm oracle set len: %v, oracle first nonce: %v, orcAddr: %v", len(oracleSet), oracleSet[0].Nonce, s.OrcAddr)

	msgs := make([]*gravitytypes.MsgOracleSetConfirm, 0)
	for _, oracle := range oracleSet {

		hash, err := client.EncodeOracleSetConfirmHash(s.GravityId, *oracle)
		if err != nil {
			logger.Warnf("Singer oracle set confirm encodeOracleSetConfirmHash fail oracle: %v, err: %v", oracle, err)
			return err
		}
		sign, err := crypto.Sign(hash, s.TronPrivKey)
		if err != nil {
			logger.Warn("Singer oracle set confirm sign fail err:", err)
			return err
		}

		msgs = append(msgs, &gravitytypes.MsgOracleSetConfirm{
			Nonce:               oracle.Nonce,
			OrchestratorAddress: s.OrcAddr,
			ExternalAddress:     s.TronAddr,
			Signature:           hex.EncodeToString(sign),
			ChainName:           fxtronbridge.Tron,
		})
	}

	return s.batchSendMsgOracleSetConfirm(msgs, fxtronbridge.BatchSendMsgCount)
}

func (s *Singer) batchSendMsgOracleSetConfirm(msgs []*gravitytypes.MsgOracleSetConfirm, batchNumber int) error {
	if len(msgs) <= 0 {
		return nil
	}

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Nonce < msgs[j].Nonce
	})

	batchCount := len(msgs) / batchNumber
	var endIndex int
	for i := 0; i < batchCount; i++ {
		startIndex := i * batchNumber
		endIndex = startIndex + batchNumber
		if err := s.sendMsgOracleSetConfirm(msgs[startIndex:endIndex]); err != nil {
			return err
		}
	}

	return s.sendMsgOracleSetConfirm(msgs[endIndex:])
}

func (s *Singer) sendMsgOracleSetConfirm(msgs []*gravitytypes.MsgOracleSetConfirm) error {
	sdkMsgs := make([]sdk.Msg, len(msgs))
	for i, msg := range msgs {
		sdkMsgs[i] = msg
	}

	fxtronbridge.FxUpdateOracleSetProm.Add(float64(len(msgs)))

	txRaw, err := s.CrossChainClient.BuildTx(s.OrcPrivKey, sdkMsgs)
	if err != nil {
		logger.Warnf("Singer oracle set confirm build tx fail msgsLen: %v, err: %v", len(sdkMsgs), err)
		return err
	}

	txResp, err := s.CrossChainClient.BroadcastTx(txRaw)
	if err != nil {
		logger.Warnf("Singer oracle set confirm build broadcast tx fail msgsLen: %v, err: %v", len(sdkMsgs), err)
		return err
	}
	if txResp.Code != 0 {
		logger.Warnf("Singer oracle set confirm send msg fail fxcoreHeight: %v, fxcoreHash: %v, code: %v, nonce: %v", txResp.Height, txResp.TxHash, txResp.Code, toNonceArray(msgs))
		return fxtronbridge.ErrSendTx
	}

	logger.Infof("Singer oracle set confirm send msg fxcoreHeight: %v, fxcoreHash: %v, nonce: %v", txResp.Height, txResp.TxHash, toNonceArray(msgs))
	return nil
}

func toNonceArray(msgs []*gravitytypes.MsgOracleSetConfirm) []string {
	nonces := make([]string, len(msgs))
	for i, msg := range msgs {
		nonces[i] = strconv.FormatUint(msg.Nonce, 10)
	}
	return nonces
}
