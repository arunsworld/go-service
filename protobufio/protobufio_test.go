package protobufio

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/arunsworld/go-service/protobufio/testapi"
)

func TestWriteProtobufsToZipFile(t *testing.T) {
	var zipFile io.Writer
	zipFile = &bytes.Buffer{}

	zw, err := NewZipWriter(zipFile)
	if err != nil {
		t.Fatal(err)
	}

	tt, _ := time.Parse("Mon Jan 2 15:04:05 MST 2006", "Wed Jan 9 13:17:22 IST 2019")

	msg := &testapi.Test{ID: 1, Date: MustConvertTimestamp(tt), Message: "hello there"}
	if err := zw.Write(msg); err != nil {
		t.Fatal(err)
	}

	msg = &testapi.Test{ID: 2, Date: MustConvertTimestamp(tt), Message: "bye now"}
	if err := zw.Write(msg); err != nil {
		t.Fatal(err)
	}

	zw.Close()

	checksum := fmt.Sprintf("%x", md5.Sum(zipFile.(*bytes.Buffer).Bytes()))
	expectedChecksum := "af049987915eb4391091d6e5640afa8d"
	if checksum != expectedChecksum {
		t.Fatalf("Expected checksum: %s, got: %s", expectedChecksum, checksum)
	}
}

func generateHugeProtoZipFile() ([]byte, error) {
	zipFile := &bytes.Buffer{}
	zw, err := NewZipWriter(zipFile)
	if err != nil {
		return nil, err
	}

	tt, _ := time.Parse("Mon Jan 2 15:04:05 MST 2006", "Wed Jan 9 13:17:22 IST 2019")
	ttt := MustConvertTimestamp(tt)

	for i := 0; i < 100000; i++ {
		msg := &testapi.Test{ID: 1, Date: ttt, Message: "hello there"}
		if err := zw.Write(msg); err != nil {
			return nil, err
		}
	}

	zw.Close()

	return zipFile.Bytes(), nil
}

func TestReadProtobufsToZipFile(t *testing.T) {
	data, err := generateHugeProtoZipFile()
	if err != nil {
		t.Fatal(err)
	}
	tmpfile, err := ioutil.TempFile("", "protobuf.zip")
	if err != nil {
		t.Fatal("Unable to create temp file:", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(data); err != nil {
		t.Fatal("Unable to write to temp file:", err)
	}
	tmpfile.Close()

	zr, err := NewZipReader(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	msg := &testapi.Test{}
	if err := zr.Read(msg); err != nil {
		t.Fatal(err)
	}
	if msg.ID != 1 {
		t.Fatal("Expected message ID = 1 but got:", msg.ID)
	}

	counter := 0
	for {
		err := zr.Read(msg)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		counter++
	}
	if counter != 99999 {
		t.Fatal("Expecting to get 99999 messages but got:", counter)
	}
}
