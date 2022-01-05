package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func UpdateAddressPrefix(prefix string) {
	config := sdk.GetConfig()
	*config = *sdk.NewConfig()

	sdk.SetCoinDenomRegex(func() string {
		return `[a-zA-Z][a-zA-Z0-9/-]{1,127}`
	})

	config.SetBech32PrefixForAccount(prefix, prefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(prefix+sdk.PrefixValidator+sdk.PrefixOperator, prefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(prefix+sdk.PrefixValidator+sdk.PrefixConsensus, prefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
	config.Seal()
}
