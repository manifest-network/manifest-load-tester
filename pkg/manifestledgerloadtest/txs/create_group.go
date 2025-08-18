package txs

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	manifesttypes "github.com/manifest-network/manifest-load-tester/pkg/manifestledgerloadtest/types"
	"github.com/manifest-network/manifest-load-tester/pkg/manifestledgerloadtest/utils"
)

type CreateGroupTxGenerator struct{}

func (g *CreateGroupTxGenerator) GenerateMsg(ctx client.Context, params manifesttypes.Params) (*keyring.Record, types.Msg, error) {
	userRandomIdx := rand.Perm(len(params.Users))[0:1]
	r1 := params.Users[userRandomIdx[0]]

	addr1, err := r1.GetAddress()
	if err != nil {
		return nil, nil, err
	}

	msg := &grouptypes.MsgCreateGroup{
		Admin: addr1.String(),
		Members: []grouptypes.MemberRequest{
			{
				Address:  addr1.String(),
				Weight:   "1",
				Metadata: "user",
			},
		},
		Metadata: utils.RandomString(params.CreateGroupMetadataSize),
	}

	return r1, msg, nil
}
