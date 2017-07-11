package protocol

// Bitcoin protocol constants for this node
var PROTOCOL_VERSION int32 = 70015
var CADDR_TIME_VERSION uint32 = 31402
var MAINNET_MAGIC uint32 = 0xD9B4BEF9 // bitcoin main network
var MAINNET_TCP_PORT uint16 = 8333 // bitcoin main network port
var TESTNET_MAGIC uint32 = 0xDAB5BFFA // bitcoin test network
var TESTNET_TCP_PORT uint16 = 18333 // bitcoin test network port
var NODE_SERVICES uint64 = 0xd
