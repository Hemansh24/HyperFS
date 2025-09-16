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

	if pathKey.Filename != expectedOriginalKey{
		t.Errorf("have %s want %s", pathKey.Filename, expectedOriginalKey)
	}
}


func TestStore(t *testing.T) {

	s := newStore() 

	defer tearDown(t,s)

	for i := range(50){

		
		key := fmt.Sprintf("pic_%d", i)
		
		data := ([]byte("some data"))
		
		if err := s.writeStream(key, bytes.NewReader(data)); err != nil{
			t.Error(err)
		}
		
		if ok := s.Has(key); !ok{
			t.Errorf("Expected to have key: %s", key)
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
		
		if err := s.Delete(key); err != nil{
			t.Error(err)
		}

		if ok := s.Has(key); ok{
			t.Errorf("Expected to NOT have key: %s", key)
		}
	}


}

func newStore() *Store{
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	return NewStore(opts)
}

func tearDown(t *testing.T, s *Store) {

	if err := s.Clear(); err != nil{
		t.Error(err)
	}

}
