# ethereum-jsonrpc-gateway

A transparent ethereum jsonrpc gateway

## Requirements

- Go version >= 1.11

## Features

- Methods filter
- Smart Contract whitelist
- HTTP
- HTTP upstream
- Websocket
- Websocket rpstream
- Websocket upstream reconnect
- Hot reload configuration
- Graceful shutdown

## Proxy Strategy

### Basic

- naive (require upstreams count == 1)
  Navie strategy is the most simple one without any magic.

### Advanced

- race (require upstreams count >= 2)
  Race strategy proxy mirrors request to the all upstreams, once it receives a response for one of them, then return.

- fallback (require upstreams count >= 2)
  Fallback strategy proxy will retry failed request in other upstreams.

## Getting Started

### Build From Source

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
2. docker run

```
docker run -t -p 3005:3005  hydroprotocolio/ethereum-jsonrpc-gateway
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
