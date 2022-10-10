// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"go/ast"
	"go/token"

	"github.com/sumologic-labs/autotel/lib"
	"golang.org/x/tools/go/packages"
)

func insert(a []ast.Stmt, index int, value ast.Stmt) []ast.Stmt {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

type HttpRewriter struct {
}

func (pass *HttpRewriter) Execute(
	node *ast.File,
	analysis *lib.Analysis,
	pkg *packages.Package,
	pkgs []*packages.Package) []lib.Import {

	var imports []lib.Import
	addImports := false
	addContext := false
	var handlerCallback *ast.Ident

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "HandlerFunc" && sel.X.(*ast.Ident).Name == "http" {
					handlerCallback = x.Args[0].(*ast.Ident)
				}
			}

		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			if ident, ok := x.Lhs[0].(*ast.Ident); ok {
				pkgPath := ""
				pkgPath = lib.GetPkgNameFromDefsTable(pkg, ident)
				if pkg.TypesInfo.Defs[ident] == nil {
					return false
				}
				if handlerCallback == nil || pkg.TypesInfo.Uses[handlerCallback] == nil {
					return false
				}
				if pkg.TypesInfo.Uses[handlerCallback].Name() == pkg.TypesInfo.Defs[ident].Name() {
					fundId := pkgPath + "." + pkg.TypesInfo.Defs[ident].Name()
					fun := lib.FuncDescriptor{
						Id:              fundId,
						DeclType:        pkg.TypesInfo.Defs[ident].Type().String(),
						CustomInjection: true}
					analysis.Callgraph[fun] = []lib.FuncDescriptor{}
				}
			}
			for _, e := range x.Rhs {
				// TODO check correctly parameter types and names
				if funLit, ok := e.(*ast.FuncLit); ok {
					reqCtx := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "__atel_child_tracing_ctx",
							},
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "req",
									},
									Sel: &ast.Ident{
										Name: "Context",
									},
								},
								Lparen:   45,
								Ellipsis: 0,
							},
						},
					}
					span := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "__atel_http_span",
							},
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "__atel_trace",
									},
									Sel: &ast.Ident{
										Name: "SpanFromContext",
									},
								},
								Lparen: 56,
								Args: []ast.Expr{
									&ast.Ident{
										Name: "__atel_child_tracing_ctx",
									},
								},
								Ellipsis: 0,
							},
						},
					}
					spanSupress := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "_",
							},
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.Ident{
								Name: "__atel_http_span",
							},
						},
					}
					funLit.Body.List = append([]ast.Stmt{reqCtx, span, spanSupress}, funLit.Body.List...)
					addImports = true
					addContext = true

					imports = append(imports, lib.Import{"__atel_trace", "go.opentelemetry.io/otel/trace", lib.Add})

				}
			}
		}
		return true
	})
	var handlerIdent *ast.Ident
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			handlerIndex := -1
			for _, body := range x.Body.List {
				handlerIndex = handlerIndex + 1
				if assignment, ok := body.(*ast.AssignStmt); ok {
					if call, ok := assignment.Rhs[0].(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if sel.Sel.Name == "HandlerFunc" && sel.X.(*ast.Ident).Name == "http" {
								handlerCallback = call.Args[0].(*ast.Ident)
								handlerIdent = assignment.Lhs[0].(*ast.Ident)
								break
							}
						}
					}
				}
			}

			if len(x.Body.List) > 1 && handlerCallback != nil && handlerIdent != nil {
				copy(x.Body.List[handlerIndex:], x.Body.List[handlerIndex+1:])
				x.Body.List[len(x.Body.List)-1] = nil
				otelHadlerStmt := &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: handlerIdent.Name,
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "__atel_otelhttp",
								},
								Sel: &ast.Ident{
									Name: "NewHandler",
								},
							},
							Lparen: 61,
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "http",
										},
										Sel: &ast.Ident{
											Name: "HandlerFunc",
										},
									},
									Lparen: 78,
									Args: []ast.Expr{
										&ast.Ident{
											Name: handlerCallback.Name,
										},
									},
									Ellipsis: 0,
								},
								&ast.Ident{
									Name: `"` + handlerCallback.Name + `"`,
								},
							},
							Ellipsis: 0,
						},
					},
				}
				insert(x.Body.List, handlerIndex, otelHadlerStmt)
				addImports = true
				addContext = true
			}
		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			var clientVar *ast.Ident
			clientVarIndex := -1
			for _, body := range x.Body.List {
				clientVarIndex = clientVarIndex + 1
				if assignment, ok := body.(*ast.AssignStmt); ok {
					if lit, ok := assignment.Rhs[0].(*ast.CompositeLit); ok {
						if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
							if sel.Sel.Name == "Client" && sel.X.(*ast.Ident).Name == "http" {
								clientVar = assignment.Lhs[0].(*ast.Ident)
								break
							}
						}
					}
				}
			}

			if len(x.Body.List) > 1 && clientVar != nil {
				copy(x.Body.List[clientVarIndex:], x.Body.List[clientVarIndex+1:])
				x.Body.List[len(x.Body.List)-1] = nil
				newClientVar := &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: clientVar.Name,
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CompositeLit{
							Type: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "http",
								},
								Sel: &ast.Ident{
									Name: "Client",
								},
							},
							Elts: []ast.Expr{
								&ast.KeyValueExpr{
									Key: &ast.Ident{
										Name: "Transport",
									},
									Colon: 58,
									Value: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "__atel_otelhttp",
											},
											Sel: &ast.Ident{
												Name: "NewTransport",
											},
										},
										Lparen: 81,
										Args: []ast.Expr{
											&ast.SelectorExpr{
												X: &ast.Ident{
													Name: "http",
												},
												Sel: &ast.Ident{
													Name: "DefaultTransport",
												},
											},
										},
										Ellipsis: 0,
									},
								},
							},
							Incomplete: false,
						},
					},
				}
				insert(x.Body.List, clientVarIndex, newClientVar)
				addImports = true
				addContext = true
			}
		}
		return true
	})

	if addContext {
		imports = append(imports, lib.Import{"__atel_context", "context", lib.Add})
	}
	if addImports {
		imports = append(imports, lib.Import{"__atel_otel", "go.opentelemetry.io/otel", lib.Add})
		imports = append(imports, lib.Import{"__atel_otelhttp", "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp", lib.Add})
	}
	return imports
}
