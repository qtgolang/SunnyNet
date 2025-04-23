//go:build !mini
// +build !mini

package GoScriptCode

import (
	"fmt"
	"github.com/qtgolang/SunnyNet/src/Call"
	"github.com/qtgolang/SunnyNet/src/Compress"
	"github.com/qtgolang/SunnyNet/src/GoScriptCode/base"
	"github.com/qtgolang/SunnyNet/src/GoScriptCode/check"
	"github.com/qtgolang/SunnyNet/src/GoScriptCode/yaegi/interp"
	"github.com/qtgolang/SunnyNet/src/GoScriptCode/yaegi/stdlib"
	"github.com/qtgolang/SunnyNet/src/Interface"
	"github.com/qtgolang/SunnyNet/src/RSA"
	"github.com/qtgolang/SunnyNet/src/protobuf"
	"github.com/qtgolang/SunnyNet/src/public"
	"reflect"
	"strconv"
	"strings"
)

type GoScriptTypeHTTP func(Interface.ConnHTTPScriptCall)
type GoScriptTypeWS func(Interface.ConnWebSocketScriptCall)
type GoScriptTypeTCP func(Interface.ConnTCPScriptCall)
type GoScriptTypeUDP func(Interface.ConnUDPScriptCall)

func extractImport(s string) map[string]bool {
	arrayMap := make(map[string]bool)
	str := ""
	start := false
	start2 := false
	for _, v := range s {
		if v == '\n' {
			if strings.HasPrefix(str, "import ") {
				str = strings.ReplaceAll(str, "import ", "")
				arrayMap[strings.TrimSpace(str)] = true
				str = ""
				continue
			}
			if str != "" && start {
				arrayMap[strings.TrimSpace(str)] = true
			}
			str = ""
		} else {
			if start2 && string(v) == "\"" {
				arrayMap[strings.TrimSpace(" \""+str+"\"")] = true
				str = ""
				start2 = false
				continue
			}
			str += string(v)
			if start && str == ")" {
				start = false
			}
			if str == "import (" {
				start = true
				str = ""
				continue
			}
			if str == "import \"" {
				start2 = true
				str = ""
				continue
			}

		}
	}
	return arrayMap
}

func extractCodeBody(s string) string {
	str := ""
	res := ""
	statr1 := false
	statr2 := false
	statr3 := false
	for _, v := range s {
		if v == '\n' {
			if statr3 {
				str = ""
				statr3 = false
				continue
			}
			if strings.HasPrefix(str, "package ") {
				str = ""
				continue
			}
			if strings.HasPrefix(str, "import (") {
				str = ""
				statr1 = true
				continue
			}
			if strings.HasPrefix(str, "import \"") {
				str = ""
				statr2 = true
				continue
			}
			if strings.HasPrefix(str, "import ") {
				str = ""
				statr3 = true
				continue
			}
			if !statr1 && !statr2 && !statr3 {
				res += str + "\n"
				str = ""
			}
		} else {
			if statr1 {
				if string(v) == ")" {
					statr1 = false
					str = ""
					continue
				}
			}
			if statr2 {
				if string(v) == "\"" {
					statr2 = false
					str = ""
					continue
				}
			}
			str += string(v)
		}
	}
	return res
}

