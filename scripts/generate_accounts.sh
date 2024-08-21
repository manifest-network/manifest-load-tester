#!/usr/bin/env bash

# Use this script to generate a list of accounts with their mnemonics and a genesis file with the initial balances.

CHAIN_CMD="manifestd"

usage() {
  echo "Usage: $0 -n <number_of_accounts> -m <mnemonics_output_file> -g <genesis_output_file>"
  exit 1
}

while getopts "n:m:g:" opt; do
  case $opt in
    n) NB_ACCOUNTS=$OPTARG ;;
    m) MNEMONICS_OUTPUT=$OPTARG ;;
    g) GENESIS_OUTPUT=$OPTARG ;;
    *) usage ;;
  esac
done

if [[ -z "$NB_ACCOUNTS" || -z "$MNEMONICS_OUTPUT" || -z "$GENESIS_OUTPUT" ]]; then
  usage
fi

if ! [[ "$NB_ACCOUNTS" =~ ^[0-9]+$ ]] || [ "$NB_ACCOUNTS" -le 0 ]; then
  echo "Error: Number of accounts must be a positive integer."
  usage
fi

WORKDIR=$(mktemp -d)
MNEMONICS_FILE="${WORKDIR}/mnemonics.txt"
COUNTER=1

trap 'rm -rf -- "$WORKDIR"' EXIT

# Overwrite the mnemonic file if it exists
if [[ -f "$MNEMONICS_OUTPUT" ]]; then
  true > "$MNEMONICS_OUTPUT"
fi

echo "Generating $1 mnemonics..."

for ((i=1; i <= NB_ACCOUNTS; i++));
do
  $CHAIN_CMD keys mnemonic 2>/dev/null >> "$MNEMONICS_FILE";
done

ACCOUNTS=()
while IFS="" read -r mnemonic || [ -n "$mnemonic" ]
do
  KEY="user$COUNTER"
  ADDR=$(echo "$mnemonic" | $CHAIN_CMD keys add $KEY --keyring-backend memory --recover --home="$WORKDIR" 2>/dev/null | grep "address" | awk '{print $3}')
  printf "%s\n" "$mnemonic" >> "$MNEMONICS_OUTPUT"
  ACCOUNTS+=("$ADDR")
  ((COUNTER++))
done < "$MNEMONICS_FILE"

for address in "${ACCOUNTS[@]}"; do
  jq -n \
    --arg addr "$address" \
    '{
      "balances": [
        {
          "address": $addr,
          "coins": [
            {
              "denom": "umfx",
              "amount": "10000000"
            }
          ]
        }
      ]
    }'
done | jq -s 'flatten' > "$GENESIS_OUTPUT"