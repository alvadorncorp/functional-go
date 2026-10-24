package archtest

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strconv"
)

type surfaceSymbol struct {
	pkg  string
	name string
}

type shadowSpan struct {
	start token.Pos
	end   token.Pos
}

func collectSurfaceSelectors(file *ast.File, consumed map[surfaceSymbol]bool) error {
	aliases := map[string]string{}
	surface := exactSurface()
	for _, imported := range file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			return err
		}
		if _, tracked := surface[path]; !tracked {
			continue
		}
		alias := shortPackageName(path)
		if imported.Name != nil {
			alias = imported.Name.Name
		}
		switch alias {
		case ".":
			return fmt.Errorf("dot import of tracked surface package %s", path)
		case "_":
			continue
		default:
			aliases[alias] = path
		}
	}
	shadows := importShadowSpans(file, aliases)
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		qualifier, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		path, tracked := aliases[qualifier.Name]
		if tracked && !slices.ContainsFunc(shadows[qualifier.Name], func(span shadowSpan) bool {
			return span.start <= qualifier.Pos() && qualifier.Pos() < span.end
		}) {
			consumed[surfaceSymbol{path, selector.Sel.Name}] = true
		}
		return true
	})
	return nil
}

// Only import aliases need binding evidence; compilation owns type validity.
func importShadowSpans(file *ast.File, aliases map[string]string) map[string][]shadowSpan {
	spans := map[string][]shadowSpan{}
	add := func(name *ast.Ident, start, end token.Pos) {
		if name != nil && aliases[name.Name] != "" {
			spans[name.Name] = append(spans[name.Name], shadowSpan{start, end})
		}
	}
	var stack []ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		scopeEnd := lexicalScopeEnd(stack)
		switch declaration := node.(type) {
		case *ast.ValueSpec:
			for _, name := range declaration.Names {
				add(name, declaration.End(), scopeEnd)
			}
		case *ast.TypeSpec:
			add(declaration.Name, declaration.Name.Pos(), scopeEnd)
			addTypeParameterSpans(declaration.TypeParams, declaration.Name.End(), declaration.End(), add)
		case *ast.AssignStmt:
			addAssignmentSpans(declaration, stack, scopeEnd, add)
		case *ast.RangeStmt:
			if declaration.Tok == token.DEFINE {
				for _, expression := range []ast.Expr{declaration.Key, declaration.Value} {
					if name, ok := expression.(*ast.Ident); ok {
						add(name, declaration.Body.Pos(), declaration.Body.End())
					}
				}
			}
		case *ast.FuncDecl:
			addFunctionBodySpans(declaration.Recv, declaration.Type, declaration.Body, add)
			addTypeParameterSpans(declaration.Type.TypeParams, declaration.Name.End(), declaration.End(), add)
			addReceiverTypeSpans(declaration, add)
		case *ast.FuncLit:
			addFunctionBodySpans(nil, declaration.Type, declaration.Body, add)
		}
		stack = append(stack, node)
		return true
	})
	return spans
}

func lexicalScopeEnd(stack []ast.Node) token.Pos {
	for i := len(stack) - 1; i >= 0; i-- {
		switch node := stack[i].(type) {
		case *ast.BlockStmt, *ast.CaseClause, *ast.CommClause,
			*ast.IfStmt, *ast.ForStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			return node.End()
		}
	}
	return token.NoPos
}

func addAssignmentSpans(statement *ast.AssignStmt, stack []ast.Node, end token.Pos,
	add func(*ast.Ident, token.Pos, token.Pos),
) {
	if statement.Tok != token.DEFINE {
		return
	}
	if parent, ok := stack[len(stack)-1].(*ast.TypeSwitchStmt); ok && parent.Assign == statement {
		for _, clause := range parent.Body.List {
			branch := clause.(*ast.CaseClause)
			for _, expression := range statement.Lhs {
				if name, named := expression.(*ast.Ident); named {
					add(name, branch.Colon, branch.End())
				}
			}
		}
		return
	}
	for _, expression := range statement.Lhs {
		if name, ok := expression.(*ast.Ident); ok {
			add(name, statement.End(), end)
		}
	}
}

func addTypeParameterSpans(parameters *ast.FieldList, start, end token.Pos,
	add func(*ast.Ident, token.Pos, token.Pos),
) {
	if parameters == nil {
		return
	}
	for _, parameter := range parameters.List {
		for _, name := range parameter.Names {
			add(name, start, end)
		}
	}
}

func addFunctionBodySpans(receiver *ast.FieldList, signature *ast.FuncType, body *ast.BlockStmt,
	add func(*ast.Ident, token.Pos, token.Pos),
) {
	if body == nil {
		return
	}
	for _, fields := range []*ast.FieldList{receiver, signature.Params, signature.Results} {
		addTypeParameterSpans(fields, body.Pos(), body.End(), add)
	}
}

func addReceiverTypeSpans(function *ast.FuncDecl, add func(*ast.Ident, token.Pos, token.Pos)) {
	if function.Recv == nil {
		return
	}
	for _, receiver := range function.Recv.List {
		expression := receiver.Type
		if pointer, ok := expression.(*ast.StarExpr); ok {
			expression = pointer.X
		}
		var parameters []ast.Expr
		switch indexed := expression.(type) {
		case *ast.IndexExpr:
			parameters = []ast.Expr{indexed.Index}
		case *ast.IndexListExpr:
			parameters = indexed.Indices
		}
		for _, parameter := range parameters {
			if name, ok := parameter.(*ast.Ident); ok {
				add(name, function.Name.End(), function.End())
			}
		}
	}
}
