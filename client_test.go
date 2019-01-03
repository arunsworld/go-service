package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFileUploaderClient(t *testing.T) {
	var bodyContentsStr string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Body
		defer body.Close()

		bodyContents, err := ioutil.ReadAll(body)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bodyContentsStr = string(bodyContents)

		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	data := `this is test content of the file that will be uploaded.`

	spec := UploadClientSpec{
		URL:      ts.URL,
		Content:  strings.NewReader(data),
		Filename: "test.txt",
	}
	resp, err := FileUploaderClient(spec)
	if err != nil {
		t.Error("Error during file upload: ", err)
	}

	if !strings.Contains(bodyContentsStr, "Content-Disposition: form-data; name=\"file\";") {
		t.Error("Did not find: Content-Disposition: form-data; name=\"file\";")
	}
	if !strings.Contains(bodyContentsStr, "filename=\"test.txt\"") {
		t.Error("Did not find: filename=\"test.txt\"")
	}
	if !strings.Contains(bodyContentsStr, data) {
		t.Error("Did not find uploaded file data.")
	}
	if string(resp) != "OK" {
		t.Error("Did not find the right response.")
	}

}
