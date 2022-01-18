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

const bridgeAddrFlag = "bridge-addr"
const tronGrpcFlag = "tron-grpc"
const fxGrpcFlag = "fx-grpc"

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
	}

	rootCmd.AddCommand(newBridgeCommand())
	for _, command := range rootCmd.Commands() {
		command.Flags().String(bridgeAddrFlag, "", "tron contract bridge-token address")
		command.Flags().String(tronGrpcFlag, "", "tron chain node")
		command.Flags().String(fxGrpcFlag, "", "fx chain node grpc")

		utils.CheckErr(command.MarkFlagRequired(bridgeAddrFlag))
		utils.CheckErr(command.MarkFlagRequired(tronGrpcFlag))
		utils.CheckErr(command.MarkFlagRequired(fxGrpcFlag))
	}

	rootCmd.AddCommand(fxtronbridge.NewVersionCmd())

	rootCmd.PersistentFlags().String(fxtronbridge.LogLevelFlag, "info", "the logging level (debug|info|warn|error|dpanic|panic|fatal)")
	utils.SilenceCmdErrors(rootCmd)
	utils.CheckErr(rootCmd.Execute())
}

func newBridgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "bridge",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			bridgeAddr := viper.GetString(bridgeAddrFlag)
			tronGrpc := viper.GetString(tronGrpcFlag)
			fxGrpc := viper.GetString(fxGrpcFlag)
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
	utils.AddFlags(cmd, "start-block-number", "", "tron start block number", true)
	utils.AddFlags(cmd, "fx-key", "", "fx key", true)
	utils.AddFlags(cmd, "fx-pwd", "", "fx pwd", false)
	utils.AddFlags(cmd, "tron-key", "", "tron key", true)
	utils.AddFlags(cmd, "tron-pwd", "", "tron pwd", false)
	utils.AddFlags(cmd, "fees", "FX", "fees", false)
	return cmd
}
