package dumbwrap

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type regexList []string

func (r *regexList) String() string {
	return strings.Join(*r, ", ")
}

func (r *regexList) Set(value string) error {
	if _, err := regexp.Compile(value); err != nil {
		return err
	}

	*r = append(*r, value)

	return nil
}

var (
	ignoreExported   bool
	ignoreMethods    bool
	ignoreGenerated  bool
	ignoreSigRegexps regexList
)

var Analyzer = &analysis.Analyzer{
	Name: "dumbwrap",
	Doc:  "report functions that only forward arguments to another function",
	Run:  run,
}

func init() {
	Analyzer.Flags.BoolVar(
		&ignoreExported,
		"ignore-exported",
		false,
		"ignore exported forwarding functions",
	)
	Analyzer.Flags.BoolVar(
		&ignoreMethods,
		"ignore-methods",
		false,
		"ignore forwarding methods",
	)
	Analyzer.Flags.BoolVar(
		&ignoreGenerated,
		"ignore-generated",
		true,
		"ignore generated Go files",
	)
	Analyzer.Flags.Var(
		&ignoreSigRegexps,
		"ignore-sig-regexps",
		"regexp of called function signature to ignore (repeatable)",
	)
}

func run(pass *analysis.Pass) (any, error) {
	var regexps []*regexp.Regexp

	for _, pattern := range ignoreSigRegexps {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid signature regexp %q: %w", pattern, err)
		}

		regexps = append(regexps, re)
	}

	for _, file := range pass.Files {
		if ignoreGenerated && ast.IsGenerated(file) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}

			if fn.Recv == nil && (fn.Name.Name == "main" || fn.Name.Name == "init") {
				return false
			}

			if ignoreExported && fn.Name.IsExported() {
				return false
			}

			if ignoreMethods && fn.Recv != nil {
				return false
			}

			if !isWrapper(pass, fn) {
				return false
			}

			if len(regexps) > 0 {
				call := wrapperCall(fn)
				if call != nil {
					sig := callSignature(pass, call)
					for _, re := range regexps {
						if re.MatchString(sig) {
							return false
						}
					}
				}
			}

			pass.Reportf(fn.Name.Pos(), "function %s only forwards its arguments to another function", fn.Name.Name)

			return false
		})
	}

	return nil, nil
}

func wrapperCall(fn *ast.FuncDecl) *ast.CallExpr {
	switch stmt := fn.Body.List[0].(type) {
	case *ast.ReturnStmt:
		if len(stmt.Results) != 1 {
			return nil
		}

		call, _ := stmt.Results[0].(*ast.CallExpr)

		return call
	case *ast.ExprStmt:
		call, _ := stmt.X.(*ast.CallExpr)
		return call
	}

	return nil
}

func callSignature(pass *analysis.Pass, call *ast.CallExpr) string {
	var obj types.Object

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		obj = pass.TypesInfo.Uses[fun]
	case *ast.SelectorExpr:
		obj = pass.TypesInfo.Uses[fun.Sel]
	}

	fn, ok := obj.(*types.Func)
	if !ok {
		return ""
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return ""
	}

	if recv := sig.Recv(); recv != nil {
		recvType := types.TypeString(recv.Type(), func(pkg *types.Package) string {
			return pkg.Path()
		})
		recvType = strings.TrimPrefix(recvType, "*")

		return recvType + "." + fn.Name() + "("
	}

	if fn.Pkg() != nil {
		return fn.Pkg().Path() + "." + fn.Name() + "("
	}

	return fn.Name() + "("
}

func isWrapper(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if len(fn.Body.List) != 1 {
		return false
	}

	var call *ast.CallExpr

	switch stmt := fn.Body.List[0].(type) {
	case *ast.ReturnStmt:
		if len(stmt.Results) != 1 {
			return false
		}

		call, _ = stmt.Results[0].(*ast.CallExpr)
	case *ast.ExprStmt:
		call, _ = stmt.X.(*ast.CallExpr)
	default:
		return false
	}

	if call == nil {
		return false
	}

	switch call.Fun.(type) {
	case *ast.Ident, *ast.SelectorExpr:
	default:
		return false
	}

	if tv, ok := pass.TypesInfo.Types[call.Fun]; ok && tv.IsType() {
		return false
	}

	if id, ok := call.Fun.(*ast.Ident); ok {
		if pass.TypesInfo.Uses[id] ==
			pass.TypesInfo.Defs[fn.Name] {
			return false
		}
	}

	var params []types.Object

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if len(field.Names) == 0 {
				return false
			}

			for _, name := range field.Names {
				if name.Name == "_" {
					return false
				}

				obj := pass.TypesInfo.Defs[name]
				if obj == nil {
					return false
				}

				params = append(params, obj)
			}
		}
	}

	if len(call.Args) != len(params) {
		return false
	}

	for i, arg := range call.Args {
		id, ok := arg.(*ast.Ident)
		if !ok || pass.TypesInfo.Uses[id] != params[i] {
			return false
		}
	}

	return true
}
