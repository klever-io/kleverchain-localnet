#!/bin/bash

VALIDATORS=1

if [ ! -z "$VALIDATORS_NUM" ]; then
    VALIDATORS=$VALIDATORS_NUM
fi

for (( INDEX=0; INDEX<VALIDATORS; INDEX++ )); do
    # Create directories
    mkdir -p "$(pwd)/dbs/node-$INDEX"
    mkdir -p "$(pwd)/logs/node-$INDEX"
    
    # Print the index and result for debugging
    echo "Creating directories for node-$INDEX"
done
