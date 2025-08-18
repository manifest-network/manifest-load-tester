package txs

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	manifesttypes "github.com/manifest-network/manifest-load-tester/pkg/manifestledgerloadtest/types"
)

type TxGenerator interface {
	// GenerateMsg generates a transaction message
	GenerateMsg(ctx client.Context, params manifesttypes.Params) (*keyring.Record, types.Msg, error)
}
