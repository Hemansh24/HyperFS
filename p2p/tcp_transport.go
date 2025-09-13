package p2p

import (
	"fmt"
	"net"
	"sync"
)

//TCPPeer represents a remote node/peer in a TCP connection
type TCPPeer struct {

	//conn is the underlying connection to the peer
	conn net.Conn 

	//If we request to connect to a peer, then outbound is true
	//if we accept and retrieve a conn then inbound is true
	//and outbound is false
	outbound bool

}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer{

	return &TCPPeer{
		conn : conn,
		outbound : outbound,
	}
}

type TCPTransportOpts struct{

	//This stores the address of the peer as a string 
	ListenAddr 	string

	//We will initiate a handshake when we connect to a peer
	//if the handshake is bad we drop the conn
	HandshakeFunc	HandshakeFunc

	//This will be used to decode the incoming messages from the peer
	Decoder		Decode
}

type TCPTransport struct{

	TCPTransportOpts
	//This will listen to the address above and hand over the incoming conncections
	listener net.Listener 

	


	//This allows mulptiple go routines to access the peers map
	//and makes sure that only one go routine can write to it at a time
	mu sync.RWMutex
	//This creates a key value pair of address to peer
	peers map[net.Addr]Peer 
}


//This is a constructor function that returns a new instance of TCPTransport
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{

	//creates a new instance of TCPTransport and returns a pointer to it
	return &TCPTransport{
		TCPTransportOpts : opts,
	}
}

func(t *TCPTransport) ListenAndAccept() error{

	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)

	if(err != nil){
		return err
	}

	go t.startAcceptLoop()

	return nil

}

func (t *TCPTransport) startAcceptLoop(){

	for{
		conn, err := t.listener.Accept()
		if(err != nil){
			fmt.Printf("TCP Accept Error: %s\n", err)
		}

		
		fmt.Printf("New Incoming Connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

//type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn){

	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil{
		conn.Close()

		fmt.Printf("TCP handshake error : %s\n", err)

		return 
	}


	msg := &Message{}

	
	for{

		if err := t.Decoder.Decode(conn, msg); err != nil {

			fmt.Printf("TCP error : %s\n", err)
			continue
		}

		//returns the network address of the remote peer
		msg.From = conn.RemoteAddr()

		fmt.Printf("message : %+v\n", msg)

	}

}
