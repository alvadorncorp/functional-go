package archtest

import "go/ast"

type surfaceTypeEdges struct {
	references []surfaceSymbol
	localType  bool
}

func addSurfaceTypeEdges(graph map[surfaceSymbol]surfaceTypeEdges, pkg string, file *ast.File) {
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if declaration.Recv == nil && declaration.Name.IsExported() {
				key := surfaceSymbol{pkg, declaration.Name.Name}
				bound := typeParameterNames(declaration.Type.TypeParams)
				graph[key] = surfaceTypeEdges{references: localFieldTypes(pkg, declaration.Type.Results, bound, false)}
			}
		case *ast.GenDecl:
			for _, specification := range declaration.Specs {
				if declared, ok := specification.(*ast.TypeSpec); ok {
					key := surfaceSymbol{pkg, declared.Name.Name}
					graph[key] = surfaceTypeEdges{
						references: localTypeReferences(pkg, declared.Type, typeParameterNames(declared.TypeParams)),
						localType:  true,
					}
				}
			}
		}
	}
}

func typeParameterNames(parameters *ast.FieldList) map[string]bool {
	bound := map[string]bool{}
	if parameters != nil {
		for _, parameter := range parameters.List {
			for _, name := range parameter.Names {
				bound[name.Name] = true
			}
		}
	}
	return bound
}

func localFieldTypes(pkg string, fields *ast.FieldList, bound map[string]bool, exportedOnly bool) []surfaceSymbol {
	var references []surfaceSymbol
	if fields == nil {
		return references
	}
	for _, field := range fields.List {
		if exportedOnly && !exportedTypeField(field) {
			continue
		}
		references = append(references, localTypeReferences(pkg, field.Type, bound)...)
	}
	return references
}

func exportedTypeField(field *ast.Field) bool {
	if len(field.Names) == 0 {
		return ast.IsExported(receiverBaseName(field.Type))
	}
	for _, name := range field.Names {
		if name.IsExported() {
			return true
		}
	}
	return false
}

// Follow type positions only, never declaration names, parameters or bodies.
func localTypeReferences(pkg string, expression ast.Expr, bound map[string]bool) []surfaceSymbol {
	var children []ast.Expr
	switch expression := expression.(type) {
	case *ast.Ident:
		if !bound[expression.Name] {
			return []surfaceSymbol{{pkg, expression.Name}}
		}
	case *ast.ParenExpr:
		children = []ast.Expr{expression.X}
	case *ast.StarExpr:
		children = []ast.Expr{expression.X}
	case *ast.ArrayType:
		children = []ast.Expr{expression.Elt}
	case *ast.MapType:
		children = []ast.Expr{expression.Key, expression.Value}
	case *ast.ChanType:
		children = []ast.Expr{expression.Value}
	case *ast.Ellipsis:
		children = []ast.Expr{expression.Elt}
	case *ast.IndexExpr:
		children = []ast.Expr{expression.X, expression.Index}
	case *ast.IndexListExpr:
		children = append([]ast.Expr{expression.X}, expression.Indices...)
	case *ast.StructType:
		return localFieldTypes(pkg, expression.Fields, bound, true)
	case *ast.FuncType:
		return localFieldTypes(pkg, expression.Results, bound, false)
	case *ast.InterfaceType:
		for _, field := range expression.Methods.List {
			if len(field.Names) == 0 {
				children = append(children, field.Type)
			}
		}
	case *ast.UnaryExpr:
		children = []ast.Expr{expression.X}
	case *ast.BinaryExpr:
		children = []ast.Expr{expression.X, expression.Y}
	}
	var references []surfaceSymbol
	for _, child := range children {
		references = append(references, localTypeReferences(pkg, child, bound)...)
	}
	return references
}

func reachSurfaceTypes(consumed map[surfaceSymbol]bool, graph map[surfaceSymbol]surfaceTypeEdges) {
	var pending []surfaceSymbol
	for symbol := range consumed {
		pending = append(pending, symbol)
	}
	for len(pending) > 0 {
		symbol := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for _, referenced := range graph[symbol].references {
			if graph[referenced].localType && !consumed[referenced] {
				consumed[referenced] = true
				pending = append(pending, referenced)
			}
		}
	}
}
