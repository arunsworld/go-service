package protobufio

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// MustConvertTimestamp converts given time otherwise panics
func MustConvertTimestamp(t time.Time) *timestamp.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	return ts
}

// ConvertTimestamp converts given time into proto timestamp
func ConvertTimestamp(t time.Time) (*timestamp.Timestamp, error) {
	return ptypes.TimestampProto(t)
}

// ZipWriter holds state to write protobuf to a zip file
type ZipWriter struct {
	zipWriter *zip.Writer
	w         io.Writer
}

// Write writes proto.Message to writer
func (z *ZipWriter) Write(msg proto.Message) error {
	size := uint32(proto.Size(msg))
	err := binary.Write(z.w, binary.LittleEndian, size)
	if err != nil {
		return fmt.Errorf("failed to write size to zip file: %v", err)
	}
	content, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal proto message: %v", err)
	}
	_, err = z.w.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write proto message to zip file: %v", err)
	}
	return nil
}

// Close closes out the writer stream
func (z *ZipWriter) Close() error {
	return z.zipWriter.Close()
}

// NewZipWriter creates a new ZipWriter
func NewZipWriter(w io.Writer) (*ZipWriter, error) {
	zipWriter := zip.NewWriter(w)
	f, err := zipWriter.Create("protobuf.db")
	if err != nil {
		return nil, err
	}
	return &ZipWriter{zipWriter: zipWriter, w: f}, nil
}

// ZipReader holds state for reading from zip
type ZipReader struct {
	zipReader *zip.ReadCloser
	r         io.ReadCloser
}

func (z *ZipReader) Read(msg proto.Message) error {
	var size uint32
	err := binary.Read(z.r, binary.LittleEndian, &size)
	if err != nil {
		return err
	}
	content := make([]byte, size)
	n, err := z.r.Read(content)
	if err != nil && n < int(size) {
		return err
	}
	if n < int(size) {
		balance := int(size) - n
		leftoverContent := make([]byte, balance)
		nn, err := z.r.Read(leftoverContent)
		if err != nil {
			return err
		}
		if nn < balance {
			return fmt.Errorf("could not read the required bytes despite 2 attempts")
		}
		copy(content[n:], leftoverContent)
	}
	err = proto.Unmarshal(content, msg)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the open handlers for the ZipReader
func (z *ZipReader) Close() error {
	err1 := z.r.Close()
	err2 := z.zipReader.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// NewZipReader creates a new ZipReader
func NewZipReader(filename string) (*ZipReader, error) {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}

	f, err := zipReader.File[0].Open()
	if err != nil {
		zipReader.Close()
		return nil, err
	}

	return &ZipReader{zipReader: zipReader, r: f}, nil
}
