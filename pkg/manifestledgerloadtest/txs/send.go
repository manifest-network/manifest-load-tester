package txs

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	manifesttypes "github.com/liftedinit/manifest-load-tester/pkg/manifestledgerloadtest/types"
)

type BankSendTxGenerator struct{}

func (g *BankSendTxGenerator) GenerateMsg(ctx client.Context, params manifesttypes.Params) (*keyring.Record, types.Msg, error) {
	userRandomIdx := rand.Perm(len(params.Users))[0:2]
	r1 := params.Users[userRandomIdx[0]]
	r2 := params.Users[userRandomIdx[1]]

	addr1, err := r1.GetAddress()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get address from record 1: %w", err)
	}
	addr2, err := r2.GetAddress()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get address from record 2: %w", err)
	}

	return r1, banktypes.NewMsgSend(addr1, addr2, types.NewCoins(types.NewInt64Coin(params.Denom, params.Amount))), nil
}
