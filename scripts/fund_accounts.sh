#!/usr/bin/env bash

# Set the path to your JSON file (passed as a parameter)
JSON_FILE="$1"

# Set the sender's wallet name
SENDER="manifest1hj5fveer5cjtn4wd6wstzugjfdxzl0xp8ws9ct"
NODE="http://localhost:26657"

# Set the chain ID and fees
CHAIN_ID="manifest-ledger-beta"
GAS_PRICES="0.0011umfx"
GAS_ADJUSTMENT="1.4"

# Loop through each balance entry in the JSON file
jq -c '.[] | .balances[]' $JSON_FILE | while read balance; do
    # Extract the recipient address, amount, and denomination
    RECIPIENT_ADDRESS=$(echo $balance | jq -r '.address')
    AMOUNT=$(echo $balance | jq -r '.coins[0].amount')
    DENOM=$(echo $balance | jq -r '.coins[0].denom')

    # Construct the amount with the denomination
    AMOUNT_WITH_DENOM="${AMOUNT}${DENOM}"

    echo "Sending $AMOUNT_WITH_DENOM from $SENDER to $RECIPIENT_ADDRESS : "

    # Run the transaction using manifestd
    manifestd --node $NODE tx bank send $SENDER $RECIPIENT_ADDRESS $AMOUNT_WITH_DENOM \
      --chain-id $CHAIN_ID --gas auto --gas-adjustment $GAS_ADJUSTMENT --gas-prices $GAS_PRICES \
      --keyring-backend test --yes

    # Optionally, sleep for a short time between transactions
    sleep 5
done