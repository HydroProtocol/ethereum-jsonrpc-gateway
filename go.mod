module github.com/HydroProtocol/ethereum-jsonrpc-gateway

go 1.13

require (
	// git.ddex.io/lib/hotconfig v0.0.0-20190812113401-7449e80bf166
	// git.ddex.io/lib/log v0.0.0-20190729100049-f91fdcf0b05c
	// git.ddex.io/lib/monitor v0.0.0-20190814083352-252de197c810
	github.com/ethereum/go-ethereum v1.9.2
	github.com/gorilla/websocket v1.4.0
	github.com/sirupsen/logrus v1.4.2
)

// replace git.ddex.io/lib/hotconfig => ../lib/hotconfig

// replace git.ddex.io/lib/log => ../lib/log

// replace git.ddex.io/lib/monitor => ../lib/monitor
