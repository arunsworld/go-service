package binaryio

import (
	"testing"
)

func TestCreateRepo(t *testing.T) {
	repo, err := NewRepo("repo.db")
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	err = repo.CreateAndWrite([]string{"Name 1", "Name 2", "Name 3", "Name 4", "Name 5", "Name 6"})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.CreateAndWrite([]string{"Name 7", "Name 8", "Name 9", "Name 10", "Name 11", "Name 12"})
	if err != nil {
		t.Fatal(err)
	}

	data := make([]int, 1000000)
	for i := 0; i < 1000000; i++ {
		data[i] = i
	}
	err = repo.CreateAndWrite(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRepo(t *testing.T) {
	repo, err := OpenRepo("repo-readonly.db")
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	if repo.Shards() != 3 {
		t.Fatal("Expected 3 shards but got:", repo.Shards())
	}
	var v []string
	err = repo.Read(0, &v)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 6 {
		t.Fatal("Expected 6 entries but got:", len(v))
	}
}
