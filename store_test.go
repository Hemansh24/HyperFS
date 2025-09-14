package main

import (
	"bytes"
	"testing"
)

func TestTransformFunc(t *testing.T) {

	key := "mykey"

	pathKey := CASPathTransformFunc(key)


	expectedOriginalKey := "816cc20437d859538736e1ef46558b7bda486c06"

	expectedPathName := "816cc/20437/d8595/38736/e1ef4/6558b/7bda4/86c06"

	if pathKey.PathName != expectedPathName{
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}

	if pathKey.Filename != expectedPathName{
		t.Errorf("have %s want %s", pathKey.Filename, expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("some data"))

	if err := s.writeStream("mykey", data); err != nil{
		t.Error(err)
	}
}
