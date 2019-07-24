# ethereum jsonrpc gateway

A transparent ethereum jsonrpc gateway

## features

- methods filter
- http
- websocket
- http upstream
- websocket upstream
- websocket upstream reconnect
- graceful shutdown


## proxy strategy

### basic

- naive     (require upstreams count == 1)

Navie strategy is the most simple one without any magic.

### advance

- race      (require upstreams count >= 2)
- fallback  (require upstreams count >= 2)

Race strategy proxy mirrors request to the all upstreams, once it receives a response for one of them, then return.

Fallback strategy proxy will retry failed request in other upstreams.
 

## config

see .env-sample file