package service

import (
	"bufio"
	"encoding/csv"
	"io"
)

// RecordProcessor is the function which is called for every record in the CSV
type RecordProcessor func(record []string, header bool)

// ErrorRecordProcessor is the function which is called for an error row during CSV processing
type ErrorRecordProcessor func(row int, e error)

// CSVSpec specifies the CSV parser
type CSVSpec struct {
	Comma rune
}

// ParseCSV parses the reader as CSV and calls the RecordProcessor for each record
func ParseCSV(r io.Reader, spec CSVSpec, processor RecordProcessor, errorProcessor ErrorRecordProcessor) {
	csvReader := csv.NewReader(r)
	if spec.Comma != 0 {
		csvReader.Comma = spec.Comma
	}
	header := true
	rowCounter := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if !header {
			rowCounter++
		}
		if err != nil {
			errorProcessor(rowCounter-1, err)
			continue
		}
		processor(record, header)
		header = false
	}
}

// LineProcessor is the function called for each line while Parsing lines
type LineProcessor func(line string)

// ParseLinesAsStrings parses lines from a reader
func ParseLinesAsStrings(r io.Reader, processor LineProcessor) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		processor(scanner.Text())
	}
}

// BinaryLineProcessor is the function called for each line while Parsing lines
type BinaryLineProcessor func(line []byte)

// ParseLinesAsBytes parses lines from a reader
func ParseLinesAsBytes(r io.Reader, processor BinaryLineProcessor) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		processor(scanner.Bytes())
	}
}

// CRFixer fixes CR to \r\n
type CRFixer struct {
	r io.Reader
	q []byte
}

func (crf *CRFixer) Read(p []byte) (int, error) {
	if len(crf.q) < len(p)/2 {
		crf.q = make([]byte, len(p)/2)
	}
	n, err := crf.r.Read(crf.q)
	counter := 0
	ignoreNL := false
	for i := 0; i < n; i++ {
		ch := crf.q[i]
		if ignoreNL && ch == '\n' {
			ignoreNL = false
			continue
		}
		p[counter] = ch
		if ch == '\r' {
			counter++
			p[counter] = '\n'
			ignoreNL = true
		}
		counter++
	}
	return counter, err
}

// CRNewLineFixer converts \r to \n
func CRNewLineFixer(in io.Reader) *CRFixer {
	return &CRFixer{r: in}
}
