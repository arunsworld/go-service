package main

import (
	"fmt"
	"log"
	"os"

	service "github.com/arunsworld/go-service"
)

func main() {
	f, err := os.Open("/Users/abarua/Downloads/SalesJan2009.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	spec := service.UploadClientSpec{
		URL: "http://localhost:8087/upload",
		// Content:  strings.NewReader("this is a test of text!!"),
		Content:  f,
		Filename: "abc.txt",
	}
	resp, err := service.FileUploaderClient(spec)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(resp))
}
