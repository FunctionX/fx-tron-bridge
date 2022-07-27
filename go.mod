module github.com/functionx/fx-tron-bridge

go 1.7

require (
	github.com/cosmos/cosmos-sdk v0.45.5
	github.com/ethereum/go-ethereum v1.10.16
	github.com/fbsobreira/gotron-sdk v0.0.0-20211206103227-17533f63f585
	github.com/functionx/fx-core/v2 v2.2.0
	github.com/gogo/protobuf v1.3.3
	github.com/prometheus/client_golang v1.12.2
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.7.5
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.47.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4

replace github.com/fbsobreira/gotron-sdk => github.com/fx0x55/gotron-sdk v0.0.0-20211206103227-17533f63f585

replace github.com/evmos/ethermint => github.com/functionx/ethermint v0.17.0-fxcore-v2.1.0
