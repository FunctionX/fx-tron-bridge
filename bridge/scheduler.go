package bridge

import (
	"time"

	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/internal/logger"
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
		if err = oracle.bridgeEvent(); err != nil {
			logger.Errorf("bridge oracle error: %s", err)
		}

		if err = singer.confirm(); err != nil {
			logger.Errorf("bridge confirm error: %s", err)
		}

		fxBridge.setFxKeyBalanceMetrics(fees)
	}
	return nil
}
