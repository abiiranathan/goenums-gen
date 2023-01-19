package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Create a regular expression to match "CREATE TYPE" statements for ENUM types.
var re = regexp.MustCompile(`(?i)CREATE TYPE (\w+) AS ENUM\s*\(([\w',\s]+)\);`)

// Create an English caser to transform constants to TitleCase.
// strings.Title is deprecated.
var enCaser = cases.Title(language.English)

type TemplateData struct {
	PkgName   string   // Name of the package.
	TypeName  string   // The Enum type in generated go code.
	Values    []string // Enum options
	FirstType bool     // Whether to write imports and package declaration
}

// Passes data to the string template that is executed into w.
func ParseTemplate(w io.Writer, data TemplateData) {
	tmpl, err := template.New("tmpl").Parse(templateString)
	if err != nil {
		log.Fatalf("can not parse template: %v\n", err)
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf("error executing template: %v\n", err)
	}
}

// Converts items to CamelCase, removes whitespace.
func TransforEnums(items []string) []string {
	for index := range items {
		items[index] = strcase.ToCamel(enCaser.String(strings.TrimSpace(items[index])))
	}
	return items
}

// Rough error handling. Panics if err != nil with context provided by msg.
func ExitOnError(msg string, err error) {
	if err != nil {
		log.Panicf("%s: %v\n", msg, err)
	}
}

var (
	pkgName        string
	inputFileName  string
	outputFileName string
)

func init() {
	flag.StringVar(&pkgName, "pkg", "", "The name of the package for the generated file")
	flag.StringVar(&inputFileName, "in", "", "The absolute path to input sql file")
	flag.StringVar(&outputFileName, "out", "", "The absolute path to output file")
}

// asserts that all flags are provided. Otherwise prints Usage and exists the program
// with exit code 1
func validateFlags() {
	if pkgName == "" || inputFileName == "" || outputFileName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	flag.Parse()

	validateFlags()

	// get abspath of input file.
	input, err := filepath.Abs(inputFileName)
	ExitOnError("Abs() error on input file", err)

	// Open input file.
	f, err := os.Open(input)
	ExitOnError("can not open input file", err)
	defer f.Close()

	// Create output file.
	output, err := os.Create(outputFileName)
	ExitOnError("can not create output file", err)
	defer output.Close()

	// Create a buffered reader for input file.
	reader := bufio.NewReader(f)

	// Create a buffer to hold the contents of CREATE TYPE statements.
	// After each statement, it's reset.
	buffer := bytes.Buffer{}
	outBuffer := bytes.Buffer{}

	count := 0 // We need to know when to write the package declaration and imports

	// Read the file character by character
	for {
		// Read the next character
		char, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			ExitOnError("reader.ReadRune(): ", err)
		}

		// Skip white space and Skip comments
		if buffer.Len() == 0 && (unicode.IsSpace(char)) {
			continue
		}

		// Skip comments --
		if buffer.Len() == 0 && char == '-' {
			reader.ReadLine()
			continue
		}

		// Skip multiline comments /* this is a comment */
		if strings.TrimSpace(buffer.String()) == "/" && char == '*' {
			// Skip all characters until the closing comment
			for {
				char, _, err := reader.ReadRune()
				if err == io.EOF {
					break
				}
				ExitOnError("reader.ReadRune()", err)
				if buffer.String()[buffer.Len()-1] == '*' && char == '/' {
					break
				}

				buffer.WriteRune(char)
			}

			// Reset the buffer and continue reading.
			buffer.Reset()
			continue
		}

		buffer.WriteRune(char)

		// Check if the buffer starts with "CREATE TYPE"
		if buffer.Len() >= 11 {
			if strings.ToUpper(strings.TrimSpace(buffer.String())[:11]) == "CREATE TYPE" {
				// Read characters until a semicolon is found
				for {
					char, _, err := reader.ReadRune()
					if err != nil {
						if err == io.EOF {
							break
						}
						ExitOnError("reader.ReadRune()", err)
					}

					buffer.WriteRune(char)
					if char == ';' {
						break
					}
				}

				// Extract the type name and values
				matches := re.FindStringSubmatch(buffer.String())

				// This Create statement of not of interest.
				if len(matches) != 3 {
					continue
				}

				// Get the typeName and values from capture groups
				typeName := strcase.ToCamel(enCaser.String(matches[1]))
				values := TransforEnums(strings.Split(matches[2], ","))

				count += 1
				// Create template context data
				data := TemplateData{
					PkgName:   pkgName,
					TypeName:  typeName,
					Values:    values,
					FirstType: count == 1,
				}

				// Format and write to the output file
				ParseTemplate(&outBuffer, data)
				output.Write(outBuffer.Bytes())

				// Reset the buffer
				buffer.Truncate(0)
			}
		}
	}

	// Format the code if we've written something.
	if count > 0 {
		// Format the output buffer
		b, err := format.Source(outBuffer.Bytes())
		ExitOnError("go/format failed", err)

		// Write formatted output to output file.
		err = os.WriteFile(outputFileName, b, 0644)
		ExitOnError("os.WriteFile error", err)
	} else {
		// Write an empty package to avoid errors
		err = os.WriteFile(outputFileName, []byte(fmt.Sprintf("package %s\n", pkgName)), 0644)
		ExitOnError("os.WriteFile", err)
	}
}

// The template string to generate enum constants and methods for the type.
var templateString = `
{{$typeName := .TypeName}}{{$values := .Values}}
{{if .FirstType}}
// Code generated by "goenums"; DO NOT EDIT.

package {{.PkgName}}
import (
	"database/sql/driver"
	"fmt"
)
{{end}}
type {{.TypeName}} string

const (
	{{range $val := $values -}}
       {{$typeName}}{{$val}} {{$typeName}} = "{{$val}}"
	{{end -}}
)

func (e {{.TypeName}}) IsValid() bool {
	validValues := []string{
		{{range $val := $values -}}
       		"{{$val -}}",
		{{end -}}
	}

	for _, val := range validValues {
		if val == string(e) {
			return true
		}
	}
	return false
}

func (e {{.TypeName}}) ValidValues() []string {
	return []string{
		{{range $val := $values -}}
       		"{{$val -}}",
		{{end -}}
	}
}


func (e *{{.TypeName}}) Scan(src interface{}) error {
	source, ok := src.(string)
	if !ok {
		return fmt.Errorf("invalid value for %s: %s", "{{.TypeName}}", source)
	}
	*e = {{.TypeName}}(source)
	return nil
}

func (e {{.TypeName}}) Value() (driver.Value, error) {
	if !e.IsValid() {
		return nil, fmt.Errorf("invalid value for %s", "{{.TypeName}}")
	}
	return string(e), nil
}
`
