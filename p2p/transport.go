package p2p

//Peer is an interface that represents a remote node/peer
//anyone that we connect to or that connects to us is a peer
type Peer interface{

}

//Transport is anything that handles the communication
//between the nodes/peers in the network
//examples of transports are TCP, UDP, WebRTC, etc

type Transport interface{

	ListenAndAccept() error

}