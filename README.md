# manifest-load-test

Custom load tester for the Manifest Network.

This software is designed to test the performance of the Manifest Network by sending a large number of requests to the network and measuring the response time.

## Requirements

- Go 1.22

## Building the binary

Build the binary using the following command:

```bash
go build -o ./build/manifest-load-tester ./cmd/manifest-load-tester/main.go 
```

## Configuration

Create a `.env` file in the root of the project with the following content:

```bash
USER1_MNEMONIC="your mnemonic here"
USER2_MNEMONIC="your mnemonic here"  
CHAIN_ID="your chain-id here"
RPC_URL="your rpc url here"
````

where `USER1_MNEMONIC` and `USER2_MNEMONIC` are the mnemonics of the accounts you want to use for the load test, `CHAIN_ID` is the chain ID of the network you want to test, and `RPC_URL` is the URL of the RPC endpoint of the network you want to test.

Tokens will be transferred from `USER1_MNEMONIC` to `USER2_MNEMONIC` during the load test.

The `USER1_MNEMONIC` account should be funded with enough tokens to cover the cost of the transactions.

## Running the load test

The `cometbft-load-test` framework supports two modes of operation: **standalone** and **coordinator/worker**.

### Standalone

In standalone mode, the load tested will simply broadcast transactions to a single endpoint from a single binary.

Run the binary using the following command:

```bash
./build/manifest-load-tester -v \
  -c 1 \    # The number of connections to open to each endpoint simultaneously
  -T 10 \   # The duration (in seconds) for which to handle the load test
  -r 25 \   # The number of transactions to generate each second on each connection, to each endpoint
  -s 250 \  # The size of each transaction, in bytes - must be greater than 40
  --broadcast-tx-method async \
  --endpoints wss://manifest-beta-rpc.liftedinit.tech/websocket
```

See the help message for more information:

```bash
./build/manifest-load-tester --help
```

### Coordinator/Worker

TODO

## Notes

The load tester is limited by how CosmosSDK handles (ordered) transactions. Most transactions will get rejected by the mempool (due to sequence number issues) and will not be included in the block. This is expected behavior.

Unordered transactions should be supported by a future CosmosSDK version. See [ADR-070](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-070-unordered-transactions.md) for more information.

### TODO

- Script to generate a large number of accounts and fund them with tokens.
- Adapt load tester to use multiple accounts for sending transactions.
