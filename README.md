# manifest-load-test

Custom load tester for the Manifest Network.

This software is designed to test the performance of the Manifest Network by sending a large number of requests to the network and measuring the response time.

## Requirements

- Go 1.22
- jq 1.6

## Building the binary

Build the binary using the following command:

```bash
go build -o ./build/manifest-load-tester ./cmd/manifest-load-tester/main.go 
```

## Configuration

Create a `.env` file in the root of the project with the following content:

```bash
USER_MNEMONICS_FILE="path to file with user mnemonics"
CHAIN_ID="your chain-id here"
RPC_URL="your rpc url here"
DENOM=umfx
AMOUNT=1
GAS_LIMIT=20000000
FEE=100000
CREATE_GROUP_METADATA_SIZE=2048
````

where `USER_MNEMONICS_FILE` is the file containing the mnemonics (one per line) of the accounts you want to use for the load test, 
      `CHAIN_ID` is the chain ID of the network you want to test, 
      `RPC_URL` is the URL of the RPC endpoint of the network you want to test, 
      `DENOM` is the denomination of the token you want to use for the load test, 
      `AMOUNT` is the amount of tokens to send in each transaction, 
      `GAS_LIMIT` is the gas limit for each transaction, 
      `FEE` is the fee for each transaction, and 
      `CREATE_GROUP_METADATA_SIZE` is the size of the metadata for the group creation transaction.

Two transaction types are supported: `create_group` and `send`. The load tester will randomly select one of the two transaction types for each transaction.

The `send` transaction will transfer tokens from a randomly selected account in the `USER_MNEMONICS_FILE` file to another randomly selected account in the same file.
The `create_group` transaction will create a group with a randomly selected account in the `USER_MNEMONICS_FILE` file as the group creator and group member. The group metadata will be a random string of size `CREATE_GROUP_METADATA_SIZE` bytes.

The `USER_MNEMONICS_FILE` file can be generated using the `generate_accounts.sh` script in the `scripts` directory.

The `generate_accounts.sh` script will also generate a genesis balance segment for each account in the `USER_MNEMONICS_FILE` file. This segment should be added to the genesis file of the network you want to test. 
The `fund_accounts.sh` script can be used to fund the accounts with the generated genesis balance segment.

See the tools section below for more information.

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

## Tools

The `scripts` directory contains the `generate_accounts.sh` script, which can be used to generate accounts and mnemonics for the load test.

```bash
# Generate 25 accounts 
# The `mnenomics.txt` file will contain the mnemonics of the generated accounts, one per line. 
# The `genesis_balance.json` file will contain the genesis balance segment for each account.
./scripts/generate_accounts.sh -n 25 -m mnemonics.txt -g genesis_balance.json
```

The `fund_accounts.sh` script can be used to fund the accounts with the generated genesis balance segment.

```bash 
# Fund the accounts with the generated genesis balance segment
# The `genesis_balance.json` file will be used to fund the accounts.
./scripts/fund_accounts.sh genesis_balance.json
```

