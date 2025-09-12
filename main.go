package main

import (


	"github.com/Hemansh24/HyperFS/p2p"

	"log"
)



func main(){

	tr := p2p.NewTCPTransport(":3000")

	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}

	select{}

}