var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols = stdlib.Symbols
	for k, v := range base.Symbols {
		Symbols[k] = v
	}
	Symbols["SunnyNet/src/Call/Call"] = map[string]reflect.Value{
		"Call":          reflect.ValueOf(Call.Call),
		"ConnHTTP":      reflect.ValueOf((*Interface.ConnHTTPScriptCall)(nil)),
		"ConnWebSocket": reflect.ValueOf((*Interface.ConnWebSocketScriptCall)(nil)),
		"ConnTCP":       reflect.ValueOf((*Interface.ConnTCPScriptCall)(nil)),
		"ConnUDP":       reflect.ValueOf((*Interface.ConnUDPScriptCall)(nil)),
	}
	Symbols["SunnyNet/src/mmCompress/mmCompress"] = map[string]reflect.Value{
		"DeflateCompress":   reflect.ValueOf(Compress.DeflateCompress),
		"DeflateUnCompress": reflect.ValueOf(Compress.DeflateUnCompress),
		"ZlibUnCompress":    reflect.ValueOf(Compress.ZlibUnCompress),
		"ZlibCompress":      reflect.ValueOf(Compress.ZlibCompress),
		"GzipCompress":      reflect.ValueOf(Compress.GzipCompress),
		"BrUnCompress":      reflect.ValueOf(Compress.BrUnCompress),
		"BrCompress":        reflect.ValueOf(Compress.BrCompress),
		"GzipUnCompress":    reflect.ValueOf(Compress.GzipUnCompress),
		"ZSTDCompress":      reflect.ValueOf(Compress.ZSTDCompress),
		"ZSTDDecompress":    reflect.ValueOf(Compress.ZSTDDecompress),
	}
	Symbols["SunnyNet/src/SunnyProtobuf/SunnyProtobuf"] = map[string]reflect.Value{
		"PbToJson":  reflect.ValueOf(protobuf.ToJson),
		"JsonToPB":  reflect.ValueOf(protobuf.JsonToPB),
		"JsonParse": reflect.ValueOf(protobuf.JsonParse),
	}
	Symbols["github.com/qtgolang/SunnyNet/src/public/public"] = map[string]reflect.Value{
		"Free": reflect.ValueOf(public.Free),
	}
	Symbols["github.com/qtgolang/SunnyNet/src/RSA/RSA"] = map[string]reflect.Value{
		"PubKeyIO": reflect.ValueOf(RSA.PubKeyIO),
	}
	Symbols["reflect/reflect"] = map[string]reflect.Value{
		"TypeOf": reflect.ValueOf(reflect.TypeOf),
		"Func":   reflect.ValueOf(reflect.Func),
	}
	check.Check(Symbols)
}

type LogFuncInterface func(SunnyNetContext int, info ...any)
type SaveFuncInterface func(SunnyNetContext int, code []byte)

