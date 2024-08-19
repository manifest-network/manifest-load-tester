package main

import (
	"log/slog"
	"os"

	"github.com/cometbft/cometbft-load-test/pkg/loadtest"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/joho/godotenv"
	"github.com/liftedinit/manifest-load-tester/pkg/manifestledgerloadtest"
)

const CoinType = 118

var HdPath = hd.CreateHDPath(CoinType, 0, 0)

func recordFromMnmonic(kr keyring.Keyring, name, mnemonic string) (*keyring.Record, error) {
	record, err := kr.NewAccount(name, mnemonic, "", HdPath.String(), hd.Secp256k1)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func defaultEncoding() testutil.TestEncodingConfig {
	return testutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				paramsclient.ProposalHandler,
			},
		),
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		consensus.AppModuleBasic{},
	)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	user1Mnemonic := os.Getenv("USER1_MNEMONIC")
	if user1Mnemonic == "" {
		panic("USER1_MNEMONIC env var not set")
	}

	user2Mnemonic := os.Getenv("USER2_MNEMONIC")
	if user2Mnemonic == "" {
		panic("USER2_MNEMONIC env var not set")
	}

	types.GetConfig().SetBech32PrefixForAccount("manifest", "manifestpub")

	enc := defaultEncoding()
	cdc := codec.NewProtoCodec(enc.InterfaceRegistry)
	kr := keyring.NewInMemory(cdc)

	user1, err := recordFromMnmonic(kr, "user1", user1Mnemonic)
	if err != nil {
		panic(err)
	}
	addr1, err := user1.GetAddress()
	if err != nil {
		panic(err)
	}
	slog.Info("User1 address: ", "addr", addr1.String())

	user2, err := recordFromMnmonic(kr, "user2", user2Mnemonic)
	if err != nil {
		panic(err)
	}
	addr2, err := user2.GetAddress()
	if err != nil {
		panic(err)
	}
	slog.Info("User2 address: ", "addr", addr2.String())

	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	cosmosClientFactory := manifestledgerloadtest.NewCosmosClientFactory(txConfig, kr)
	if err := loadtest.RegisterClientFactory("manifest-ledger-load-test", cosmosClientFactory); err != nil {
		panic(err)
	}

	loadtest.Run(&loadtest.CLIConfig{
		AppName:              "manifest-load-tester",
		AppShortDesc:         "Load testing application for the Manifest Ledger App (TM)",
		DefaultClientFactory: "manifest-ledger-load-test",
	})
}
