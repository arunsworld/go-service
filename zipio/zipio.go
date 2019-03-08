package zipio

import (
	"archive/zip"
	"io"
	"os"
)

type zipWriter struct {
	file io.WriteCloser
	zw   *zip.Writer
	w    io.Writer
}

func (z *zipWriter) Write(p []byte) (n int, err error) {
	return z.w.Write(p)
}

func (z *zipWriter) Close() error {
	z.zw.Close()
	z.file.Close()
	return nil
}

// Create creates a new single file zip writer
func Create(filename string) (io.WriteCloser, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("file")
	if err != nil {
		zw.Close()
		f.Close()
		return nil, err
	}
	return &zipWriter{
		file: f,
		zw:   zw,
		w:    w,
	}, nil
}

type zipReader struct {
	zr *zip.ReadCloser
	r  io.ReadCloser
}

func (z *zipReader) Read(p []byte) (n int, err error) {
	return z.r.Read(p)
}

func (z *zipReader) Close() error {
	z.r.Close()
	z.zr.Close()
	return nil
}

// Open opens a new single file zip for reading
func Open(filename string) (io.ReadCloser, error) {
	zr, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	r, err := zr.File[0].Open()
	if err != nil {
		zr.Close()
		return nil, err
	}
	return &zipReader{
		zr: zr,
		r:  r,
	}, nil
}
