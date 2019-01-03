package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	service "github.com/arunsworld/go-service"
)

func main() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	})

	mux := mux.NewRouter()
	mux.HandleFunc("/abcd", h)
	mux.HandleFunc("/upload", service.GetUploadHandler(service.UploadHandlerSpec{
		DownloadURL: "http://localhost:8087/uploads/",
	}))
	mux.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("/tmp/"))))

	handler := service.SecureGivenHandler(mux)
	handler = service.AllowCORSForDevTesting(handler)

	srv := http.Server{
		Addr:         ":8087",
		Handler:      handler,
		ReadTimeout:  time.Minute * 3,
		WriteTimeout: time.Minute * 3,
	}
	log.Fatal(srv.ListenAndServe())
}
