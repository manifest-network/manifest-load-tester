package manifestledgerloadtest

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	txsgen "github.com/liftedinit/manifest-load-tester/pkg/manifestledgerloadtest/txs"
	manitesttypes "github.com/liftedinit/manifest-load-tester/pkg/manifestledgerloadtest/types"
	"github.com/liftedinit/manifest-load-tester/pkg/manifestledgerloadtest/utils"
)
import "github.com/cometbft/cometbft-load-test/pkg/loadtest"

// CosmosClientFactory creates instances of CosmosClient
type CosmosClientFactory struct {
	clientCtx client.Context
	params    manitesttypes.Params
	txGens    map[string]txsgen.TxGenerator
	txWeights map[string]int
}

// CosmosClientFactory implements loadtest.ClientFactory
var _ loadtest.ClientFactory = (*CosmosClientFactory)(nil)

func NewCosmosClientFactory(clientCtx client.Context, params manitesttypes.Params) *CosmosClientFactory {
	cosmosClient := &CosmosClientFactory{
		clientCtx: clientCtx,
		params:    params,
		txGens:    make(map[string]txsgen.TxGenerator),
		txWeights: make(map[string]int),
	}

	cosmosClient.RegisterTxGenerator("bank_send", &txsgen.BankSendTxGenerator{}, 0.5)
	cosmosClient.RegisterTxGenerator("create_group", &txsgen.CreateGroupTxGenerator{}, 0.5)

	return cosmosClient
}

func (c *CosmosClientFactory) RegisterTxGenerator(name string, gen txsgen.TxGenerator, weight int) {
	c.txGens[name] = gen
	c.txWeights[name] = weight
}

// CosmosClient is responsible for generating transactions. Only one client
// will be created per connection to the remote Tendermint RPC endpoint, and
// each client will be responsible for maintaining its own state in a
// thread-safe manner.
type CosmosClient struct {
	clientCtx client.Context
	params    manitesttypes.Params
	txGens    map[string]txsgen.TxGenerator
	txWeights map[string]int
}

// CosmosClient implements loadtest.Client
var _ loadtest.Client = (*CosmosClient)(nil)

func (f *CosmosClientFactory) ValidateConfig(cfg loadtest.Config) error {
	// Do any checks here that you need to ensure that the load test
	// configuration is compatible with your client.
	return nil
}

func (f *CosmosClientFactory) NewClient(cfg loadtest.Config) (loadtest.Client, error) {
	return &CosmosClient{
		clientCtx: f.clientCtx,
		params:    f.params,
		txGens:    f.txGens,
		txWeights: f.txWeights,
	}, nil
}

// GenerateTx must return the raw bytes that make up the transaction for your
// ABCI app. The conversion to base64 will automatically be handled by the
// loadtest package, so don't worry about that. Only return an error here if you
// want to completely fail the entire load test operation.
func (c *CosmosClient) GenerateTx() ([]byte, error) {
	txType := c.selectTxType()
	generator, ok := c.txGens[txType]
	if !ok {
		return nil, fmt.Errorf("unknown transaction type: %s", txType)
	}

	sender, msg, err := generator.GenerateMsg(c.clientCtx, c.params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message: %w", err)
	}

	return c.buildAndSignTx(sender, msg)
}

func (c *CosmosClient) selectTxType() string {
	if len(c.txGens) == 1 {
		for txType := range c.txGens {
			return txType
		}
	}

	var totalWeight int
	for _, weight := range c.txWeights {
		totalWeight += weight
	}

	r := rand.Intn(totalWeight)
	var cumulativeWeight int
	for txType, weight := range c.txWeights {
		cumulativeWeight += weight
		if r < cumulativeWeight {
			return txType
		}
	}

	for txType := range c.txGens {
		return txType
	}
	return ""
}

func (c *CosmosClient) buildAndSignTx(sender *keyring.Record, msg sdk.Msg) ([]byte, error) {
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to set message: %w", err)
	}

	txBuilder.SetGasLimit(c.params.GasLimit)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(c.params.Denom, c.params.Fee)))
	txBuilder.SetMemo(utils.RandomString(255))

	return c.signTx(sender, txBuilder)
}

func (c *CosmosClient) signTx(sender *keyring.Record, txBuilder client.TxBuilder) ([]byte, error) {
	addr, err := sender.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	defaultSignMode, err := authsigning.APISignModeToInternal(c.clientCtx.TxConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, fmt.Errorf("failed to get default sign mode: %w", err)
	}

	senderPub, err := sender.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	acc, err := c.clientCtx.AccountRetriever.GetAccount(c.clientCtx, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// First round: set empty signature to gather signer infos
	sigV2 := signing.SignatureV2{
		PubKey: senderPub,
		Data: &signing.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: acc.GetSequence(),
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signature: %w", err)
	}

	// Get private key
	senderLocal := sender.GetLocal()
	senderPrivAny := senderLocal.PrivKey
	if senderPrivAny == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	senderPriv, ok := senderPrivAny.GetCachedValue().(cryptotypes.PrivKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast private key")
	}

	// Second round: sign with private key
	signerData := authsigning.SignerData{
		ChainID:       c.clientCtx.ChainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
		PubKey:        senderPub,
	}

	sigV2, err = tx.SignWithPrivKey(
		context.TODO(), defaultSignMode, signerData,
		txBuilder, senderPriv, c.clientCtx.TxConfig, acc.GetSequence())
	if err != nil {
		return nil, fmt.Errorf("failed to sign with private key: %w", err)
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signature: %w", err)
	}

	return c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
}
