# ethereum-jsonrpc-gateway

A transparent gateway on top of ethereum node for load-balancing, permissions checking, with multi flexible accesses.

## Why this project matters?

To avoid single point failure, we can use several ethereum nodes. It doesn’t guarantee all of these nodes are all available at the same time, but at least one of them will be working. It’s because we may upgrade some of the nodes regularly, or some nodes may be in a syncing state. That why we need a transparent gateway on top of these nodes. This layer gateway can temporarily get rid of the underling nodes which are not working, so the upper layer services will not notice about the unusable. This gateway also benefits load balances about rpc requests. Furthermore, we can add some permission check in the gateway layer. Only specific contracts or addresses are allowed to access and specific methods are allowed to call.

## Features

- Permisson check - Methods filter
- Permisson check - Smart Contract whitelist
- HTTP
- HTTP upstream
- Websocket
- Websocket rpstream
- Websocket upstream reconnect
- Hot reload configuration
- Graceful shutdown
- Archive data router

## Proxy Strategy

### Basic

- naive (require upstreams count == 1)
  Navie strategy is the most simple one without any magic.
  <img src="./assets/strategy1.png">

### Advanced

- race (require upstreams count >= 2)
  Race strategy proxy mirrors request to the all upstreams, once it receives a response for one of them, then return.
  <img src="./assets/strategy2.png">

- fallback (require upstreams count >= 2)
  Fallback strategy proxy will retry failed request in other upstreams.
  <img src="./assets/strategy3.png">

## Getting Started

### Build From Source

Requirements Go version >= 1.11

1. Clone this repo
2. Copy .config.sample.json to .config.json and Set valid Configuration
3. Install the dependencies:

```
go mod download
```

4. Run

```
go build cmd/main.go
./main # Started on port 3005
```

### Run Using Docker

1. Pull the docker image

```
docker pull  hydroprotocolio/ethereum-jsonrpc-gateway
```

2. docker run

```
docker run -t -p 3005:3005  hydroprotocolio/ethereum-jsonrpc-gateway
```

### Use it

We call `eth_blockNumber` method (When set `methodLimitationEnabled` true, or `eth_blockNumber` in `allowedMethods`)

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:3005

{"jsonrpc":"2.0","id":1,"result":"0x6c1100"}%
```

And if we set `methodLimitationEnabled` true, and `eth_blockNumber` is not in `allowedMethods`, when we call `eth_blockNumber`, gateway will deny the reqeust.

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:3005

{"error":{"code":-32602,"message":"not allowed method"},"id":1,"jsonrpc":"2.0"}%
```

## Configuration

Copy .config.sample.json to .config.json then edit .config.json

### upstreams

Support http, https, ws, wss.
eg.

```
  "upstreams": [
    "https://example.com/api/v1"
  ]
```

### strategy

Support NAIVE, RACE, FALLBACK
Learn More about the Proxy Strategy
eg.

```
  "strategy": "NAIVE"
```

### methodLimitationEnabled

ture or false, if set false will ignore `allowedMethods` and `contractWhitelist`.
eg.

```
  "methodLimitationEnabled": false
```

### allowedMethods

Allowed call methods, Can be ignored when set `methodLimitationEnabled` false
eg.

```
  "allowedMethods": ["eth_getBalance"]
```

### contractWhitelist

Contract Whitelist, Can be ignored when set `methodLimitationEnabled` false

```
  "contractWhitelist": ["0x..."]
```

## Contributing

1. Fork it (<https://github.com/HydroProtocol/ethereum-jsonrpc-gateway/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details
