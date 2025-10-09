package main

//Capital is public
import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/Hemansh24/HyperFS/p2p"
)

type FileServerOpts struct {
	Enckey 				[]byte
	StorageRoot       	string
	PathTransformFunc 	PathTransformFunc
	Transport         	p2p.Transport
	BootstrapNodes	  	[]string
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
		peers: 			make(map[string]p2p.Peer),
	}
}

func (s *FileServer) stream(msg *Message) error {
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

func (s *FileServer) broadcast(msg *Message) error {

	buf := new(bytes.Buffer)
	//encodes the storage key for transmission
	if err := gob.NewEncoder(buf).Encode(msg); err != nil{
		return err
	}

	//sends the small, gob encoded metadata message to every peer
	for _, peer := range(s.peers){
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil{
			return err
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

type MessageGetFile struct{
	Key string
}

//checks if the server already has the key or not
func (s *FileServer) Get(key string) (io.Reader, error){
	if s.store.Has(key){
		fmt.Printf("[%s] serving file (%s) from local\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(key)
		return r, err
	}
	fmt.Printf("[%s] Dont have file (%s) locally, fetching from network... \n", s.Transport.Addr(), key)
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil{
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers{
		//First read the file size so that we can limit the amount of bytes we read from the connection
		//so that we don't keep hanging
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.Write(key, io.LimitReader(peer, fileSize)); 
		if err != nil{
			return nil, err
		}

		fmt.Printf("[%s] Recieved bytes (%d) over the network from (%s)\n",s.Transport.Addr(), n, peer.RemoteAddr())
		peer.CloseStream()
	}

	_, r, err := s.store.Read(key)
	return r, err
}

func (s *FileServer) Store(key string, r io.Reader) error{

	var (
		filebuffer = new(bytes.Buffer)
		tee = io.TeeReader(r, filebuffer)
	)
	size, err := s.store.Write(key, tee)
	if err != nil{
		 return err
	}
		
	msg := Message{
		Payload: MessageStoreFile{
			Key : key,
			Size: size + 16,
		},
	}
	if err := s.broadcast(&msg); err != nil{
		return err
	}

	time.Sleep(time.Millisecond * 5)

	//sends the payload message to every peer
	for _, peer := range(s.peers){
		peer.Send([]byte{p2p.IncomingStream})
		n, err := copyEncrypt(s.Enckey, filebuffer, peer)
		if err != nil{
			return err
		}
		// n, err := io.Copy(peer, filebuffer)
		// if err != nil{
		// 	return err
		// }
		fmt.Println("recv and written to disk: ", n)
		time.Sleep(time.Second * 1)
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
		log.Println("File server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for{
		select {

		case rpc := <- s.Transport.Consume():
			var msg Message
			//decodes the msg struct
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil{
				log.Println("decoding error: ",err)
			}
			if err := s.handleMessage(rpc.From, &msg); err !=nil{
				log.Println("handle message error: ",err)
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
		case MessageGetFile:
			return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error{
	if !s.store.Has(msg.Key){
		return fmt.Errorf("[%s] need to serve file but (%s) does not exist on disk", s.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] Serving file (%s) over the network\n",s.Transport.Addr(), msg.Key)
	fileSize, r, err := s.store.Read(msg.Key)
	if err != nil{
		return err
	}

	if rc, flag := r.(io.ReadCloser); flag{
		fmt.Println("Closing ReadCloser")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}
	//first send the "IncomingStream" byte to the peer and then we can send the file size
	//as an INT64
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil{
		return err
	}
	fmt.Printf("[%s] written (%d) bytes over the network to %s\n",s.Transport.Addr(), n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error{
	peer, ok := s.peers[from]
	if !ok{
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil{
		return err
	}
	fmt.Printf("[%s] written %d bytes to disk \n",s.Transport.Addr(), n)
	peer.CloseStream()
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
	gob.Register(MessageGetFile{})

}
