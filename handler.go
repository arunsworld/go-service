package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gofrs/uuid"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

// SecureGivenHandler secures the given handler and returns it
func SecureGivenHandler(h http.Handler) http.Handler {
	secureMiddleware := secure.New(secure.Options{
		ContentTypeNosniff: true,
		FrameDeny:          true,
		BrowserXssFilter:   true,
	})

	h = secureMiddleware.Handler(h)
	return h
}

// AllowCORSForDevTesting modifies and returns the handler to set the headers to allow CORS for API calls etc.
func AllowCORSForDevTesting(h http.Handler) http.Handler {
	c := cors.New(cors.Options{})
	return c.Handler(h)
}

// AllowCORSForSpecificOrigins modifies and returns the handler to set the headers to allow CORS for specific origins
func AllowCORSForSpecificOrigins(h http.Handler, origins []string) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: origins,
	})
	return c.Handler(h)
}

// UploadHandlerSpec captures the specification for the upload
type UploadHandlerSpec struct {
	Param          string
	UploadLocation string
	DownloadURL    string
}

// GetUploadHandler gets an upload handler based on the spec
func GetUploadHandler(spec UploadHandlerSpec) http.HandlerFunc {
	if spec.Param == "" {
		spec.Param = "file"
	}
	if spec.UploadLocation == "" {
		spec.UploadLocation = "/tmp/"
	}
	if spec.DownloadURL == "" {
		spec.DownloadURL = "http://localhost/uploads/"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		f, header, err := r.FormFile(spec.Param)
		if err != nil {
			log.Println("Bad request being sent to UploadHandler. Error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		u, err := uuid.NewV4()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newFilename := u.String() + header.Filename
		targetFile, err := os.Create(spec.UploadLocation + newFilename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer targetFile.Close()
		io.Copy(targetFile, f)
		fmt.Fprint(w, spec.DownloadURL+newFilename)
	}
}
