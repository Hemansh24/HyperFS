package p2p

//HandshakeFunc is a function that takes in any type and returns an error
type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error {return nil}

