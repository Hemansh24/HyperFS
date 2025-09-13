package p2p

import "net"

//Message hold any arbitary data that is being sent
//over the transport between 2 nodes in the network
type Message struct{

	From net.Addr 
	// A standardized envelope for all your network communications
	Payload []byte
}