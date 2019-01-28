package service

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

// UploadClientSpec specifies the requirements for the UploadClient
type UploadClientSpec struct {
	URL      string
	Content  io.Reader
	Filename string
}

// FileUploaderClient uploads file per the spec
func FileUploaderClient(spec UploadClientSpec) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", spec.Filename)
	if err != nil {
		return nil, err
	}
	io.Copy(part, spec.Content)
	writer.Close()

	resp, err := http.Post(spec.URL, writer.FormDataContentType(), body)
	if err != nil {
		return nil, err
	}
	respBody := resp.Body
	defer respBody.Close()
	respContent, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, err
	}
	return respContent, nil
}

// HTMLParser is a callback function once the client gets a successful response
type HTMLParser func(io.Reader)

// HTMLParserClientSpec is the spec that defines the HTMLParserClient
type HTMLParserClientSpec struct {
	URL    string
	Parser HTMLParser
}

// HTMLParserClient does a GET on the spec URL and calls the included Parser function
func HTMLParserClient(spec HTMLParserClientSpec) error {
	resp, err := http.Get(spec.URL)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected 200 status code. Got: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	spec.Parser(resp.Body)
	return nil
}
