//go:build !mini
// +build !mini

package Interface

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

//go:embed interface.go
var interfaceFile string

func init() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "example.go", interfaceFile, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}
	interfaceMethods := make(map[string][]*ast.Field)
	ast.Inspect(node, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		it, ok := ts.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		interfaceMethods[ts.Name.Name] = it.Methods.List
		return false
	})
	ExportEvent.HTTPScriptEvent = collectMethods(interfaceMethods, "ConnHTTPScriptCall")
	ExportEvent.TCPScriptEvent = collectMethods(interfaceMethods, "ConnTCPScriptCall")
	ExportEvent.UDPScriptEvent = collectMethods(interfaceMethods, "ConnUDPScriptCall")
	ExportEvent.WebSocketScriptEvent = collectMethods(interfaceMethods, "ConnWebSocketScriptCall")

	//测试时使用
	//A, _ := json.Marshal(ExportEvent)
	//os.WriteFile("G:\\Sunny\\SunnyNetV4\\src\\Resource\\Script\\src\\assets\\EventFunc.json", A, 0777)
}
func collectMethods(Methods map[string][]*ast.Field, name string) []EventFunc {
	var eventFuncs []EventFunc
	if methods, found := Methods[name]; found {
		for _, field := range methods {
			comment := ""
			if field.Doc != nil {
				comment = strings.TrimSpace(field.Doc.Text())
			}

			var methodNames []string
			for _, ident := range field.Names {
				methodNames = append(methodNames, ident.Name)
			}

			if len(methodNames) == 0 {
				if ident, ok := field.Type.(*ast.Ident); ok {
					array := collectMethods(Methods, ident.Name)
					for _, v := range array {
						ok = false
						for n, vv := range eventFuncs {
							if vv.Name == ident.Name {
								ok = true
								eventFuncs[n] = v
								break
							}
						}
						if !ok {
							eventFuncs = append(eventFuncs, v)
						}
					}
				}
			} else {
				var args []EventFuncArgs
				var results []string
				if funcType, ok := field.Type.(*ast.FuncType); ok {
					for _, p := range funcType.Params.List {
						typeName := exprToString(p.Type)
						for _, paramName := range p.Names {
							args = append(args, EventFuncArgs{
								Name: paramName.Name,
								Type: typeName,
							})
						}
					}
					if funcType.Results != nil {
						for _, r := range funcType.Results.List {
							typeName := exprToString(r.Type)
							results = append(results, typeName)
						}
					}
				}
				arrComment := strings.Split(strings.TrimSpace(comment), "\n")
				comment = ""
				for x, v := range arrComment {
					if x > 0 {
						if strings.Contains(v, "public.") {
							continue
						}
						if len(strings.TrimSpace(strings.ReplaceAll(v, "\t", ""))) < 1 {
							continue
						}
						if comment == "" {
							comment = v
						} else {
							comment += "\n" + v
						}
					}
				}
				eventFunc := EventFunc{
					Name:    strings.Join(methodNames, ", "),
					Args:    args,
					Returns: results,
					Comment: comment,
				}
				ok := false
				for n, vv := range eventFuncs {
					if vv.Name == eventFunc.Name {
						ok = true
						eventFuncs[n] = eventFunc
						break
					}
				}
				if !ok {
					eventFuncs = append(eventFuncs, eventFunc)
				}
			}
		}
	}
	return eventFuncs
}
func argsReplace(a string) string {
	if !strings.HasPrefix(a, "&") {
		return a
	}
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(a, "&{", ""), "}", ""), " ", ".")
}
func exprToString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + exprToString(t.Sel)
	case *ast.Ellipsis:
		return exprToString(t.Elt)
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	default:
		panic("解析参数错误")
		return argsReplace(fmt.Sprintf("%s", e))
	}
}
