package p2p


//Message hold any arbitary data that is being sent
//over the transport between 2 nodes in the network
type RPC struct{

	From string
	// A standardized envelope for all your network communications
	Payload []byte
}