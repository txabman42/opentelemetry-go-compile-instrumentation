// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"fmt"
	"log/slog"

	"github.com/dave/dst"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/internal/ast"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/internal/rule"
)

const (
	OtelRuntimeFile = "otel.runtime.go"
)

//nolint:gochecknoglobals // This is a constant
var requiredImports = map[string]string{
	"runtime/debug": "_otel_debug", // The getstack function depends on runtime/debug
	"log":           "_otel_log",   // The printstack function depends on log
	"unsafe":        "_",           // The golinkname tag depends on unsafe
}

func genImportDecl(matched []*rule.InstFuncRule) []dst.Decl {
	for _, m := range matched {
		requiredImports[m.Path] = "_"
	}
	importDecls := make([]dst.Decl, 0)
	for k, v := range requiredImports {
		importDecls = append(importDecls, ast.ImportDecl(v, k))
	}
	return importDecls
}

func genVarDecl(matched []*rule.InstFuncRule) []dst.Decl {
	decls := make([]dst.Decl, 0, len(matched))
	uniquePath := map[string]bool{}
	for i, m := range matched {
		if _, ok := uniquePath[m.Path]; ok {
			continue
		}
		uniquePath[m.Path] = true
		// First variable declaration
		// //go:linkname _getstatck%d %s.OtelGetStackImpl
		// var _getstatck%d = _otel_debug.Stack
		value := ast.SelectorExpr(ast.Ident("_otel_debug"), "Stack")
		getStackVar := ast.VarDecl(fmt.Sprintf("_getstatck%d", i), value)
		getStackVar.Decs = dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.NewLine,
				Start: dst.Decorations{
					fmt.Sprintf("//go:linkname _getstatck%d %s.OtelGetStackImpl",
						i, m.Path),
				},
			},
		}
		// Second variable declaration
		// //go:linkname _printstack%d %s.OtelPrintStackImpl
		// var _printstack%d = func (bt []byte){ _otel_log.Printf(string(bt)) }
		val := &dst.FuncLit{
			Type: &dst.FuncType{
				Params: &dst.FieldList{
					List: []*dst.Field{
						{
							Names: []*dst.Ident{
								{Name: "bt"},
							},
							Type: &dst.ArrayType{
								Elt: &dst.Ident{Name: "byte"},
							},
						},
					},
				},
			},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					&dst.ExprStmt{
						X: &dst.CallExpr{
							Fun: &dst.SelectorExpr{
								X:   &dst.Ident{Name: "_otel_log"},
								Sel: &dst.Ident{Name: "Printf"},
							},
							Args: []dst.Expr{
								&dst.CallExpr{
									Fun: &dst.Ident{Name: "string"},
									Args: []dst.Expr{
										&dst.Ident{Name: "bt"},
									},
								},
							},
						},
					},
				},
			},
		}
		printStackVar := ast.VarDecl(fmt.Sprintf("_printstack%d", i), val)
		printStackVar.Decs = dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.NewLine,
				Start: dst.Decorations{
					fmt.Sprintf("//go:linkname _printstack%d %s.OtelPrintStackImpl",
						i, m.Path),
				},
			},
		}
		decls = append(decls, getStackVar, printStackVar)
	}
	return decls
}

// genHookLinkNames generates go:linkname directives for hook functions.
// This eliminates the need for hook authors to manually add go:linkname directives.
func genHookLinkNames(matched []*rule.InstFuncRule) []dst.Decl {
	decls := make([]dst.Decl, 0)
	for _, m := range matched {
		slog.Info("hook rule", fmt.Sprintf("%+v", *m))
	}

	// Track unique hook names to avoid duplicates when multiple rules reference same hooks
	seenHooks := make(map[string]bool)

	for _, m := range matched {
		if m.Before != "" && !seenHooks[m.Before] {
			seenHooks[m.Before] = true
			// Generate linkname for Before hook with variadic signature
			// //go:linkname <HookName> <HookPackage>.<HookName>
			// func <HookName>(...interface{})
			hookDecl := &dst.FuncDecl{
				Name: ast.Ident(m.Before),
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{
								Type: &dst.Ellipsis{
									Elt: &dst.InterfaceType{
										Methods: &dst.FieldList{},
									},
								},
							},
						},
					},
				},
				Decs: dst.FuncDeclDecorations{
					NodeDecs: dst.NodeDecs{
						Before: dst.NewLine,
						Start: dst.Decorations{
							fmt.Sprintf("//go:linkname %s %s.%s",
								m.Before, m.Path, m.Before),
						},
					},
				},
			}
			decls = append(decls, hookDecl)
		}
		if m.After != "" && !seenHooks[m.After] {
			seenHooks[m.After] = true
			// Generate linkname for After hook with variadic signature
			// //go:linkname <HookName> <HookPackage>.<HookName>
			// func <HookName>(...interface{})
			hookDecl := &dst.FuncDecl{
				Name: ast.Ident(m.After),
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{
								Type: &dst.Ellipsis{
									Elt: &dst.InterfaceType{
										Methods: &dst.FieldList{},
									},
								},
							},
						},
					},
				},
				Decs: dst.FuncDeclDecorations{
					NodeDecs: dst.NodeDecs{
						Before: dst.NewLine,
						Start: dst.Decorations{
							fmt.Sprintf("//go:linkname %s %s.%s",
								m.After, m.Path, m.After),
						},
					},
				},
			}
			decls = append(decls, hookDecl)
		}
	}
	return decls
}

func buildOtelRuntimeAst(decls []dst.Decl) *dst.File {
	const comment = "// This file is generated by the opentelemetry-go-compile-instrumentation tool. DO NOT EDIT."
	return &dst.File{
		Name: ast.Ident("main"),
		Decs: dst.FileDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					comment,
				},
			},
		},
		Decls: decls,
	}
}

func (sp *SetupPhase) addDeps(matched []*rule.InstRuleSet) error {
	rules := make([]*rule.InstFuncRule, 0)
	for _, m := range matched {
		funcRules := m.GetFuncRules()
		rules = append(rules, funcRules...)
	}
	if len(rules) == 0 {
		return nil
	}

	// Add required imports
	importDecls := genImportDecl(rules)
	// Generate the variable declarations that used by otel runtime
	varDecls := genVarDecl(rules)
	// Generate go:linkname directives for hook functions
	hookLinkNames := genHookLinkNames(rules)
	// Build the ast with all declarations
	allDecls := append(importDecls, varDecls...)
	allDecls = append(allDecls, hookLinkNames...)
	root := buildOtelRuntimeAst(allDecls)
	// Write the ast to file
	err := ast.WriteFile(OtelRuntimeFile, root)
	if err != nil {
		return err
	}
	sp.keepForDebug(OtelRuntimeFile)
	return nil
}
