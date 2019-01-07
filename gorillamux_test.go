/*
This tests the github.com/gorilla/mux package more for documentation than anything else.

Serve handler as:
	srv := http.Server{
		Addr:         ":8087",
		Handler:      handler,
		ReadTimeout:  time.Minute * 3,
		WriteTimeout: time.Minute * 3,
	}
	log.Fatal(srv.ListenAndServe())
*/
package service

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

func TestNewRouterAsHTTPHandler(t *testing.T) {
	mux := mux.NewRouter()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FORM PAGE"))
	})

	t.Run("GET Root for 200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)
		reply, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Error("Error reading from response:", err)
			return
		}
		if w.Code != 200 {
			t.Error("Expected 200 OK. Got:", w.Code)
			return
		}
		if string(reply) != "OK" {
			t.Error("Expected OK but found:", string(reply))
		}
	})

	t.Run("Test 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/doesnotexist", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)
		if w.Code != 404 {
			t.Error("Expected 404. Got:", w.Code)
		}
	})

	t.Run("GET Form Page", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/form?a=b", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)
		reply, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Error("Error reading from response:", err)
			return
		}
		if w.Code != 200 {
			t.Error("Expected 200 OK. Got:", w.Code)
			return
		}
		if string(reply) != "FORM PAGE" {
			t.Error("Expected FORM PAGE but found:", string(reply))
		}
	})

}

func TestPathPrefixAndSubRouterForAPI(t *testing.T) {
	mux := mux.NewRouter()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	api := mux.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GOT USERS"))
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)
	reply, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error("Error reading from response:", err)
		return
	}
	if w.Code != 200 {
		t.Error("Expected 200 OK. Got:", w.Code)
		return
	}
	if string(reply) != "GOT USERS" {
		t.Error("Expected GOT USERS but found:", string(reply))
	}
}

func TestStaticFileServing(t *testing.T) {
	mux := mux.NewRouter()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	mux.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("."))))

	req := httptest.NewRequest("GET", "/files/gorillamux_test.go", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)
	reply, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error("Error reading from response:", err)
		return
	}
	if w.Code != 200 {
		t.Error("Expected 200 OK. Got:", w.Code)
		return
	}
	f, err := os.Open("gorillamux_test.go")
	if err != nil {
		t.Error("Did not find file: gorillamux_test.go.")
		return
	}
	defer f.Close()
	expected, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error("Could not read: gorillamux_test.go.")
		return
	}
	if string(reply) != string(expected) {
		t.Error("File contents did not match.")
	}
}
