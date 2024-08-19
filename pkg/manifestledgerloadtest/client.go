package manifestledgerloadtest

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)
import "github.com/cometbft/cometbft-load-test/pkg/loadtest"

// CosmosClientFactory creates instances of CosmosClient
type CosmosClientFactory struct {
	txConfig client.TxConfig
	kr       keyring.Keyring
}

// CosmosClientFactory implements loadtest.ClientFactory
var _ loadtest.ClientFactory = (*CosmosClientFactory)(nil)

func NewCosmosClientFactory(txConfig client.TxConfig, kr keyring.Keyring) *CosmosClientFactory {
	return &CosmosClientFactory{
		txConfig: txConfig,
		kr:       kr,
	}
}

// CosmosClient is responsible for generating transactions. Only one client
// will be created per connection to the remote Tendermint RPC endpoint, and
// each client will be responsible for maintaining its own state in a
// thread-safe manner.
type CosmosClient struct {
	txConfig client.TxConfig
	kr       keyring.Keyring
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
		txConfig: f.txConfig,
		kr:       f.kr,
	}, nil
}

// GenerateTx must return the raw bytes that make up the transaction for your
// ABCI app. The conversion to base64 will automatically be handled by the
// loadtest package, so don't worry about that. Only return an error here if you
// want to completely fail the entire load test operation.
func (c *CosmosClient) GenerateTx() ([]byte, error) {
	txBuilder := c.txConfig.NewTxBuilder()
	r1, err := c.kr.Key("user1")
	if err != nil {
		return nil, fmt.Errorf("failed to get user1 key: %w", err)
	}

	r2, err := c.kr.Key("user2")
	if err != nil {
		return nil, fmt.Errorf("failed to get user2 key: %w", err)
	}

	addr1, err := r1.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from record 1: %w", err)
	}

	addr2, err := r2.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from record 2: %w", err)
	}

	// Construct a message to send 1 umfx from addr1 to addr2
	msg1 := banktypes.NewMsgSend(addr1, addr2, types.NewCoins(types.NewInt64Coin("umfx", 1)))
	if msg1 == nil {
		return nil, fmt.Errorf("failed to create message")
	}

	err = txBuilder.SetMsgs(msg1)
	if err != nil {
		return nil, fmt.Errorf("failed to set message: %w", err)
	}

	txBuilder.SetGasLimit(200000)
	txBuilder.SetFeeAmount(types.NewCoins(types.NewInt64Coin("umfx", 5)))
	txBuilder.SetMemo("manifest-load-test")
	txBuilder.SetTimeoutHeight(5)

	defaultSignMode, err := authsigning.APISignModeToInternal(c.txConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, fmt.Errorf("failed to get default sign mode: %w", err)
	}

	r1Pub, err := r1.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key from record 1: %w", err)
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	// https://github.com/cosmos/cosmos-sdk/blob/6f30de3a41d37a4359751f9d9e508b28fc620697/baseapp/msg_service_router_test.go#L169
	sigV2 := signing.SignatureV2{
		PubKey: r1Pub,
		Data: &signing.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: 0,
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signature: %w", err)
	}

	r1Local := r1.GetLocal()
	r1PrivAny := r1Local.PrivKey
	if r1PrivAny == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	r1Priv, ok := r1PrivAny.GetCachedValue().(cryptotypes.PrivKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast private key from record 1")
	}

	// Second round: all signer infos are set, so each signer can sign.
	signerData := authsigning.SignerData{
		ChainID:       "manifest-beta-chain",
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        r1Pub,
	}

	sigV2, err = tx.SignWithPrivKey(
		context.TODO(), defaultSignMode, signerData,
		txBuilder, r1Priv, c.txConfig, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with private key: %w", err)
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signature: %w", err)
	}

	return c.txConfig.TxEncoder()(txBuilder.GetTx())
}
