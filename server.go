package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/Hemansh24/HyperFS/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes	  []string
}

type FileServer struct{
	FileServerOpts

	peerLock sync.Mutex
	peers 	map[string]p2p.Peer

	store 	*Store

	qiutch 	chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {

	storeOpts := StoreOpts{

		Root: opts.StorageRoot,

		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{

		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		qiutch: 		make(chan struct{}),	
		peers: 		make(map[string]p2p.Peer),
	}
}

func (s *FileServer) Stop(){
	close(s.qiutch)

}

//once the server is up, OnPeer will add all the new 
//peers to the list of the current peers
func (s *FileServer) OnPeer(p p2p.Peer) error{
	s.peerLock.Lock()

	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote peer %s", p.RemoteAddr())

	return nil
}

func (s *FileServer) loop(){

	defer func(){
		log.Println("File server stopped, user quit action")
		s.Transport.Close()
	}()


	for{
		select {

			case msg := <- s.Transport.Consume():

				fmt.Println(msg)
			case <- s.qiutch:
				return 
		}
	}
}

//allows the new file server to connect to already exisiting
//p2p netwrork, by dialing down a knwon bootstrap nodes
func (s *FileServer) boostrapNetwork() error{

	//loops through a list of network address stroed in BsN
	//these are already expected to be running
	for _, addr := range(s.BootstrapNodes){
	
		if len(addr) == 0{
			continue
		}
	
		go func (addr string) {
			fmt.Println("Attempting to connect with remote ", addr)

			//for each address it launches a go routine
			//which makes it connect all the address concurrently
			//not sequentially and waiting for each ohter
		

			if err := s.Transport.Dial(addr); err != nil{
			
				log.Println("Dial error: ", err)
			}

		} (addr)

	}

	return nil
}

func (s *FileServer) Start() error{

	if err := s.Transport.ListenAndAccept(); err != nil{
		return err
	}

	
	s.boostrapNetwork()

	s.loop()

	return nil

}
