package bridge

import (
	"encoding/hex"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	crosschaintypes "github.com/functionx/fx-core/v3/x/crosschain/types"

	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/contract"
	"github.com/functionx/fx-tron-bridge/internal/logger"
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
	bridger, err := s.CrossChainClient.GetOracleByBridgerAddr(s.GetBridgerAddr().String(), fxtronbridge.Tron)
	if err != nil {
		return err
	}
	if bridger.Online {
		logger.Warnf("get oracle by bridger status is not active bridger: %v", bridger)
		return nil
	}
	if bridger.ExternalAddress != s.GetTronAddr().String() {
		panic("invalid tron private key, expect " + bridger.ExternalAddress)
	}
	logger.Debugf("confirm bridger address: %s", bridger.BridgerAddress)

	if err = s.singerOracleSetConfirm(); err != nil {
		logger.Errorf("singer oracle_set confirm error: %s", err.Error())
	}

	if err = s.singerConfirmBatch(); err != nil {
		logger.Errorf("singer confirm batch error: %s", err.Error())
	}
	return nil
}

func (s *Singer) singerConfirmBatch() error {
	txBatch, err := s.CrossChainClient.LastPendingBatchRequestByAddr(s.GetBridgerAddr().String(), fxtronbridge.Tron)
	if err != nil {
		logger.Errorf("get last pending batch request by addr fail orcAddr: %s, err: %s", s.GetBridgerAddr().String(), err.Error())
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
		&crosschaintypes.MsgConfirmBatch{
			Nonce:           txBatch.BatchNonce,
			TokenContract:   txBatch.TokenContract,
			BridgerAddress:  s.GetBridgerAddr().String(),
			ExternalAddress: s.GetTronAddr().String(),
			Signature:       hex.EncodeToString(sign),
			ChainName:       fxtronbridge.Tron,
		}}, fxtronbridge.BatchSendMsgCount)
}

type IMsg interface {
	sdk.Msg
	GetNonce() uint64
}

func (s *Singer) singerOracleSetConfirm() error {
	oracleSet, err := s.CrossChainClient.LastPendingOracleSetRequestByAddr(s.GetBridgerAddr().String(), fxtronbridge.Tron)
	if err != nil {
		logger.Errorf("get last pending oracle set request by addr fail orcAddr: %s, err: %s", s.GetBridgerAddr().String(), err.Error())
		return err
	}
	if len(oracleSet) <= 0 {
		return nil
	}
	logger.Infof("singer oracle set confirm oracle set len: %d, oracle first nonce: %d, bridger address: %s", len(oracleSet), oracleSet[0].Nonce, s.GetBridgerAddr().String())

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
		iMsgs = append(iMsgs, &crosschaintypes.MsgOracleSetConfirm{
			Nonce:           oracle.Nonce,
			BridgerAddress:  s.GetBridgerAddr().String(),
			ExternalAddress: s.GetTronAddr().String(),
			Signature:       hex.EncodeToString(sign),
			ChainName:       fxtronbridge.Tron,
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
