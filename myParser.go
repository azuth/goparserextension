package main

import (
	"bytes"
	//"fmt"
	"go/ast"
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

func buildNestedSlects(list []ast.Stmt) []ast.Stmt {
	if len(list) > 0 {
		if commCl, ok := list[0].(*ast.CommClause); ok {
			stList := make([]ast.Stmt, 2)
			stList[0] = commCl
			stList[1] = &ast.CommClause{}
			if len(list) > 1 {
				stList[1].(*ast.CommClause).Body = buildNestedSlects(list[1:])
			}
			blockSt := &ast.BlockStmt{List: stList}
			selectSt := &ast.SelectStmt{Body: blockSt}
			newlist := make([]ast.Stmt, 1)
			newlist[0] = selectSt
			return newlist
		}
	}
	return make([]ast.Stmt, 0)
}

func transformToNativSlectPStmt(selectPStmt *ast.SelectPStmt) *ast.ForStmt {
	blockSt := &ast.BlockStmt{}
	blockSt.List = buildNestedSlects(selectPStmt.Body.List)
	forSt := &ast.ForStmt{Body: blockSt}

	return forSt
}

func ReplaceInspector(n ast.Node) bool {
	switch x := n.(type) {
	case *ast.SelectPStmt:
		stList := make([]ast.Stmt, 1)
		block := &ast.BlockStmt{}
		block.List = stList
		block.List[0] = transformToNativSlectPStmt(x)
		x.Body = block

	}
	return true
}

func main() {
	fset := token.NewFileSet()

	// Parse the file
	f, err := parser.ParseFile(fset, "./selectp.go", nil, 0)
	check(err)

	//ast.Print(fset, f) // print Ast

	ast.Inspect(f, ReplaceInspector)

	// pretty-print the AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, f)
	check(err)
	err = ioutil.WriteFile("./parsedfile.go", []byte(buf.String()), 0644)
	check(err)
}
