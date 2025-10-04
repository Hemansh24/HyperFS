package main

import (
	//"bytes"
	"bytes"
	"log"
	"strings"
	"time"

	"github.com/Hemansh24/HyperFS/p2p"
)

func makeServer(listenAddr string, nodes ...string)*FileServer{

	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr: 	listenAddr,

		HandshakeFunc: 	p2p.NOPHandshakeFunc,

		Decoder: 		p2p.DefaultDecoder{},

	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	safeStorageRoot := strings.TrimPrefix(listenAddr, ":") + "_network"

	fileServerOpts := FileServerOpts{
		StorageRoot: 		safeStorageRoot,
		PathTransformFunc: 	CASPathTransformFunc,
		Transport: 			tcpTransport,	
		BootstrapNodes: 	nodes,

	}

	s := NewFileServer(fileServerOpts)

	//To implement the OnPeer func, we need to have a server running
	tcpTransport.OnPeer = s.OnPeer


	return s

	
}

func main() {
    s1 := makeServer(":3000", "")
    s2 := makeServer(":4000", ":3000")

    go func() {
        log.Fatal(s1.Start())
    }()

    time.Sleep(2 * time.Second)

    go s2.Start()
   

    // Give servers a moment to connect
    time.Sleep(2 * time.Second)

	for i := 0; i < 10; i++ {
		data := bytes.NewReader([]byte("My big data file here!"))
		s2.Store("mydata", data)
		time.Sleep(5 * time.Millisecond)
	}


    

	// r, err := s2.Get("mydata")

	// if err != nil{
	// 	log.Fatal(err)
	// }

	// b, err := io.ReadAll(r)
	// if err != nil{
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(b))
 
	select{}
    
}
