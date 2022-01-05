package fxtronbridge

import (
	"fmt"
	"github.com/functionx/fx-tron-bridge/internal/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var BlockHeightProm = prometheus.NewGauge(prometheus.GaugeOpts{Subsystem: "eth_bridge_oracle", Name: "sync_block_height"})
var BlockIntervalProm = prometheus.NewGauge(prometheus.GaugeOpts{Subsystem: "eth_bridge_oracle", Name: "query_log_block_interval"})
var MsgPendingLenProm = prometheus.NewCounter(prometheus.CounterOpts{Subsystem: "eth_bridge_oracle", Name: "msg_pending_count"})

var FxKeyBalanceProm = prometheus.NewGauge(prometheus.GaugeOpts{Subsystem: "", Name: "fx_key_balance"})
var FxUpdateOracleSetProm = prometheus.NewCounter(prometheus.CounterOpts{Subsystem: "", Name: "update_oracle_set_sign"})
var FxSubmitBatchSignProm = prometheus.NewCounter(prometheus.CounterOpts{Subsystem: "", Name: "submit_batch_sign"})

func StartBridgePrometheus() {
	prometheus.DefaultRegisterer.MustRegister(BlockHeightProm)
	prometheus.DefaultRegisterer.MustRegister(BlockIntervalProm)

	prometheus.DefaultRegisterer.MustRegister(MsgPendingLenProm)

	prometheus.DefaultRegisterer.MustRegister(FxKeyBalanceProm)
	prometheus.DefaultRegisterer.MustRegister(FxUpdateOracleSetProm)
	prometheus.DefaultRegisterer.MustRegister(FxSubmitBatchSignProm)
	go func() {
		srv := &http.Server{
			Addr: ":9811",
			Handler: promhttp.InstrumentMetricHandler(
				prometheus.DefaultRegisterer, promhttp.HandlerFor(
					prometheus.DefaultGatherer,
					promhttp.HandlerOpts{MaxRequestsInFlight: 3},
				),
			),
		}
		logger.Infof("=====> start prometheus server: http://127.0.0.1%s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Sprintf("=====> start prometheus server failed: %v", err))
		}
	}()
}
