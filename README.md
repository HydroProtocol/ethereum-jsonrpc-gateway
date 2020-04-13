# ethereum-jsonrpc-gateway

[![Go Report Card](https://goreportcard.com/badge/github.com/HydroProtocol/ethereum-jsonrpc-gateway)](https://goreportcard.com/report/github.com/HydroProtocol/ethereum-jsonrpc-gateway)
[![CircleCI](https://circleci.com/gh/HydroProtocol/ethereum-jsonrpc-gateway.svg?style=svg)](https://circleci.com/gh/HydroProtocol/ethereum-jsonrpc-gateway)
[![Docker Cloud Automated build](https://img.shields.io/docker/cloud/automated/hydroprotocolio/ethereum-jsonrpc-gateway.svg)](https://hub.docker.com/r/hydroprotocolio/ethereum-jsonrpc-gateway)
[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/hydroprotocolio/ethereum-jsonrpc-gateway.svg)](https://hub.docker.com/r/hydroscanio/ethereum-jsonrpc-gateway)

A transparent gateway on top of Ethereum nodes for load-balancing, permissions checking.

## Background

Services that use the Ethereum blockchain typically need to maintain multiple [Ethereum nodes](https://docs.ethhub.io/using-ethereum/running-an-ethereum-node/) in order to interact with on-chain data. Maintaining multiple Ethereum nodes creates a vast array of complications that eth-jsonrpc-gateways helps allieviate.

Using only a single node, while simpler than running multiple, often is insufficient for practical applications (and yields a singular point of failure). Instead, using a series of multiple Ethereum nodes is a standard practice.

ethereum-jsonrpc-gateway was created as a more elegant solution for Ethereum node management. Some of the complexities it addresses are:

- Maintaining uptime while nodes are upgraded and synced frequently

Not all nodes will be available 100% of the time, but eth-jsonrpc-gateway acts as a transparent gateway on top of these nodes: assuring that at least some of them will be available to prevent application failure.

- Load balancing and permission checks are built in

The gateway also acts as a load balancer across the nodes for [rpc](https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html) requests. It can also choose to only accept calls from specific addresses and smart contracts.

## Features

- Permisson check - Methods filter. You can set allowed methods in configuration, and only allowed methods can be called.
- Permisson check - Smart Contract whitelist. Contracts only in this whitelist can be called.
- HTTP and Websocket connection. Support http, http upstream, websocket, websocket upstream and websocket reconnection.
- Server proxy strategies. There are three strategies you can choose: NAIVE, RACE, and FALLBACK.
- Hot reload configuration. When change the configuration, you don't need restart the server, it will auto load the configuration.
- Graceful shutdown. When receive shutdown signal, it will shutdown gracefully after handle current requests without bad responses.
- Archive data router. Gateway will choose an archive node can serve API request for certain RPC methods older than 128 blocks.

## Getting Started

There are two ways you can install and run eth-jsonrpc-gateway: you can build it from the source, or you can use a docker container. We'll go over both here.

### Build From Source

#### Requirements

Go version >= 1.11

#### Steps

1. Clone this repo
2. Copy .config.sample.json to .config.json and set valid configuration. [Learn More](#configuration) about configuration
3. Install the dependencies:

```
go mod download
```

4. Run

```
go build .
./ethereum-jsonrpc-gateway start     # Started on port 3005
```

### Run Using Docker

1. Clone this repo
2. Copy `.config.sample.json` to `.config.json` and set valid configuration. [Learn More](#configuration) about configuration
3. docker run

```
chmod +x docker-run.sh
./docker-run.sh
```

### Usage

We call the `eth_blockNumber` method (When set `methodLimitationEnabled` true, or `eth_blockNumber` in `allowedMethods`)

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:3005

{"jsonrpc":"2.0","id":1,"result":"0x6c1100"}%
```

And if we set `methodLimitationEnabled` true, and `eth_blockNumber` is not in `allowedMethods`, when we call `eth_blockNumber` the gateway will deny the reqeust.

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:3005

{"error":{"code":-32602,"message":"not allowed method"},"id":1,"jsonrpc":"2.0"}%
```

## Configuration

Copy `.config.sample.json` to `.config.json` then edit `.config.json`

### upstreams

Ethereum node upstream urls. You can set multiple nodes in this list. And upstream support http, https, ws, wss.
eg.

```
  "upstreams": [
    "https://example.com/api/v1"
  ]
```

### oldTrieUrl

This field is for Archive Data. If you set `oldTrieUrl`, Gateway will route Archive Data to this url. An archive node is a simplified way of identifying an Ethereum full node running in archive mode. If you are interested in inspecting historical data (data outside of the most recent 128 blocks), your request requires access to archive data.
[Learn More](https://infura.io/docs/ethereum/add-ons/archiveData) about Archive Data.
eg.

```
  "oldTrieUrl": "https://example2.com/api/v1",
```

### strategy

There are three strategies: `NAIVE`, `RACE`, `FALLBACK`. [Learn More](#proxy-strategy) about the Proxy Strategy.
eg.

```
  "strategy": "NAIVE"
```

### methodLimitationEnabled

This field is about wether enabled the method limitation. The value of this field can be ture or false, if set false will ignore `allowedMethods` and `contractWhitelist`.
eg.

```
  "methodLimitationEnabled": false
```

### allowedMethods

Allowed call methods, If `methodLimitationEnabled` is true, only methods in this list can be called. Can be ignored when set `methodLimitationEnabled` false.
eg.

```
  "allowedMethods": ["eth_getBalance"]
```

### contractWhitelist

Contract Whitelist,I f `methodLimitationEnabled` is true, only contract in in this whitelist can be called. Can be ignored when set `methodLimitationEnabled` false

```
  "contractWhitelist": ["0x..."]
```

## Proxy Strategy

Depending on the level of complexity needed, there are three proxy strategies for eth-jsonrpc-gateway: `Naive`, `Race` and `Fallback`. The pictures below display how these different proxy methods work.

### Naive

- Naive require upstreams count == 1
  Naive strategy is the most simple one without any magic.
  <img src="./assets/strategy1.png">

### Race

- Race require upstreams count >= 2
  Race strategy proxy mirrors request to the all upstreams, once it receives a response for one of them, then return.
  <img src="./assets/strategy2.png">

### Fallback

- Fallback require upstreams count >= 2
  Fallback strategy proxy will retry failed request in other upstreams.
  <img src="./assets/strategy3.png">

## Contributing

1. Fork it (<https://github.com/HydroProtocol/ethereum-jsonrpc-gateway/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details
