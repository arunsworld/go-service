package service

import (
	"bufio"
	"fmt"
	"io"
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
		t.Fatal("Error during file upload: ", err)
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

func TestHTMLParserClient(t *testing.T) {
	response := `<html>
	<head></head>
	<body>
		<p>Hello, World</p>
	</body>
</html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()

	counter := 0
	var scannerErr error

	spec := HTMLParserClientSpec{
		URL: ts.URL,
		Parser: func(r io.Reader) {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				counter++
			}
			if err := scanner.Err(); err != nil {
				scannerErr = err
			}
		},
	}
	err := HTMLParserClient(spec)
	if err != nil {
		t.Fatal(err)
	}
	if scannerErr != nil {
		t.Fatal(scannerErr)
	}
	if counter != 6 {
		t.Fatalf("Expected 6 lines instead got %d lines.", counter)
	}
}
