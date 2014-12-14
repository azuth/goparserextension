package selectpExtension

import (
	"fmt"
	"go/ast"
	"go/token"
)

var nativeSelectLoopCounter = 0

func Transform(node ast.Node) bool {
	return walkP(node)
}

// walkP traverses an AST in depth-first order
// and replaces SelectPStmt with SelectStmt
func walkP(node ast.Node) bool {
	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for _, c := range n.List {
			walkP(c)
		}

	case *ast.Field:
		if n.Doc != nil {
			walkP(n.Doc)
		}
		walkPIdentList(n.Names)
		walkP(n.Type)
		if n.Tag != nil {
			walkP(n.Tag)
		}
		if n.Comment != nil {
			walkP(n.Comment)
		}

	case *ast.FieldList:
		for _, f := range n.List {
			walkP(f)
		}

		// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			walkP(n.Elt)
		}

	case *ast.FuncLit:
		walkP(n.Type)
		walkP(n.Body)

	case *ast.CompositeLit:
		if n.Type != nil {
			walkP(n.Type)
		}
		walkPExprList(n.Elts)

	case *ast.ParenExpr:
		walkP(n.X)

	case *ast.SelectorExpr:
		walkP(n.X)
		walkP(n.Sel)

	case *ast.IndexExpr:
		walkP(n.X)
		walkP(n.Index)

	case *ast.SliceExpr:
		walkP(n.X)
		if n.Low != nil {
			walkP(n.Low)
		}
		if n.High != nil {
			walkP(n.High)
		}
		if n.Max != nil {
			walkP(n.Max)
		}

	case *ast.TypeAssertExpr:
		walkP(n.X)
		if n.Type != nil {
			walkP(n.Type)
		}

	case *ast.CallExpr:
		walkP(n.Fun)
		walkPExprList(n.Args)

	case *ast.StarExpr:
		walkP(n.X)

	case *ast.UnaryExpr:
		walkP(n.X)

	case *ast.BinaryExpr:
		walkP(n.X)
		walkP(n.Y)

	case *ast.KeyValueExpr:
		walkP(n.Key)
		walkP(n.Value)

		// Types
	case *ast.ArrayType:
		if n.Len != nil {
			walkP(n.Len)
		}
		walkP(n.Elt)

	case *ast.StructType:
		walkP(n.Fields)

	case *ast.FuncType:
		if n.Params != nil {
			walkP(n.Params)
		}
		if n.Results != nil {
			walkP(n.Results)
		}

	case *ast.InterfaceType:
		walkP(n.Methods)

	case *ast.MapType:
		walkP(n.Key)
		walkP(n.Value)

	case *ast.ChanType:
		walkP(n.Value)

		// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		walkP(n.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		walkP(n.Label)
		if IsNodeSelectPStmt(n.Stmt) {
			n.Stmt = transformToNativeSelectStmt(n.Stmt.(*ast.SelectPStmt))
		}
		walkP(n.Stmt)

	case *ast.ExprStmt:
		walkP(n.X)

	case *ast.SendStmt:
		walkP(n.Chan)
		walkP(n.Value)

	case *ast.IncDecStmt:
		walkP(n.X)

	case *ast.AssignStmt:
		walkPExprList(n.Lhs)
		walkPExprList(n.Rhs)

	case *ast.GoStmt:
		walkP(n.Call)

	case *ast.DeferStmt:
		walkP(n.Call)

	case *ast.ReturnStmt:
		walkPExprList(n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			walkP(n.Label)
		}

	case *ast.BlockStmt:
		walkPStmtList(n.List)

	case *ast.IfStmt:
		if n.Init != nil {
			walkP(n.Init)
		}
		walkP(n.Cond)
		walkP(n.Body)
		if n.Else != nil {
			walkP(n.Else)
		}

	case *ast.CaseClause:
		walkPExprList(n.List)
		walkPStmtList(n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			walkP(n.Init)
		}
		if n.Tag != nil {
			walkP(n.Tag)
		}
		walkP(n.Body)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			walkP(n.Init)
		}
		walkP(n.Assign)
		walkP(n.Body)

	case *ast.CommClause:
		if n.Comm != nil {
			walkP(n.Comm)
		}
		walkPStmtList(n.Body)

	case *ast.SelectStmt:
		walkP(n.Body)

	case *ast.SelectPStmt:
		walkP(n.Body)
		return true
	case *ast.ForStmt:
		if n.Init != nil {
			walkP(n.Init)
		}
		if n.Cond != nil {
			walkP(n.Cond)
		}
		if n.Post != nil {
			walkP(n.Post)
		}
		walkP(n.Body)

	case *ast.RangeStmt:
		walkP(n.Key)
		if n.Value != nil {
			walkP(n.Value)
		}
		walkP(n.X)
		walkP(n.Body)

		// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			walkP(n.Doc)
		}
		if n.Name != nil {
			walkP(n.Name)
		}
		walkP(n.Path)
		if n.Comment != nil {
			walkP(n.Comment)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			walkP(n.Doc)
		}
		walkPIdentList(n.Names)
		if n.Type != nil {
			walkP(n.Type)
		}
		walkPExprList(n.Values)
		if n.Comment != nil {
			walkP(n.Comment)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			walkP(n.Doc)
		}
		walkP(n.Name)
		walkP(n.Type)
		if n.Comment != nil {
			walkP(n.Comment)
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			walkP(n.Doc)
		}
		for _, s := range n.Specs {
			walkP(s)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			walkP(n.Doc)
		}
		if n.Recv != nil {
			walkP(n.Recv)
		}
		walkP(n.Name)
		walkP(n.Type)
		if n.Body != nil {
			walkP(n.Body)
		}

		// Files and packages
	case *ast.File:
		if n.Doc != nil {
			walkP(n.Doc)
		}
		walkP(n.Name)
		walkPDeclList(n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for _, f := range n.Files {
			walkP(f)
		}

	default:
		fmt.Printf("walkP: unexpected node type %T", n)
		panic("walkP")
	}
	return false
}

