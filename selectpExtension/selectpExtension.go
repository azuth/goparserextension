package selectpExtension

import (
	"fmt"
	"go/ast"
	"go/token"
)

func buildNestedSelects(list []ast.Stmt, loopIdent string) []ast.Stmt {
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
					stList[1].(*ast.CommClause).Body = buildNestedSelects(list[1:], loopIdent)
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

var nativeSelectLoopCounter = 0

func transformToNativeSelectPStmt(selectPStmt *ast.SelectPStmt) *ast.LabeledStmt {
	loopIdent := "NativeSelectLoop" + fmt.Sprintf("%d", nativeSelectLoopCounter)
	nativeSelectLoopCounter++
	blockSt := &ast.BlockStmt{}
	blockSt.List = buildNestedSelects(selectPStmt.Body.List, loopIdent)
	forSt := &ast.ForStmt{Body: blockSt}

	identSt := &ast.Ident{Name: loopIdent}
	lbldSt := &ast.LabeledStmt{Label: identSt, Stmt: forSt}
	return lbldSt
}

func CheckNode(node ast.Node) bool {
	switch node.(type) {
	case *ast.SelectPStmt:
		return true
	}
	return false
}

func Transform(node ast.Node) bool {
	return WalkNew(node)
}

// WalkNew traverses an AST in depth-first order
// and replaces SelectPStmt with SelectStmt
func WalkNew(node ast.Node) bool {
	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for _, c := range n.List {
			WalkNew(c)
		}

	case *ast.Field:
		if n.Doc != nil {
			WalkNew(n.Doc)
		}
		walkNewIdentList(n.Names)
		WalkNew(n.Type)
		if n.Tag != nil {
			WalkNew(n.Tag)
		}
		if n.Comment != nil {
			WalkNew(n.Comment)
		}

	case *ast.FieldList:
		for _, f := range n.List {
			WalkNew(f)
		}

		// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			WalkNew(n.Elt)
		}

	case *ast.FuncLit:
		WalkNew(n.Type)
		WalkNew(n.Body)

	case *ast.CompositeLit:
		if n.Type != nil {
			WalkNew(n.Type)
		}
		walkNewExprList(n.Elts)

	case *ast.ParenExpr:
		WalkNew(n.X)

	case *ast.SelectorExpr:
		WalkNew(n.X)
		WalkNew(n.Sel)

	case *ast.IndexExpr:
		WalkNew(n.X)
		WalkNew(n.Index)

	case *ast.SliceExpr:
		WalkNew(n.X)
		if n.Low != nil {
			WalkNew(n.Low)
		}
		if n.High != nil {
			WalkNew(n.High)
		}
		if n.Max != nil {
			WalkNew(n.Max)
		}

	case *ast.TypeAssertExpr:
		WalkNew(n.X)
		if n.Type != nil {
			WalkNew(n.Type)
		}

	case *ast.CallExpr:
		WalkNew(n.Fun)
		walkNewExprList(n.Args)

	case *ast.StarExpr:
		WalkNew(n.X)

	case *ast.UnaryExpr:
		WalkNew(n.X)

	case *ast.BinaryExpr:
		WalkNew(n.X)
		WalkNew(n.Y)

	case *ast.KeyValueExpr:
		WalkNew(n.Key)
		WalkNew(n.Value)

		// Types
	case *ast.ArrayType:
		if n.Len != nil {
			WalkNew(n.Len)
		}
		WalkNew(n.Elt)

	case *ast.StructType:
		WalkNew(n.Fields)

	case *ast.FuncType:
		if n.Params != nil {
			WalkNew(n.Params)
		}
		if n.Results != nil {
			WalkNew(n.Results)
		}

	case *ast.InterfaceType:
		WalkNew(n.Methods)

	case *ast.MapType:
		WalkNew(n.Key)
		WalkNew(n.Value)

	case *ast.ChanType:
		WalkNew(n.Value)

		// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		WalkNew(n.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		WalkNew(n.Label)
		if CheckNode(n.Stmt) {
			n.Stmt = transformToNativeSelectPStmt(n.Stmt.(*ast.SelectPStmt))
		}
		WalkNew(n.Stmt)

	case *ast.ExprStmt:
		WalkNew(n.X)

	case *ast.SendStmt:
		WalkNew(n.Chan)
		WalkNew(n.Value)

	case *ast.IncDecStmt:
		WalkNew(n.X)

	case *ast.AssignStmt:
		walkNewExprList(n.Lhs)
		walkNewExprList(n.Rhs)

	case *ast.GoStmt:
		WalkNew(n.Call)

	case *ast.DeferStmt:
		WalkNew(n.Call)

	case *ast.ReturnStmt:
		walkNewExprList(n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			WalkNew(n.Label)
		}

	case *ast.BlockStmt:
		walkNewStmtList(n.List)

	case *ast.IfStmt:
		if n.Init != nil {
			WalkNew(n.Init)
		}
		WalkNew(n.Cond)
		WalkNew(n.Body)
		if n.Else != nil {
			WalkNew(n.Else)
		}

	case *ast.CaseClause:
		walkNewExprList(n.List)
		walkNewStmtList(n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			WalkNew(n.Init)
		}
		if n.Tag != nil {
			WalkNew(n.Tag)
		}
		WalkNew(n.Body)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			WalkNew(n.Init)
		}
		WalkNew(n.Assign)
		WalkNew(n.Body)

	case *ast.CommClause:
		if n.Comm != nil {
			WalkNew(n.Comm)
		}
		walkNewStmtList(n.Body)

	case *ast.SelectStmt:
		WalkNew(n.Body)

	case *ast.SelectPStmt:
		WalkNew(n.Body)
		return true
	case *ast.ForStmt:
		if n.Init != nil {
			WalkNew(n.Init)
		}
		if n.Cond != nil {
			WalkNew(n.Cond)
		}
		if n.Post != nil {
			WalkNew(n.Post)
		}
		WalkNew(n.Body)

	case *ast.RangeStmt:
		WalkNew(n.Key)
		if n.Value != nil {
			WalkNew(n.Value)
		}
		WalkNew(n.X)
		WalkNew(n.Body)

		// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			WalkNew(n.Doc)
		}
		if n.Name != nil {
			WalkNew(n.Name)
		}
		WalkNew(n.Path)
		if n.Comment != nil {
			WalkNew(n.Comment)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			WalkNew(n.Doc)
		}
		walkNewIdentList(n.Names)
		if n.Type != nil {
			WalkNew(n.Type)
		}
		walkNewExprList(n.Values)
		if n.Comment != nil {
			WalkNew(n.Comment)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			WalkNew(n.Doc)
		}
		WalkNew(n.Name)
		WalkNew(n.Type)
		if n.Comment != nil {
			WalkNew(n.Comment)
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			WalkNew(n.Doc)
		}
		for _, s := range n.Specs {
			WalkNew(s)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			WalkNew(n.Doc)
		}
		if n.Recv != nil {
			WalkNew(n.Recv)
		}
		WalkNew(n.Name)
		WalkNew(n.Type)
		if n.Body != nil {
			WalkNew(n.Body)
		}

		// Files and packages
	case *ast.File:
		if n.Doc != nil {
			WalkNew(n.Doc)
		}
		WalkNew(n.Name)
		walkNewDeclList(n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for _, f := range n.Files {
			WalkNew(f)
		}

	default:
		fmt.Printf("WalkNew: unexpected node type %T", n)
		panic("WalkNew")
	}
	return false
}

func walkNewIdentList(list []*ast.Ident) {
	for _, x := range list {
		WalkNew(x)
	}
}

func walkNewExprList(list []ast.Expr) {
	for _, x := range list {
		WalkNew(x)
	}
}

func walkNewStmtList(list []ast.Stmt) {
	for i, x := range list {
		if CheckNode(x) {
			list[i] = transformToNativeSelectPStmt(x.(*ast.SelectPStmt))
		}
		WalkNew(x)
	}
}

func walkNewDeclList(list []ast.Decl) {
	for _, x := range list {
		WalkNew(x)
	}
}
