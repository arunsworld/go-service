package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	service "github.com/arunsworld/go-service"
)

func main() {
	webDownload()
}

func webDownload() {
	resp, err := http.Get("http://samplecsvs.s3.amazonaws.com/TechCrunchcontinentalUSA.csv")
	if err != nil {
		log.Fatal(err)
	}
	r := resp.Body
	defer r.Close()
	service.ParseCSV(service.CRNewLineFixer(r), service.CSVSpec{}, func(record []string, header bool) {
		fmt.Println(strings.Join(record, "\t"))
	}, func(row int, err error) {
		log.Printf("ERROR in row %d: %v.\n", row, err)
	})
}
