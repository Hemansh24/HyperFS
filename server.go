package main
//Capital is public

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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



func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			log.Printf("broadcast to peer %s failed: %s", peer.RemoteAddr(), err)
		}
	}

	return nil
}
type Message struct{
	Payload any
}

type MessageStoreFile struct{
	Key string
	Size int64
}

func (s *FileServer) StoreData(key string, r io.Reader) error{

	// 1. Store this file in disk

	// 2. Broadcast this file to every known peer in the network

	buf := new(bytes.Buffer)

	tee := io.TeeReader(r, buf)

	size, err := s.store.Write(key, tee)
	if err != nil{
		 return err
	}
		
	msg := Message{
		Payload: MessageStoreFile{
			Key : key,
			Size: size,
		},
	}
	msgbuf := new(bytes.Buffer)

	//encodes the storage key for transmission
	if err := gob.NewEncoder(msgbuf).Encode(msg); err != nil{
		return err
	}

	//sends the small, gob encoded metadata message to every peer
	for _, peer := range(s.peers){

		if err := peer.Send(msgbuf.Bytes()); err != nil{
			return err
		}
	}

	time.Sleep(time.Second * 3)

	//sends the payload message to every peer
	for _, peer := range(s.peers){
		n, err := io.Copy(peer, buf)

		if err != nil{
			return err
		}

		fmt.Println("recv and written to disk: ", n)
	}

	return nil

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

		case rpc := <- s.Transport.Consume():

			var msg Message
			//decodes the msg struct
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil{
				log.Println(err)
				return
			}

			if err := s.handleMessage(rpc.From, &msg); err !=nil{
				log.Println(err)
				return 
			}

		case <- s.qiutch:
			return 
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error{
	switch v := msg.Payload.(type) {

		case MessageStoreFile:

			return s.handleMessageStoreFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error{
	
	peer, ok := s.peers[from]

	if !ok{
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	if _, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil{
		return err
	}


	peer.(*p2p.TCPPeer).Wg.Done()

	return nil

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

func init(){
	gob.Register(MessageStoreFile{})
}
