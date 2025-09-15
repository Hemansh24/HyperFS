package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

//Content Addressable Storage
//
func CASPathTransformFunc(key string) PathKey{

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

	return PathKey{
		PathName : strings.Join(paths, "/"),
		Filename: hashStr,
	}

}	


//takes a key and transforms it into a path
type PathTransformFunc func(string) PathKey

type PathKey struct{
	PathName string
	Filename string
}

func (p PathKey) FullPath() string{
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

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

func (s *Store) Has(key string) bool{
	pathKey := s.PathTransformFunc(key)

	_, err := os.Stat(pathKey.FullPath())

	if err == fs.ErrNotExist{
		return false
	}

	return true
}

func (s *Store) Delete(key string) error{
	
	pathKey := s.PathTransformFunc(key)

	defer func (){
		log.Printf("Deleted [%s] from disk", pathKey.Filename)
	}()
	
	//os.RemoveAll removes the entire directory tree
	return os.RemoveAll(pathKey.FullPath())
}


func (s *Store) Read(key string) (io.Reader, error){

	f, err := s.readStream(key)

	if err != nil{
		return nil, err
	}

	//defer makes sure the file is closed before the Read func ends
	defer f.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error){
	pathKey := s.PathTransformFunc(key)

	return os.Open(pathKey.FullPath())

}

func (s *Store) writeStream(key string, r io.Reader) error{

	//transform the key into a path
	pathKey := s.PathTransformFunc(key)

	//MKdirAll makes the directory if it does not exist using
	//the path name as the directory name
	//os.ModePerm uses the default permissions which are read write execute
	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil{
		return err
	}

	fullPath := pathKey.FullPath()

	f, err := os.Create(fullPath)

	if err != nil{
		return err
	}

	defer f.Close()

	n, err := io.Copy(f, r)

	if err != nil{
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, fullPath)

	return nil
}

