package p2p

import "net"
import "sync"
import "fmt"


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


type TCPTransport struct{
	//This stores the address of the peer as a string 
	listenAddress string
	//This will listen to the address above and hand over the incoming conncections
	listener net.Listener 

	//We will initiate a handshake when we connect to a peer
	//if the handshake is bad we drop the conn
	shakeHands HandshakeFunc

	decoder Decoder

	//This allows mulptiple go routines to access the peers map
	//and makes sure that only one go routine can write to it at a time
	mu sync.RWMutex
	//This creates a key value pair of address to peer
	peers map[net.Addr]Peer 
}


//This is a constructor function that returns a new instance of TCPTransport
func NewTCPTransport(listenAddr string) *TCPTransport{

	//creates a new instance of TCPTransport and returns a pointer to it
	return &TCPTransport{
		shakeHands: NOPHandshakeFunc,
		listenAddress: listenAddr,
	}
}

func(t *TCPTransport) ListenAndAccept() error{

	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)

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

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn){

	peer := NewTCPPeer(conn, true)

	if err := t.shakeHands(peer); err != nil{

	}


	msg := &Temp{}
	
	for{
		if err := t.decoder.Decode(conn, msg); err != nil {

			fmt.Printf("TCP error : %s\n", err)
			continue
		}

	}

}
