package p2p
import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

//TCPPeer represents a remote node/peer in a TCP connection
type TCPPeer struct {
	//conn is the underlying connection to the peer, which
	//in this case is a TCP conn
	net.Conn 

	//If we request to connect to a peer, then outbound is true
	//if we accept and retrieve a conn then inbound is true
	//and outbound is false
	outbound bool
	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer{
	return &TCPPeer{
		Conn : conn,
		outbound : outbound,
		wg: 	&sync.WaitGroup{},
	}
}

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
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

func (p *TCPPeer) Send(b []byte) error{
	_, err := p.Conn.Write(b)
	return err
}

//This is a constructor function that returns a new instance of TCPTransport
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{
	//creates a new instance of TCPTransport and returns a pointer to it
	return &TCPTransport{
		TCPTransportOpts : 	opts,
		rpcch: 				make(chan RPC, 1024),
	}
}
//Addr implements tranport interface, returns the address
//from which the transport is accepting connections
func (t *TCPTransport) Addr() string{
	return t.ListenAddr
}

//Consume represents the Transport interface method, which will
//return read only channel of RPC messages. recieved from another peer
func (t *TCPTransport) Consume() <- chan RPC{
	return t.rpcch
}

//Close implements the Transport interface 
func (t *TCPTransport) Close() error{

	return t.listener.Close()
}

//This initaites an outbound call to other peers
//which means we can connect to those and then move on
//to the peers in that network
func (t *TCPTransport) Dial(addr string) error{

	conn, err := net.Dial("tcp", addr)

	if err != nil{
		return err
	}

	go t.handleConn(conn, true)

	return nil

}

func(t *TCPTransport) ListenAndAccept() error{

	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)



	if(err != nil){
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP Transport Listening on Port: %s\n", t.ListenAddr)

	return nil

}

func (t *TCPTransport) startAcceptLoop(){

	for{
		conn, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed){
			return
		}

		if(err != nil){
			fmt.Printf("TCP Accept Error: %s\n", err)
		}

		go t.handleConn(conn, false)
	}
}


func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		t.rpcch <- rpc
	}
}