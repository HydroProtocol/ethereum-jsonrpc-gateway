#!/bin/sh
docker rm -f ethereum-jsonrpc-gateway
# docker create --name ethereum-jsonrpc-gateway -p 3005:3005 420fdb046251
docker create --name ethereum-jsonrpc-gateway -p 3005:3005 hydroprotocolio/eth-jsonrpc-gateway
docker cp ./config.json ethereum-jsonrpc-gateway:/app/config.json
docker start ethereum-jsonrpc-gateway
docker ps | grep ethereum-jsonrpc-gateway