package main

import (
	"bytes"
	"fmt"
	"io"
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

func TestStoreDelete(t *testing.T){

	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	key := "mypic"

	data := ([]byte("some data"))

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil{
		t.Error(err)
	}

	if err := s.Delete(key); err != nil{
		t.Error(err)
	}

}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	key := "mypic"

	data := ([]byte("some data"))

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil{
		t.Error(err)
	}

	r, err := s.Read(key)

	if err != nil{
		t.Error(err)
	}

	b, _ := io.ReadAll(r)

	fmt.Println(string(b))

	if string(b) != string(data){
		t.Errorf("want %s have %s", data, b)
	}


	s.Delete(key)

}