func walkPIdentList(list []*ast.Ident) {
	for _, x := range list {
		walkP(x)
	}
}

func walkPExprList(list []ast.Expr) {
	for _, x := range list {
		walkP(x)
	}
}

func walkPStmtList(list []ast.Stmt) {
	for i, x := range list {
		if IsNodeSelectPStmt(x) {
			list[i] = transformToNativeSelectStmt(x.(*ast.SelectPStmt))
		}
		walkP(x)
	}
}

func walkPDeclList(list []ast.Decl) {
	for _, x := range list {
		walkP(x)
	}
}

// checks ast.node if its selectPStmt
func IsNodeSelectPStmt(node ast.Node) bool {
	switch node.(type) {
	case *ast.SelectPStmt:
		return true
	}
	return false
}

//transforms selectPStatmt to nested SelectStmt with lable,breakpoints and busy loop
func transformToNativeSelectStmt(selectPStmt *ast.SelectPStmt) *ast.LabeledStmt {
	loopIdent := "NativeSelectLoop" + fmt.Sprintf("%d", nativeSelectLoopCounter)
	nativeSelectLoopCounter++
	blockSt := &ast.BlockStmt{}
	blockSt.List = buildNestedSelectStmts(selectPStmt.Body.List, loopIdent)
	forSt := &ast.ForStmt{Body: blockSt}

	identSt := &ast.Ident{Name: loopIdent}
	lbldSt := &ast.LabeledStmt{Label: identSt, Stmt: forSt}
	return lbldSt
}

// walks recursiv throught cases and builds nested SelectStmts
func buildNestedSelectStmts(list []ast.Stmt, loopIdent string) []ast.Stmt {
	if len(list) > 0 {
		if commCl, ok := list[0].(*ast.CommClause); ok {
			resultList := make([]ast.Stmt, 1)

			// commands + break
			commCl.Body = append(commCl.Body, &ast.BranchStmt{Tok: token.BREAK, Label: &ast.Ident{Name: loopIdent}})
			stList := make([]ast.Stmt, 2)
			stList[0] = commCl

			// missing -> if only a default is in selectp

			stList[1] = &ast.CommClause{} // stList[1] = empty commClause if no default case in selectp (busy loop)

			// build nested selects in default case; commCl.Stmt == nil is default case
			if len(list) >= 2 {
				if len(list) == 2 && list[1].(*ast.CommClause).Comm == nil { // if default case in selectP
					stList[1] = list[1]
					stList[1].(*ast.CommClause).Body = append(stList[1].(*ast.CommClause).Body, &ast.BranchStmt{Tok: token.BREAK, Label: &ast.Ident{Name: loopIdent}})
				} else {
					stList[1].(*ast.CommClause).Body = buildNestedSelectStmts(list[1:], loopIdent)
				}
			}

			// build select statement
			blockSt := &ast.BlockStmt{List: stList}
			selectSt := &ast.SelectStmt{Body: blockSt}

			resultList = make([]ast.Stmt, 1)
			resultList[0] = selectSt

			return resultList
		}
	}
	return make([]ast.Stmt, 0) //nil
}
