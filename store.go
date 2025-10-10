package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

//Content Addressable Storage
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

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")

	if len(paths) == 0{
		return ""
	}
	return paths[0]
}

//Adds both PathName of a Pathkey and 
// the filename of a Pathkey
func (p PathKey) FullPath() string{
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

//options for the store
type StoreOpts struct {
	//Root is the folder name of the root, containing
	//all the folders/files of the system
	Root 				string
	PathTransformFunc  	PathTransformFunc
}

//does not transform the path, just returns the key as is
var DefaultPathTransformFunc = func(key string) PathKey{
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

//the main store structure
type Store struct {
	StoreOpts
}


//Constructor for strore
func NewStore(opts StoreOpts) *Store{

	if len(opts.Root) == 0{
		opts.Root = defaultRootFolderName
	}

	if opts.PathTransformFunc == nil{
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	return &Store{
		StoreOpts : opts,
	}
}

func (s *Store) Has(key string) bool{
	pathKey := s.PathTransformFunc(key)

	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error{
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error{
	
	pathKey := s.PathTransformFunc(key)

	defer func (){
		log.Printf("Deleted [%s] from disk", pathKey.Filename)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())
	
	//os.RemoveAll removes the entire directory tree
	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Write(key string, r io.Reader) (int64, error){
	return s.writeStream(key,r)
}

func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (int64, error){
	
	f, err := s.openFileForWriting(key)
	if err != nil{
		return 0, err
	}
	defer f.Close()
	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Store) writeStream(key string, r io.Reader) (int64 ,error){
	f, err := s.openFileForWriting(key)
	if err != nil{
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

func (s *Store) Read(key string) (int64,io.Reader, error){
	return s.readStream(key)
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error){
	pathKey := s.PathTransformFunc(key)
	fullPathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	file, err := os.Open(fullPathKeyWithRoot)
	if err != nil{
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil{
		return 0, nil, err
	}
	return fi.Size(), file, nil
}



func (s *Store) openFileForWriting(key string) (*os.File, error){
	//transform the key into a path
	pathKey := s.PathTransformFunc(key)

	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	//MKdirAll makes the directory if it does not exist using
	//the path name as the directory name
	//os.ModePerm uses the default permissions which are read write execute
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil{
		return nil, err
	}

	fullPath := pathKey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, fullPath)

	return os.Create(fullPathWithRoot)
}



