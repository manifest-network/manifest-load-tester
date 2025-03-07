package types

import "github.com/cosmos/cosmos-sdk/crypto/keyring"

type Params struct {
	Users                   []*keyring.Record
	Fee                     int64
	Amount                  int64
	Denom                   string
	GasLimit                uint64
	CreateGroupMetadataSize uint64
}
