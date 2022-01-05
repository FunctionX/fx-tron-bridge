package bridge

import (
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	gravitytypes "github.com/functionx/fx-core/x/crosschain/types"
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/contract"
	"github.com/functionx/fx-tron-bridge/internal/logger"
	"sort"
)

type Singer struct {
	*FxTronBridge
	gravityId string
	fees      string
}

func NewSinger(fxBridge *FxTronBridge, fees string) (*Singer, error) {
	params, err := fxBridge.CrossChainClient.Params(fxtronbridge.Tron)
	if err != nil {
		logger.Errorf("get gravityId fail err: %s", err.Error())
		return nil, err
	}
	return &Singer{
		FxTronBridge: fxBridge,
		gravityId:    params.GravityId,
		fees:         fees,
	}, nil
}

func (s Singer) confirm() error {
	orchestrator, err := s.CrossChainClient.GetOracleByOrchestrator(s.GetOrcAddr().String(), fxtronbridge.Tron)
	if err != nil {
		return err
	}
	if orchestrator.Jailed {
		logger.Warnf("get orchestrator oracle status is not active orchestrator: %v", orchestrator)
		return nil
	}
	if orchestrator.ExternalAddress != s.GetTronAddr().String() {
		panic("invalid tron private key, expect " + orchestrator.ExternalAddress)
	}
	logger.Debugf("confirm orchestrator address: %s", orchestrator.OrchestratorAddress)

	if err = s.singerOracleSetConfirm(); err != nil {
		logger.Errorf("singer oracle_set confirm error: %s", err.Error())
	}

	if err = s.singerConfirmBatch(); err != nil {
		logger.Errorf("singer confirm batch error: %s", err.Error())
	}
	return nil
}

func (s *Singer) singerConfirmBatch() error {
	txBatch, err := s.CrossChainClient.LastPendingBatchRequestByAddr(s.GetOrcAddr().String(), fxtronbridge.Tron)
	if err != nil {
		logger.Errorf("get last pending batch request by addr fail orcAddr: %s, err: %s", s.GetOrcAddr().String(), err.Error())
		return err
	}
	if txBatch == nil {
		return nil
	}
	logger.Infof("singer confirm batch tokenContract: %s, batchNonce: %d, blockHeight: %d", txBatch.TokenContract, txBatch.BatchNonce, txBatch.Block)

	confirmBatchHash, err := contract.EncodeConfirmBatchHash(s.gravityId, *txBatch)
	if err != nil {
		logger.Errorf("singer confirm batch encodeConfirmBatchHash fail txBatch: %s, err: %s", txBatch.String(), err.Error())
		return err
	}
	sign, err := crypto.Sign(confirmBatchHash, s.TronPrivKey)
	if err != nil {
		logger.Errorf("singer confirm batch sign fail err: %s", err.Error())
		return err
	}
	return s.BatchSendMsg([]sdk.Msg{
		&gravitytypes.MsgConfirmBatch{
			Nonce:               txBatch.BatchNonce,
			TokenContract:       txBatch.TokenContract,
			OrchestratorAddress: s.GetOrcAddr().String(),
			ExternalAddress:     s.GetTronAddr().String(),
			Signature:           hex.EncodeToString(sign),
			ChainName:           fxtronbridge.Tron,
		}}, fxtronbridge.BatchSendMsgCount)
}

type IMsg interface {
	sdk.Msg
	GetNonce() uint64
}

func (s *Singer) singerOracleSetConfirm() error {
	oracleSet, err := s.CrossChainClient.LastPendingOracleSetRequestByAddr(s.GetOrcAddr().String(), fxtronbridge.Tron)
	if err != nil {
		logger.Errorf("get last pending oracle set request by addr fail orcAddr: %s, err: %s", s.GetOrcAddr().String(), err.Error())
		return err
	}
	if len(oracleSet) <= 0 {
		return nil
	}
	logger.Infof("singer oracle set confirm oracle set len: %d, oracle first nonce: %d, orcAddr: %s", len(oracleSet), oracleSet[0].Nonce, s.GetOrcAddr().String())

	iMsgs := make([]IMsg, 0)
	for _, oracle := range oracleSet {
		hash, err := contract.EncodeOracleSetConfirmHash(s.gravityId, *oracle)
		if err != nil {
			logger.Errorf("singer oracle set confirm encodeOracleSetConfirmHash fail oracle: %s, err: %s", oracle, err.Error())
			return err
		}
		sign, err := crypto.Sign(hash, s.TronPrivKey)
		if err != nil {
			logger.Errorf("singer oracle set confirm sign fail err: %s", err.Error())
			return err
		}
		iMsgs = append(iMsgs, &gravitytypes.MsgOracleSetConfirm{
			Nonce:               oracle.Nonce,
			OrchestratorAddress: s.GetOrcAddr().String(),
			ExternalAddress:     s.GetTronAddr().String(),
			Signature:           hex.EncodeToString(sign),
			ChainName:           fxtronbridge.Tron,
		})
	}
	sort.Slice(iMsgs, func(i, j int) bool {
		return iMsgs[i].GetNonce() < iMsgs[j].GetNonce()
	})
	msgs := make([]sdk.Msg, 0)
	for _, imsg := range iMsgs {
		msgs = append(msgs, imsg.(sdk.Msg))
	}
	return s.BatchSendMsg(msgs, fxtronbridge.BatchSendMsgCount)
}
