#!/bin/bash

if [ -z "$1" ]; then
    echo "No argument supplied. Usage $0 <number_of_peers>"
    exit 1
fi

if [ "$1" -gt 10 ]; then
    echo "Do not run too many nodes on a single machine."
    exit 1
fi

trap "exit" INT TERM ERR
trap "kill 0" EXIT

go get
go build -ldflags="-s -w -X main.version=0.6.0"
rm -rf ./data*

for ((n = 0; n < $1; n++)); do
    echo -e "\e[96mPeer$n\033[0m"
    if [ "$n" -eq 0 ]; then
        ./blocace s --dir ./data$n --porthttp 6899 --portP2p 6091 -l debug &
    else
        ./blocace s --dir ./data$n --porthttp $(expr 6899 + $n) --portP2p $(expr 6091 + $n) --peerAddresses localhost:6091 -l debug &
    fi
    sleep 1
done

wait
