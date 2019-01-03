package service

import (
	"bytes"
	"encoding/csv"
	"io/ioutil"
	"strings"
	"testing"
)

func TestCSVParser(t *testing.T) {
	data := `first_name,last_name,username
"Rob","Pike",rob
bad record,test
Ken,Thompson,ken
second bad record,""test""
Arun,"Barua",abarua
`

	spec := CSVSpec{}

	t.Run("Header", func(t *testing.T) {
		expected := []string{"first_name", "last_name", "username"}
		headerFound := false
		r := strings.NewReader(data)
		ParseCSV(r, spec, func(record []string, header bool) {
			if !header {
				return
			}
			headerFound = true
			if len(record) != len(expected) {
				t.Errorf("Header length not as expected: %d instead of %d.", len(record), len(expected))
				return
			}
			for i, v := range expected {
				if record[i] != v {
					t.Errorf("Unexpected value in header for column %d. %s instead of %s.", i, record[i], v)
					return
				}
			}
		}, func(row int, e error) {})
		if !headerFound {
			t.Error("Header not found.")
		}
	})

	t.Run("Rows", func(t *testing.T) {
		expected := [][]string{
			{"Rob", "Pike", "rob"},
			{"Ken", "Thompson", "ken"},
			{"Arun", "Barua", "abarua"},
		}
		found := make([]bool, len(expected))
		counter := 0
		r := strings.NewReader(data)
		ParseCSV(r, spec, func(record []string, header bool) {
			if header {
				return
			}
			if counter == len(expected) {
				t.Error("Found more rows than expected!")
				return
			}
			exp := expected[counter]
			if len(exp) != len(record) {
				t.Errorf("For row %d expected: %d, got %d.", counter, len(exp), len(record))
				return
			}
			for i, v := range exp {
				if record[i] != v {
					t.Errorf("For row %d and column %d, expected: %s instead of %s.", counter+1, i+1, v, record[i])
				}
			}
			found[counter] = true
			counter++
		}, func(row int, e error) {})
		for i, v := range found {
			if !v {
				t.Errorf("Did not find row %d.", i)
			}
		}
	})

	t.Run("Errors", func(t *testing.T) {
		r := strings.NewReader(data)
		errorFound := []bool{false, false}
		ParseCSV(r, spec, func(record []string, header bool) {}, func(row int, e error) {
			parseError, ok := e.(*csv.ParseError)
			if !ok {
				t.Errorf("Expected parseError, instead found: %T for row: %d", e, row)
				return
			}
			if row == 1 {
				if parseError.Err != csv.ErrFieldCount {
					t.Error("Expected ErrFieldCount instead got: ", parseError.Err)
				}
				errorFound[0] = true
			}
			if row == 3 {
				if parseError.Err != csv.ErrQuote {
					t.Error("Expected ErrQuote instead got: ", parseError.Err)
				}
				errorFound[1] = true
			}
		})
		for i, e := range errorFound {
			if e {
				continue
			}
			t.Errorf("Error %d expected but not found.", i+1)
		}
	})

	t.Run("Delimiter", func(t *testing.T) {
		data := `A|B|C
D|E|F`
		expected := [][]string{
			{"A", "B", "C"},
			{"D", "E", "F"},
		}
		found := make([]bool, len(expected))
		counter := 0
		r := strings.NewReader(data)
		spec := CSVSpec{Comma: '|'}
		ParseCSV(r, spec, func(record []string, header bool) {
			if counter == len(expected) {
				t.Error("Found more rows than expected!")
				return
			}
			exp := expected[counter]
			if len(exp) != len(record) {
				t.Errorf("For row %d expected: %d, got %d.", counter, len(exp), len(record))
				return
			}
			for i, v := range exp {
				if record[i] != v {
					t.Errorf("For row %d and column %d, expected: %s instead of %s.", counter+1, i+1, v, record[i])
				}
			}
			found[counter] = true
			counter++
		}, func(row int, e error) {})
		for i, v := range found {
			if !v {
				t.Errorf("Did not find row %d.", i)
			}
		}
	})
}

func TestLineParser(t *testing.T) {
	data := `line 1
line 2
line 3`

	t.Run("Rows As Strings", func(t *testing.T) {
		r := strings.NewReader(data)

		expected := []string{
			"line 1",
			"line 2",
			"line 3",
		}
		found := make([]bool, len(expected))
		counter := 0
		ParseLinesAsStrings(r, func(line string) {
			if counter == len(expected) {
				t.Error("Found an un-expected line!", line)
				return
			}
			if line != expected[counter] {
				t.Errorf("Expected %s, found %s.", expected[counter], line)
			}
			found[counter] = true
			counter++
		})
		for i, f := range found {
			if !f {
				t.Errorf("Line %d not found.", i+1)
			}
		}
	})

	t.Run("Rows As Byte Arrays", func(t *testing.T) {
		r := strings.NewReader(data)

		expected := [][]byte{
			[]byte("line 1"),
			[]byte("line 2"),
			[]byte("line 3"),
		}
		found := make([]bool, len(expected))
		counter := 0
		ParseLinesAsBytes(r, func(line []byte) {
			if counter == len(expected) {
				t.Error("Found an expected line!", line)
				return
			}
			if bytes.Compare(expected[counter], line) != 0 {
				t.Errorf("Expected %v, found %v.", expected[counter], line)
			}
			found[counter] = true
			counter++
		})
		for i, f := range found {
			if !f {
				t.Errorf("Line %d not found.", i+1)
			}
		}
	})

}

func TestFixCROnlyNewLine(t *testing.T) {

	t.Run("Simple Happy Path", func(t *testing.T) {
		data := "this is a test.\rAnother line here.\r"
		expected := "this is a test.\r\nAnother line here.\r\n"

		r := strings.NewReader(data)
		newR := CRNewLineFixer(r)

		result, err := ioutil.ReadAll(newR)
		if err != nil {
			t.Error("Encountered error: ", err)
			return
		}

		if string(result) != expected {
			t.Errorf("Expected %s, got %s.", expected, string(result))
		}
	})

	t.Run("Avoid Doubles", func(t *testing.T) {
		data := "this is a test.\r\nAnother line here.\r\n"
		expected := "this is a test.\r\nAnother line here.\r\n"

		r := strings.NewReader(data)
		newR := CRNewLineFixer(r)

		result, err := ioutil.ReadAll(newR)
		if err != nil {
			t.Error("Encountered error: ", err)
			return
		}

		if string(result) != expected {
			t.Errorf("Expected %s, got %s.", expected, string(result))
		}
	})

	t.Run("All CR", func(t *testing.T) {
		data := "\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r\r"
		data = data + data + data + data + data + data + data + data
		data = data + data + data + data + data + data + data + data
		data = data + data + data + data + data + data + data + data
		data = data + data + data + data + data + data + data + data
		data = data + data + data + data + data + data + data + data
		expected := "\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n"
		expected = expected + expected + expected + expected + expected + expected + expected + expected
		expected = expected + expected + expected + expected + expected + expected + expected + expected
		expected = expected + expected + expected + expected + expected + expected + expected + expected
		expected = expected + expected + expected + expected + expected + expected + expected + expected
		expected = expected + expected + expected + expected + expected + expected + expected + expected

		r := strings.NewReader(data)
		newR := CRNewLineFixer(r)

		result, err := ioutil.ReadAll(newR)
		if err != nil {
			t.Error("Encountered error: ", err)
			return
		}

		if string(result) != expected {
			t.Error("Strings did not match.")
		}
	})
}
