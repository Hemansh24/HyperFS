package p2p

const(
	IncomingMessage = 0x1
	IncomingStream = 0x2
)

//Message hold any arbitary data that is being sent
//over the transport between 2 nodes in the network
type RPC struct{

	From 	string
	// A standardized envelope for all your network communications
	Payload []byte
	//lets us know if we are streaming or not
	Stream 	bool
}