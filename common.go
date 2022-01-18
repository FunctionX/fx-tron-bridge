package fxtronbridge

import (
	"errors"
	"time"
)

const FxAddressPrefixEnv = "FX_ADDRESS_PREFIX"
const LogLevelFlag = "log-level"
const Tron = "tron"

var ErrSendTx = errors.New("tron bridge send tx failed")

const FxAvgBlockMillisecond = 6 * time.Second

const (
	TronBlockDelay        = 25
	TronDelayBlockWarn    = 3000
	TronRestartDelayBlock = 28800
	TronHome              = "$HOME/.tronBridge"
)

const (
	BatchSendMsgCount            = 100
	totalPower                   = 2834678415 // 66% of uint32_max
	thresholdVotePowerProportion = 66         // 66%
	ThresholdVotePower           = totalPower * thresholdVotePowerProportion / 100
)
