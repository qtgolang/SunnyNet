//go:build !mini
// +build !mini

package GoScriptCode

import _ "embed"

//go:embed BuiltFunc.txt
var GoBuiltFuncCode []byte

//go:embed GoFunc.txt
var GoFunc []byte

//go:embed DefaultCode.txt
var DefaultCode []byte

//go:embed DefaultHTTPCode.txt
var DefaultHTTPCode []byte

//go:embed DefaultWSCode.txt
var DefaultWSCode []byte

//go:embed DefaultTCPCode.txt
var DefaultTCPCode []byte

//go:embed DefaultUDPCode.txt
var DefaultUDPCode []byte
