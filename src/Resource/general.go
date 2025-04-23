package Resource

import (
	_ "embed"
)

//go:embed icon.ico
var Icon []byte

//go:embed nfapi/sys/tdi/amd64/netfilter2.sys
var TdiAmd64Netfilter2 []byte

//go:embed nfapi/sys/tdi/i386/netfilter2.sys
var TdiI386Netfilter2 []byte

//go:embed nfapi/sys/wfp/amd64/netfilter2.sys
var WfpAmd64Netfilter2 []byte

//go:embed nfapi/sys/wfp/i386/netfilter2.sys
var WfpI386Netfilter2 []byte

//go:embed nfapi/dll/win32/nfapi.dll
var NfapiWin32Nfapi []byte

//go:embed nfapi/dll/x64/nfapi.dll
var NfapiX64Nfapi []byte

//go:embed Proxifier/x64/PrxerDrv.dll
var X64PrxerDrv []byte

//go:embed Proxifier/x32/PrxerDrv.dll
var X32PrxerDrv []byte

//go:embed Proxifier/x64/PrxerNsp.dll
var X64PrxerNsp []byte

//go:embed Proxifier/x32/PrxerNsp.dll
var X32PrxerNsp []byte

//go:embed Proxifier/x64/InstallLSP.exe
var X64InstallLSP []byte

//go:embed Proxifier/x32/InstallLSP.exe
var X32InstallLSP []byte
