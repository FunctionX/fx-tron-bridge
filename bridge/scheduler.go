package bridge

import (
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/utils/logger"
	"time"
)

func Run(fxBridge *FxTronBridge, startBlockNumber uint64, fees string) error {
	if startBlockNumber > 0 {
		startBlockNumber--
	}
	oracle, err := NewOracle(fxBridge, startBlockNumber)
	if err != nil {
		return err
	}
	singer, err := NewSinger(fxBridge, fees)
	if err != nil {
		return err
	}

	eventHandlerTicker := time.NewTicker(fxtronbridge.FxAvgBlockMillisecond)
	for range eventHandlerTicker.C {
		if err = oracle.HandleEvent(); err != nil {
			if err == fxtronbridge.ErrSendTx {
				continue
			}
			logger.Error("Bridge oracle error:", err)
		}

		if err = singer.signer(); err != nil {
			if err == fxtronbridge.ErrSendTx {
				continue
			}
			logger.Error("Bridge signer error:", err)
		}
	}

	return nil
}
