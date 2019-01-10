package service

import (
	"bufio"
	"encoding/csv"
	"errors"
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

// LineProcessorWithAbort is the function called for each line that supports returning an Abort error
type LineProcessorWithAbort func(line string) error

// ParseLinesAsStringsWithAbort parses line from a reader with option to abort
func ParseLinesAsStringsWithAbort(r io.Reader, processor LineProcessorWithAbort) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := processor(scanner.Text()); err != nil {
			return err
		}
	}
	return nil
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

// WriteCSV writes to a CSV writer the records
func WriteCSV(spec CSVSpec, records [][]string, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	if spec.Comma != 0 {
		csvWriter.Comma = spec.Comma
	}
	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return errors.New("error writing record to csv: " + err.Error())
		}
	}
	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return err
	}
	return nil
}
