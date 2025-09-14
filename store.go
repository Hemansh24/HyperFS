package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

//Content Addressable Storage
//
func CASPathTransformFunc(key string) string{

	//hashes the key using sha1
	hash := sha1.Sum([]byte(key)) //[20]byte -> []byte : [:]

	//make the hash readable into a string
	hashStr := hex.EncodeToString(hash[:])

	//split the hash into blocks of 5 characters
	blocksize := 5

	//number of blocks
	sliceLen := len(hashStr) / blocksize

	// store individual paths and then join them with "/"
	paths := make([]string, sliceLen)


	for i := 0; i < sliceLen; i++{

		from , to := i * blocksize, (i * blocksize) + blocksize


		//slice the hash string into blocks of 5 characters
		//and store them in the paths slice
		paths[i] = hashStr[from:to]
		
	}

	//join the paths with "/" to make a directory structure
	return strings.Join(paths, "/")
}	


//takes a key and transforms it into a path
type PathTransformFunc func(string) string

//options for the store
type StoreOpts struct {
	PathTransformFunc  PathTransformFunc
}

//does not transform the path, just returns the key as is
var DefaultPathTransformFunc = func(key string) string{
	return key
}

//the main store structure
type Store struct {
	StoreOpts
}


//Constructor for strore
func NewStore(opts StoreOpts) *Store{
	return &Store{
		StoreOpts : opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error{

	//transform the key into a path
	pathName := s.PathTransformFunc(key)

	//MKdirAll makes the directory if it does not exist using
	//the path name as the directory name
	//os.ModePerm uses the default permissions which are read write execute
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil{
		return err
	}

	filename := "somefilename"

	pathAndFilename := pathName + "/" + filename

	f, err := os.Create(pathAndFilename)

	if err != nil{
		return err
	}

	n, err := io.Copy(f, r)

	if err != nil{
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, pathAndFilename)

	return nil
}