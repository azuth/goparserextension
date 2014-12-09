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
			if len(list) >= 2 {
				if len(list) == 2 && list[1].(*ast.CommClause).Comm == nil {
					stList[1].(*ast.CommClause).Body = list[1].(*ast.CommClause).Body
				} else {
					stList[1].(*ast.CommClause).Body = buildNestedSlects(list[1:])
				}
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

// Might be incomplete
func walkThroughStmnt(list []ast.Stmt) {
	for i, stmt := range list {

		if labeled, ok := stmt.(*ast.LabeledStmt); ok {
			if selectPSt, ok := labeled.Stmt.(*ast.SelectPStmt); ok {
				walkThroughStmnt(selectPSt.Body.List)
				list[i].(*ast.LabeledStmt).Stmt = transformToNativSlectPStmt(selectPSt)
			}
		} else if blk, ok := stmt.(*ast.BlockStmt); ok {
			walkThroughStmnt(blk.List)
		} else if ifst, ok := stmt.(*ast.IfStmt); ok {
			walkThroughStmnt(ifst.Body.List)
		} else if caseCl, ok := stmt.(*ast.CaseClause); ok {
			walkThroughStmnt(caseCl.Body)
		} else if switchSt, ok := stmt.(*ast.SwitchStmt); ok {
			walkThroughStmnt(switchSt.Body.List)
		} else if typeSwitchSt, ok := stmt.(*ast.TypeSwitchStmt); ok {
			walkThroughStmnt(typeSwitchSt.Body.List)
		} else if commClause, ok := stmt.(*ast.CommClause); ok {
			walkThroughStmnt(commClause.Body)
		} else if selectSt, ok := stmt.(*ast.SelectStmt); ok {
			walkThroughStmnt(selectSt.Body.List)
		} else if selectPSt, ok := stmt.(*ast.SelectPStmt); ok {
			walkThroughStmnt(selectPSt.Body.List)
			list[i] = transformToNativSlectPStmt(selectPSt)
		} else if forSt, ok := stmt.(*ast.ForStmt); ok {
			walkThroughStmnt(forSt.Body.List)
		} else if rangeSt, ok := stmt.(*ast.RangeStmt); ok {
			walkThroughStmnt(rangeSt.Body.List)
		}

	}
}

// Might be incomplete
func changeAst(list []ast.Decl) []ast.Decl {
	for _, decl := range list {
		//if gen, ok := decl.(*ast.GenDecl); ok {
		//} else

		if fun, ok := decl.(*ast.FuncDecl); ok {
			walkThroughStmnt(fun.Body.List)
		}
	}

	return list
}

func main() {
	fset := token.NewFileSet()

	// Parse the file
	f, err := parser.ParseFile(fset, "./selectp.go", nil, 0)
	check(err)

	//ast.Print(fset, f) // print Ast

	f.Decls = changeAst(f.Decls)

	//ast.Print(fset, f)

	// pretty-print the AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, f)
	check(err)
	err = ioutil.WriteFile("./parsedfile.go", []byte(buf.String()), 0644)
	check(err)
}
