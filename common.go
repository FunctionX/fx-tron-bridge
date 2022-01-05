package fxtronbridge

import (
	"time"
)

const Tron = "tron"

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
