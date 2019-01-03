package service

import (
	"bytes"
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
