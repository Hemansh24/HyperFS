package p2p
import (
	"encoding/gob"
	"io"
)

// a contract, if you want to be a decoder,
//you must hava a function Decode that takes
//in an io.Reader and a pointer to Message
type Decode interface {
	Decode(io.Reader, *RPC) error
}

// Structed Decoder using GOB encoding
// useful for comm bw 2 Go programs
type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error{
	return gob.NewDecoder(r).Decode(msg)
}

//Raw data decoder
//Reads the message stream and places it in the Payload field
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error{
	
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil{
		return nil
	}
	
	stream := peekBuf[0] == IncomingStream
	//Ub tge case of a stream, we are not decoding what is being
	//sent over the network, We are just sending stream true
	//so we can handle that in a logic.
	if stream{
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil{
		return err
	}

	msg.Payload = buf[:n]
	return nil
}