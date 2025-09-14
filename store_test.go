package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestTransformFunc(t *testing.T) {

	key := "mykey"

	pathName := CASPathTransformFunc(key)

	fmt.Println("Path Name: ", pathName)

	expectedPath := "816cc/20437/d8595/38736/e1ef4/6558b/7bda4/86c06"

	if pathName != expectedPath{
		t.Errorf("have %s want %s", pathName, expectedPath)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("some data"))

	if err := s.writeStream("mykey", data); err != nil{
		t.Error(err)
	}
}
