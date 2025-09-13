package p2p

import (
	"fmt"
	"net"

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
	ListenAddr 		string

	//We will initiate a handshake when we connect to a peer
	//if the handshake is bad we drop the conn
	HandshakeFunc	HandshakeFunc

	//This will be used to decode the incoming messages from the peer
	Decoder			Decode


	//This function will be called when a new peer connects
	OnPeer 			func(Peer) error 
}

type TCPTransport struct{

	TCPTransportOpts
	//This will listen to the address above and hand over the incoming conncections
	listener net.Listener 

	rpcch chan RPC

}

//Consume represents the Transport interface method, which will
//return read only channel of RPC messages. recieved from another peer
func (t *TCPTransport) Consume() <- chan RPC{
	return t.rpcch
}

//Close inplements the Peer interface method
func (p *TCPPeer) Close() error{
	return p.conn.Close()
}


//This is a constructor function that returns a new instance of TCPTransport
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{

	//creates a new instance of TCPTransport and returns a pointer to it
	return &TCPTransport{
		TCPTransportOpts : 	opts,
		rpcch: 				make(chan RPC),
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


func (t *TCPTransport) handleConn(conn net.Conn){

	var err error

	defer func(){
	fmt.Printf("Dropping Peer Conn: %s", err)

	conn.Close()

	}()



	peer := NewTCPPeer(conn, true)

	
	if err = t.HandshakeFunc(peer); err != nil{
		return 
	}

	if t.OnPeer != nil{
		if err = t.OnPeer(peer); err != nil{
			return
		}
	}

	rpc := RPC{}

	
	for{

		err = t.Decoder.Decode(conn, &rpc)

		if err != nil{
			return
		}

		//returns the network address of the remote peer
		rpc.From = conn.RemoteAddr()

		t.rpcch <- rpc

		fmt.Printf("message : %+v\n", rpc)

	}

}
