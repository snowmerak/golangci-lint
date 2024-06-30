package snowygo

import (
	"github.com/golangci/golangci-lint/pkg/config"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"strings"
)

func NewAnalyzerWithConfig(cfg *config.SnowyGoSettings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "snowygo",
		Doc:  "snowmerak's custom linter for Go.",
		Run:  runAnalyzer(cfg),
	}
}

type Reporter struct {
	report func(pos token.Pos, format string, args ...interface{})
	pos    token.Pos
}

type PairElement struct {
	Name     string
	Reporter Reporter
}

type Pair struct {
	First  *PairElement
	Second *PairElement
}

func runAnalyzer(cfg *config.SnowyGoSettings) func(pass *analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		a2b := make(map[string]*Pair)

		for _, f := range pass.Files {
			ast.Inspect(f, func(node ast.Node) bool {
				// when call expression is found
				if node, ok := node.(*ast.CallExpr); ok {
					// when selector expression is found
					if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
						// check if fmt.Print* is used
						if x, ok := sel.X.(*ast.Ident); ok {
							if x.Name == "fmt" {
								switch sel.Sel.Name {
								case "Print":
									pass.Reportf(node.Pos(), "fmt.Print should not be used")
								case "Printf":
									pass.Reportf(node.Pos(), "fmt.Printf should not be used")
								case "Println":
									pass.Reportf(node.Pos(), "fmt.Println should not be used")
								}
							}
						}
					}
				}

				// when variable declaration is found
				if genDecl, ok := node.(*ast.GenDecl); ok {
					if genDecl.Tok == token.VAR {
						for _, spec := range genDecl.Specs {
							if valueSpec, ok := spec.(*ast.ValueSpec); ok {
								if len(valueSpec.Names) > 0 {
									if valueSpec.Names[0].IsExported() {
										pass.Reportf(valueSpec.Pos(), "global variable %s should not be exported", valueSpec.Names[0].Name)
									}
								}
							}
						}
					}
				}

				// when function declaration is found
				if fn, ok := node.(*ast.FuncDecl); ok {
					// check if function name has prefix, 'Request', 'Reply', 'Send', 'Receive', 'Publish', 'Subscribe', 'Set', 'Get'
					trimmedName := strings.TrimPrefix(fn.Name.Name, "Request")
					trimmedName = strings.TrimPrefix(trimmedName, "Reply")
					trimmedName = strings.TrimPrefix(trimmedName, "Send")
					trimmedName = strings.TrimPrefix(trimmedName, "Receive")
					trimmedName = strings.TrimPrefix(trimmedName, "Publish")
					trimmedName = strings.TrimPrefix(trimmedName, "Subscribe")
					trimmedName = strings.TrimPrefix(trimmedName, "Set")
					trimmedName = strings.TrimPrefix(trimmedName, "Get")
					p := a2b[f.Name.Name+"."+trimmedName]
					if p == nil {
						p = &Pair{}
						a2b[f.Name.Name+"."+trimmedName] = p
					}
					reporter := Reporter{
						report: pass.Reportf,
						pos:    fn.Pos(),
					}

					switch {
					case strings.HasPrefix(fn.Name.Name, "Request"):
						p.First = &PairElement{
							Name:     "Request",
							Reporter: reporter,
						}
					case strings.HasPrefix(fn.Name.Name, "Reply"):
						p.Second = &PairElement{
							Name:     "Reply",
							Reporter: reporter,
						}
					case strings.HasPrefix(fn.Name.Name, "Send"):
						p.First = &PairElement{
							Name:     "Send",
							Reporter: reporter,
						}
					case strings.HasPrefix(fn.Name.Name, "Receive"):
						p.Second = &PairElement{
							Name:     "Receive",
							Reporter: reporter,
						}
					case strings.HasPrefix(fn.Name.Name, "Publish"):
						p.First = &PairElement{
							Name:     "Publish",
							Reporter: reporter,
						}
					case strings.HasPrefix(fn.Name.Name, "Subscribe"):
						p.Second = &PairElement{
							Name:     "Subscribe",
							Reporter: reporter,
						}
					case strings.HasPrefix(fn.Name.Name, "Set"):
						p.First = &PairElement{
							Name:     "Set",
							Reporter: reporter,
						}
					case strings.HasPrefix(fn.Name.Name, "Get"):
						p.Second = &PairElement{
							Name:     "Get",
							Reporter: reporter,
						}
					}

					// when function has parameters
					if fn.Type.Params != nil {
						// check if the function has a context.Context parameter
						hasCtx := false
						isCtxFirst := false
						for i, param := range fn.Type.Params.List {
							if len(param.Names) >= 1 && param.Names[0].Name == "ctx" {
								hasCtx = true
								if i == 0 {
									isCtxFirst = true
								}
								break
							}
						}

						// report if the function has a context.Context parameter but it's not the first parameter
						if hasCtx && !isCtxFirst {
							pass.Reportf(fn.Pos(), "context.Context should be the first parameter")
						}
					}

					// when function has return values
					if fn.Type.Results != nil {
						hasErr := false
						isErrLast := false
						for i, param := range fn.Type.Results.List {
							if ident, ok := param.Type.(*ast.Ident); ok && ident.Name == "error" {
								hasErr = true
								if i == len(fn.Type.Results.List)-1 {
									isErrLast = true
								}
							}
						}

						// report if the function has an error return value but it's not the last return value
						if hasErr && !isErrLast {
							pass.Reportf(fn.Pos(), "error should be the last return value")
						}
					}

					// when function has a body
					if len(fn.Body.List) != 0 {
						for _, stmt := range fn.Body.List {
							// check if if statement has else branch
							if stmt, ok := stmt.(*ast.IfStmt); ok {
								if stmt.Else != nil {
									pass.Reportf(stmt.Pos(), "if statement should not have an else branch, use early return or switch statement instead")
								}
							}

							// when return statement is found
							if stmt, ok := stmt.(*ast.ReturnStmt); ok {
								// check if return statement has error and no wrapping
								if len(stmt.Results) > 0 {
									for _, result := range stmt.Results[1:] {
										if ident, ok := result.(*ast.Ident); ok && ident.Name == "err" {
											pass.Reportf(stmt.Pos(), "error should not be wrapped")
										}
									}
								}
							}
						}
					}
				}

				return true
			})
		}

		for _, pair := range a2b {
			if pair.First != nil && pair.Second == nil {
				opposite := ""
				switch pair.First.Name {
				case "Request":
					opposite = "Reply"
				case "Send":
					opposite = "Receive"
				case "Publish":
					opposite = "Subscribe"
				case "Set":
					opposite = "Get"
				default:
					opposite = "unknown"
				}
				pair.First.Reporter.report(pair.First.Reporter.pos, "missing %s function", opposite)
			} else if pair.First == nil && pair.Second != nil {
				opposite := ""
				switch pair.Second.Name {
				case "Reply":
					opposite = "Request"
				case "Receive":
					opposite = "Send"
				case "Subscribe":
					opposite = "Publish"
				case "Get":
					opposite = "Set"
				default:
					opposite = "unknown"
				}
				pair.Second.Reporter.report(pair.Second.Reporter.pos, "missing %s function", opposite)
			}
		}

		return nil, nil
	}
}