func RunCode(SunnyNetContext int, UserScriptCode []byte, log LogFuncInterface) (resError string, h GoScriptTypeHTTP, w GoScriptTypeWS, t GoScriptTypeTCP, u GoScriptTypeUDP) {
	defer func() {
		if p := recover(); p != nil {
			errorSrc := fmt.Sprintf("%v", p)
			errorLine := ""
			_tmp := strings.Split(errorSrc, ":")
			if len(_tmp) >= 1 {
				errorLine = "错误位置:第" + _tmp[0] + "行,这行代码有问题请检查！"
			} else {
				errorLine = "出现了异常:" + errorSrc
			}
			resError = errorLine
			//fmt.Println(resError)
		}
	}()
	var iEval = interp.New(interp.Options{})
	iEval.Use(Symbols)
	ca := ""
	if len(UserScriptCode) < 100 {
		ca = string(DefaultCode) + string(GoFunc)
	} else {
		ca = string(UserScriptCode) + string(GoFunc)
	}
	//检查默认入口
	{
		if !strings.Contains(ca, "func Event_HTTP(Conn HTTPEvent)") {
			return "错误: 默认结构体已被更改,请检查代码", nil, nil, nil, nil
		}
		if !strings.Contains(ca, "func Event_WebSocket(Conn WebSocketEvent)") {
			return "错误: 默认结构体已被更改,请检查代码", nil, nil, nil, nil
		}
		if !strings.Contains(ca, "func Event_TCP(Conn TCPEvent)") {
			return "错误: 默认结构体已被更改,请检查代码", nil, nil, nil, nil
		}
		if !strings.Contains(ca, "func Event_UDP(Conn UDPEvent)") {
			return "错误: 默认结构体已被更改,请检查代码", nil, nil, nil, nil
		}
	}
	//分析出用户编写的脚本中引用的包
	UserImport := extractImport(ca)
	src := string(GoBuiltFuncCode)
	//分析内置函数引用的包
	SystemPort := extractImport(src)
	for k, _ := range SystemPort {
		if UserImport[k] == false {
			_, _ = iEval.Eval("import " + k)
		}
	}
	CodeBody := extractCodeBody(src)
	S := ca + CodeBody

	_, err := iEval.Eval(S)
	if err != nil {
		errorSrc := strings.ReplaceAll(err.Error(), "_.go:", "")
		errorSrc = strings.ReplaceAll(errorSrc, "[]uint8", "[]byte")
		errorLine := ""
		_tmp := strings.Split(errorSrc, ":")
		if len(_tmp) >= 1 {
			errorLine = "错误位置:第" + _tmp[0] + "行,"
		} else {
			errorLine = "错误位置:第 -1 行,"
		}

		ar := strings.Split(errorSrc, "error: unable to find source related to:")
		if len(ar) >= 2 {
			ar1 := strings.Split(ar[0], ": import")
			if len(ar1) >= 2 {
				return errorLine + "找不到引入包 [ " + ar1[1] + " ]", nil, nil, nil, nil
			}
		}
		ar = strings.Split(errorSrc, ":")
		if len(ar) >= 2 {
			like, _ := strconv.Atoi(ar[0])
			like2 := len(strings.Split(string(UserScriptCode), "\n"))
			if like > like2 {
				return "错误: 默认结构体已被更改,请检查代码", nil, nil, nil, nil
			}
		}
		ar = strings.Split(errorSrc, ": expected declaration, found")
		if len(ar) >= 2 {
			return errorLine + "无效的字符 [ " + ar[1] + " ]", nil, nil, nil, nil
		}
		ar = strings.Split(errorSrc, ": expected ';', found")
		if len(ar) >= 2 {
			ar1 := strings.Split(ar[1], " (and")
			if len(ar1) > 1 {
				return errorLine + "无效的字符 [ " + ar1[0] + " ]", nil, nil, nil, nil
			}
			return errorLine + "无效的字符 [ " + ar[1] + " ]", nil, nil, nil, nil
		}
		ar = strings.Split(errorSrc, ": undefined: ")
		if len(ar) >= 2 {
			return errorLine + "未定义的 [ " + ar[1] + " ]", nil, nil, nil, nil
		}
		ar = strings.Split(errorSrc, ": expected operand, found")
		if len(ar) >= 2 {
			return errorLine + "参数不正确 请检查传递的参数", nil, nil, nil, nil
		}
		ar = strings.Split(errorSrc, ": undefined selector: ")
		if len(ar) >= 2 {
			return errorLine + "找不到方法 [ " + ar[1] + " ]", nil, nil, nil, nil
		}
		if strings.Contains(errorSrc, "mismatched types func") {
			ar = strings.Split(errorSrc, " and untyped")
			if len(ar) > 0 {
				s := strings.TrimSpace(ar[len(ar)-1])
				s = strings.ReplaceAll(s, "[]uint8", "[]byte")
				return errorLine + "不能将 方法函数 类型转换为 " + s + " 类型", nil, nil, nil, nil
			}
		}
		if strings.Contains(errorSrc, "cannot use type func") {
			ar = strings.Split(errorSrc, "as type ")
			if len(ar) > 0 {
				s := strings.TrimSpace(ar[len(ar)-1])
				s = strings.ReplaceAll(s, "[]uint8", "[]byte")
				return errorLine + "不能将 方法函数 转换为 " + s + " 类型", nil, nil, nil, nil
			}
		}
		if strings.Contains(errorSrc, " found ')' (and") || strings.Contains(errorSrc, " found '(' (and") {
			return errorLine + "括号不匹配", nil, nil, nil, nil
		}
		if strings.Contains(errorSrc, "assignment") && strings.Contains(errorSrc, "cannot use type") && strings.Contains(errorSrc, "as type func") {
			return errorLine + "赋值错误[不能将值,赋值给方法函数]", nil, nil, nil, nil
		}
		if strings.Contains(errorSrc, "too many arguments") {
			return errorLine + "[ 参数太多 ]", nil, nil, nil, nil
		}

		//invalid operation: mismatched types func() int and untyped int
		ar = strings.Split(errorSrc, ": illegal character ")
		if len(ar) >= 2 {
			ar1 := strings.Split(errorSrc, " ")
			if len(ar1) >= 1 {
				return errorLine + "非法字符 " + ar1[len(ar1)-1] + " ", nil, nil, nil, nil
			}
			return errorLine + "非法字符 " + ar[1], nil, nil, nil, nil
		}
		if strings.Index(errorSrc, ": package ") != -1 && strings.Index(errorSrc, "has no symbol ") != -1 {
			ar = strings.Split(errorSrc, ": package ")
			if len(ar) >= 2 {
				ar = strings.Split(ar[1], " ")
				if len(ar) >= 2 {
					ar = strings.Split(ar[1], " ")
					pack := ar[0]
					ar = strings.Split(errorSrc, "has no symbol ")
					if len(ar) >= 2 {
						funcName := ar[1]
						return errorLine + "在包 " + pack + " 中 找不到函数 -> \"" + funcName + "\"", nil, nil, nil, nil
					}
				}
			}
		}
		return errorLine + "错误信息:" + errorSrc, nil, nil, nil, nil
	}
	v, err := iEval.Eval("main.NewHttpSunny")
	if err != nil {
		return err.Error(), nil, nil, nil, nil
	}
	_httpFunc := v.Interface().(func(Interface.ConnHTTPScriptCall))
	if _httpFunc == nil {
		return "找不到NewHttpSunny", nil, nil, nil, nil
	}
	defer func() {
		if p := recover(); p != nil {
			resError = fmt.Sprintf("%v", p)
		}
	}()
	_httpFunc(nil)
	v, err = iEval.Eval("main.NewWebsocketSunny")
	if err != nil {
		return err.Error(), nil, nil, nil, nil
	}
	_wsFunc := v.Interface().(func(Interface.ConnWebSocketScriptCall))
	if _wsFunc == nil {
		return "找不到NewWebsocketSunny", nil, nil, nil, nil
	}
	defer func() {
		if p := recover(); p != nil {
			resError = fmt.Sprintf("%v", p)
		}
	}()
	_wsFunc(nil)
	v, err = iEval.Eval("main.NewTCPSunny")
	if err != nil {
		return err.Error(), nil, nil, nil, nil
	}
	_tcpFunc := v.Interface().(func(Interface.ConnTCPScriptCall))
	if _tcpFunc == nil {
		return "找不到NewTCPSunnyy", nil, nil, nil, nil
	}
	defer func() {
		if p := recover(); p != nil {
			resError = fmt.Sprintf("%v", p)
		}
	}()
	_tcpFunc(nil)

	v, err = iEval.Eval("main.NewUDPSunny")
	if err != nil {
		return err.Error(), nil, nil, nil, nil
	}
	_udpFunc := v.Interface().(func(Interface.ConnUDPScriptCall))
	if _udpFunc == nil {
		return "找不到NewUDPSunnyy", nil, nil, nil, nil
	}
	defer func() {
		if p := recover(); p != nil {
			resError = fmt.Sprintf("%v", p)
		}
	}()
	_udpFunc(nil)
	v, err = iEval.Eval("main.SetLogFunc")
	if err != nil {
		return err.Error(), nil, nil, nil, nil
	}
	SetLogFunc := v.Interface().(func(func(info ...any)))
	if SetLogFunc == nil {
		return "SetLogFunc", nil, nil, nil, nil
	}
	SetLogFunc(func(info ...any) {
		if log != nil {
			log(SunnyNetContext, info...)
		}
	})
	return "", _httpFunc, _wsFunc, _tcpFunc, _udpFunc
}
