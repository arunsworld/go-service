package binaryio

import (
	"archive/zip"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sync"
)

// Repo is a repository for binary data
type Repo interface {
	Close()
	Create() (io.Writer, error)
	CreateAndWrite(interface{}) error
	Shards() int
	Open(int) (io.ReadCloser, error)
	Read(int, interface{}) error
}

// NewRepo creates a new repo for binary data
func NewRepo(filename string) (Repo, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	zw := zip.NewWriter(f)
	return &repo{
		fileWriter: f,
		zipWriter:  zw,
	}, nil
}

// OpenRepo opens an existing repo
func OpenRepo(filename string) (Repo, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %v", err)
	}
	zr, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening repo: %v", err)
	}
	return &readrepo{
		fileReader: f,
		zipReader:  zr,
	}, nil
}

type repo struct {
	fileWriter io.WriteCloser
	zipWriter  *zip.Writer
	counter    int
	mu         sync.Mutex
}

func (r *repo) Close() {
	r.zipWriter.Close()
	r.fileWriter.Close()
}

func (r *repo) Create() (io.Writer, error) {
	var counter int
	r.mu.Lock()
	counter = r.counter
	r.counter++
	r.mu.Unlock()

	filename := fmt.Sprintf("file%d", counter)
	w, err := r.zipWriter.Create(filename)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *repo) CreateAndWrite(v interface{}) error {
	w, err := r.Create()
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(w)
	err = enc.Encode(v)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) Shards() int {
	return 0
}

func (r *repo) Open(int) (io.ReadCloser, error) {
	return nil, fmt.Errorf("cannot open a repo being written to")
}

func (r *repo) Read(int, interface{}) error {
	return fmt.Errorf("cannot read from a repo being written to")
}

type readrepo struct {
	fileReader io.ReadCloser
	zipReader  *zip.ReadCloser
}

func (r *readrepo) Close() {
	r.zipReader.Close()
	r.fileReader.Close()
}

func (r *readrepo) Create() (io.Writer, error) {
	return nil, fmt.Errorf("cannot create in a readonly repo")
}

func (r *readrepo) CreateAndWrite(v interface{}) error {
	return fmt.Errorf("cannot create & write in a readonly repo")
}

func (r *readrepo) Shards() int {
	return len(r.zipReader.File)
}

func (r *readrepo) Open(idx int) (io.ReadCloser, error) {
	if idx < 0 || idx > r.Shards() {
		return nil, fmt.Errorf("no shard at requested index: %d", idx)
	}
	return r.zipReader.File[idx].Open()
}

func (r *readrepo) Read(idx int, v interface{}) error {
	shard, err := r.Open(idx)
	if err != nil {
		return err
	}
	defer shard.Close()

	dec := gob.NewDecoder(shard)
	err = dec.Decode(v)
	if err != nil {
		return err
	}
	return nil
}
