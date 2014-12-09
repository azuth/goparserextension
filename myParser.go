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
	fset := token.NewFileSet()
	// Parse the file
	f, err := parser.ParseFile(fset, "./selectp.go", nil, parser.AllErrors)
	check(err)

	selectpExtension.Transform(f)

	// pretty-print the AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, f)
	check(err)

	//ast.Print(fset, f)

	err = ioutil.WriteFile("./parsedfile.go", []byte(buf.String()), 0644)
	check(err)
}
