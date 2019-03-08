package zipio

import (
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestZipWriter(t *testing.T) {
	var err error
	info, err := os.Stat("testfiles/writeTest.zip")
	if info != nil {
		if err := os.Remove("testfiles/writeTest.zip"); err != nil {
			t.Fatal(err)
		}
	}

	// This is the business logic
	var w io.WriteCloser
	w, err = Create("testfiles/writeTest.zip")
	if err != nil {
		t.Fatal(err)
	}
	enc := gob.NewEncoder(w)
	enc.Encode([...]string{"arun", "barua", "likes", "golang"})
	w.Close()

	// These are the assertions
	info, err = os.Stat("testfiles/writeTest.zip")
	if err != nil {
		t.Fatal("Expected test.zip file but did not find it...")
	}

	f, err := os.Open("testfiles/writeTest.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	contents, err := ioutil.ReadAll(f)
	checkSum := fmt.Sprintf("%x", md5.Sum(contents))
	if checkSum != "ddf2fadc8153c95390e8e6526dbcc8ea" {
		t.Fatalf("Checksum didn't match. Expecting %s. Got %s.", "ddf2fadc8153c95390e8e6526dbcc8ea", checkSum)
	}

}

func TestZipReader(t *testing.T) {
	var r io.ReadCloser
	var err error
	r, err = Open("testfiles/readTest.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	var content []string
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&content); err != nil {
		t.Fatal(err)
	}
	expected := [...]string{"arun", "barua", "likes", "golang"}
	if len(expected) != len(content) {
		t.Fatalf("Expected and received content length didn't match. Got: %d vs %d.", len(content), len(expected))
	}
	for i := 0; i < len(expected); i++ {
		if content[i] != expected[i] {
			t.Fatalf("Value in position %d did not match. Got: %s vs %s", i+1, content[i], expected[i])
		}
	}
}

func BenchmarkZipWriter(b *testing.B) {
	content := []string{}
	for i := 0; i < 1000000; i++ {
		content = append(content, fmt.Sprintf("hello %d", i))
	}
	for n := 0; n < b.N; n++ {
		w, err := Create("testfiles/writePerfTest.zip")
		if err != nil {
			b.Fatal(err)
		}
		enc := gob.NewEncoder(w)
		enc.Encode(content)
		w.Close()
	}
}

func BenchmarkZipReader(b *testing.B) {
	for n := 0; n < b.N; n++ {
		r, err := Open("testfiles/readPerfTest.zip")
		if err != nil {
			b.Fatal(err)
		}
		var content []string
		decoder := gob.NewDecoder(r)
		if err := decoder.Decode(&content); err != nil {
			b.Fatal(err)
		}
		r.Close()
	}
}
