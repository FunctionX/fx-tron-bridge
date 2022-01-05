module github.com/functionx/fx-tron-bridge

go 1.7

require (
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/ethereum/go-ethereum v1.10.11
	github.com/fbsobreira/gotron-sdk v0.0.0-20211206103227-17533f63f585
	github.com/functionx/fx-core v0.0.0-20211012015917-76e025af60f1
	github.com/gogo/protobuf v1.3.3
	github.com/prometheus/client_golang v1.8.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.9
	go.uber.org/zap v1.15.0
	google.golang.org/grpc v1.38.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4

replace github.com/cosmos/cosmos-sdk => github.com/functionx/cosmos-sdk v0.42.5-0.20210927070625-89306d0caf62

replace github.com/fbsobreira/gotron-sdk => github.com/fx0x55/gotron-sdk v0.0.0-20211206103227-17533f63f585
