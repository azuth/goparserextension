package main

import (
	"bytes"
	"github.com/azuth/goparserextension/selectpExtension"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	// Parse the file "./selectp.go" with selectpstmts
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "./selectp.go", nil, parser.AllErrors)
	check(err)

	// find selectpStmt and transform them to nativ golang code
	selectpExtension.Transform(f)

	// pretty-print the AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, f)
	check(err)

	// save pretty-printed Ast to file "./parsedfile.go"
	err = ioutil.WriteFile("./parsedfile.go", []byte(buf.String()), 0644)
	check(err)
}
