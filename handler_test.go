package service

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

// TestContentTypeNoSniff - refer to https://www.owasp.org/index.php/OWASP_Secure_Headers_Project#tab=Headers - X-Content-Type-Options
func TestContentTypeNoSniff(t *testing.T) {
	h := getGenericHandler()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := SecureGivenHandler(h)
	handler.ServeHTTP(w, req)

	headers := w.Header()
	nosniff := headers.Get("X-Content-Type-Options")
	if nosniff != "nosniff" {
		t.Error("X-Content-Type-Options not set or set to wrong value: ", nosniff)
	}
}

func getGenericHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestContentTypeNoSniffWithMuxRouter(t *testing.T) {
	h := mux.NewRouter()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := SecureGivenHandler(h)
	handler.ServeHTTP(w, req)

	headers := w.Header()
	nosniff := headers.Get("X-Content-Type-Options")
	if nosniff != "nosniff" {
		t.Error("X-Content-Type-Options not set or set to wrong value: ", nosniff)
	}
}

// TestDenyXFrame - refer to X-Frame-Options - going with DENY instead of deny. Not sure which is right.
func TestDenyXFrame(t *testing.T) {
	h := getGenericHandler()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := SecureGivenHandler(h)
	handler.ServeHTTP(w, req)

	headers := w.Header()
	deny := headers.Get("X-Frame-Options")
	if deny != "DENY" {
		t.Error("X-Frame-Options not set or set to wrong value: ", deny)
	}
}

// TestXSSProtection - refer to X-XSS-Protection
func TestXSSProtection(t *testing.T) {
	h := getGenericHandler()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := SecureGivenHandler(h)
	handler.ServeHTTP(w, req)

	headers := w.Header()
	protection := headers.Get("X-XSS-Protection")
	if protection != "1; mode=block" {
		t.Error("X-XSS-Protection not set or set to wrong value: ", protection)
	}
}

// TestAllowCORSForDevTesting - refer to https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
func TestAllowCORSForDevTesting(t *testing.T) {
	h := getGenericHandler()
	w := httptest.NewRecorder()

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Add("Origin", "iapps365.com")
	req.Header.Add("Access-Control-Request-Method", "POST")

	handler := AllowCORSForDevTesting(h)
	handler.ServeHTTP(w, req)

	headers := w.Header()
	allowedOrigin := headers.Get("Access-Control-Allow-Origin")
	if allowedOrigin != "*" {
		t.Error("Access-Control-Allow-Origin not set or set to wrong value during OPTIONS (expected *): ", allowedOrigin)
	}
	allowedMethods := headers.Get("Access-Control-Allow-Methods")
	if allowedMethods != "POST" {
		t.Error("Access-Control-Allow-Methods not set or set to wrong value (expected POST): ", allowedMethods)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/", nil)
	req.Header.Add("Origin", "iapps365.com")
	handler.ServeHTTP(w, req)
	headers = w.Header()
	allowedOrigin = headers.Get("Access-Control-Allow-Origin")
	if allowedOrigin != "*" {
		t.Error("Access-Control-Allow-Origin not set or set to wrong value during POST (expected *): ", allowedOrigin)
	}
}

func TestAllowCORSForSpecificOrigin(t *testing.T) {
	h := getGenericHandler()
	w := httptest.NewRecorder()

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Add("Origin", "iapps365.com")
	req.Header.Add("Access-Control-Request-Method", "POST")

	handler := AllowCORSForSpecificOrigins(h, []string{"google.com", "iapps365.com"})
	handler.ServeHTTP(w, req)

	headers := w.Header()
	allowedOrigin := headers.Get("Access-Control-Allow-Origin")
	if allowedOrigin != "iapps365.com" {
		t.Error("Access-Control-Allow-Origin not set or set to wrong value (expected iapps365.com): ", allowedOrigin)
	}
}

func TestCORSForOriginShouldDenyOtherOrigins(t *testing.T) {
	h := getGenericHandler()
	w := httptest.NewRecorder()

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Add("Origin", "iapps365X.com")
	req.Header.Add("Access-Control-Request-Method", "POST")

	handler := AllowCORSForSpecificOrigins(h, []string{"google.com", "iapps365.com"})
	handler.ServeHTTP(w, req)

	headers := w.Header()
	allowedOrigin := headers.Get("Access-Control-Allow-Origin")
	if allowedOrigin != "" {
		t.Errorf("Access-Control-Allow-Origin is set to %v. It shouldn't be.", allowedOrigin)
	}
}

func TestFileUploadHandler(t *testing.T) {

	t.Run("Happy Path", GoodFileUploadHandler)

}

func GoodFileUploadHandler(t *testing.T) {
	data := `this is test content of the file that will be uploaded.`
	filename := "test.txt"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Error("Error in TestFileUploadHandler: ", err)
		return
	}
	io.Copy(part, strings.NewReader(data))
	writer.Close()

	// resp, err := ioutil.ReadAll(body)
	// if err != nil {
	// 	t.Error("Error in TestFileUploadHandler: ", err)
	// 	return
	// }
	// fmt.Println(string(resp))

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	spec := UploadHandlerSpec{
		Param:          "file",
		UploadLocation: "/tmp/",
		DownloadURL:    "http://abc.com/uploads/",
	}
	handler := GetUploadHandler(spec)

	handler(w, req)

	resp, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error("Error in TestFileUploadHandler: ", err)
		return
	}
	if w.Result().StatusCode != 200 {
		t.Error("Did not get 200 error code. Instead got: ", w.Result().StatusCode)
	}
	respStr := string(resp)
	if !strings.HasPrefix(respStr, spec.DownloadURL) {
		t.Error("Did not get the right prefix to download URL: ", respStr)
	}
	if !strings.HasSuffix(respStr, filename) {
		t.Error("Did not get the right suffix to download URL: ", respStr)
	}
}
