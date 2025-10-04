package p2p
import "net"

//Peer is an interface that represents a remote node/peer
//anyone that we connect to or that connects to us is a peer
type Peer interface{
	net.Conn
	Send(([]byte)) 	error
	CloseStream()
}

//Transport is anything that handles the communication
//between the nodes/peers in the network
//examples of transports are TCP, UDP, WebRTC, etc
type Transport interface{
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	// This is a method for receiving messages from the network
	// returns a recieve only channel which will deliver messages
	//of RPC type
	Consume() <- chan RPC
	Close() error
}