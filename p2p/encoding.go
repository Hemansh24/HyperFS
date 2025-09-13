package p2p

import (

	"encoding/gob"

	"io"
)

// a contract, if you want to be a decoder,
//you must hava a function Decode that takes
//in an io.Reader and a pointer to Message
type Decode interface {
	Decode(io.Reader, *Message) error
}


// Structed Decoder using GOB encoding
// useful for comm bw 2 Go programs
type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *Message) error{

	return gob.NewDecoder(r).Decode(msg)
}

//Raw data decoder
//Reads the message stream and places it in the Payload field
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *Message) error{
	
	buf := make([]byte, 1028)

	n, err := r.Read(buf)

	if err != nil{
		return err
	}

	msg.Payload = buf[:n]
	return nil
}