package main

import (
	fxtronbridge "github.com/functionx/fx-tron-bridge"
	"github.com/functionx/fx-tron-bridge/bridge"
	"github.com/functionx/fx-tron-bridge/utils"
	"github.com/functionx/fx-tron-bridge/utils/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	var prefix = os.Getenv(fxtronbridge.FxAddressPrefixEnv)
	if len(prefix) > 0 {
		utils.UpdateAddressPrefix(prefix)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "fxtronbridge",
		Short: "FunctionX Chain fx tron bridge",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			logger.Init(viper.GetString(fxtronbridge.LogLevelFlag))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			bridgeAddr := viper.GetString("bridge-addr")
			tronGrpc := viper.GetString("tron-grpc")
			fxGrpc := viper.GetString("fx-grpc")
			orcPrivKey, err := utils.DecryptFxPrivateKey(viper.GetString("fx-key"), viper.GetString("fx-pwd"))
			if err != nil {
				return err
			}
			tronPrivateKey, err := utils.DecryptEthPrivateKey(viper.GetString("tron-key"), viper.GetString("tron-pwd"))
			if err != nil {
				return err
			}
			fxTronBridge, err := bridge.NewFxTronBridge(bridgeAddr, tronGrpc, fxGrpc, orcPrivKey, tronPrivateKey)
			if err != nil {
				return err
			}

			fxTronBridge.WaitNewBlock()
			fxtronbridge.StartBridgePrometheus()
			return bridge.Run(fxTronBridge, viper.GetUint64("start-block-number"), viper.GetString("fees"))
		},
	}

	utils.AddFlags(rootCmd, "start-block-number", "", "tron start block number", true)
	utils.AddFlags(rootCmd, "fx-key", "", "fx key", true)
	utils.AddFlags(rootCmd, "fx-pwd", "", "fx pwd", false)
	utils.AddFlags(rootCmd, "tron-key", "", "tron key", true)
	utils.AddFlags(rootCmd, "tron-pwd", "", "tron pwd", false)
	utils.AddFlags(rootCmd, "fees", "FX", "fees", false)
	utils.AddFlags(rootCmd, "bridge-addr", "", "tron contract bridge-token address", true)
	utils.AddFlags(rootCmd, "tron-grpc", "", "tron chain node", true)
	utils.AddFlags(rootCmd, "fx-grpc", "", "fx chain node grpc", true)

	rootCmd.AddCommand(fxtronbridge.NewVersionCmd())
	rootCmd.PersistentFlags().String(fxtronbridge.LogLevelFlag, "info", "the logging level (debug|info|warn|error|dpanic|panic|fatal)")
	utils.SilenceCmdErrors(rootCmd)
	utils.CheckErr(rootCmd.Execute())
}
