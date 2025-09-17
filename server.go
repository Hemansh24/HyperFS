package main

import (
	"fmt"
	"log"

	"github.com/Hemansh24/HyperFS/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
}

type FileServer struct{
	FileServerOpts

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
	}
}

func (s *FileServer) Stop(){
	close(s.qiutch)

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

func (s *FileServer) Start() error{

	if err := s.Transport.ListenAndAccept(); err != nil{
		return err
	}

	s.loop()

	return nil

}
