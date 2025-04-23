/*
本类为所有动态库导出函数集合
*/
package main

import "C"
import (
	"errors"
	"github.com/qtgolang/SunnyNet/Api"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/public"
	"unsafe"
)

/*
GetSunnyVersion 获取SunnyNet版本
*/
//export GetSunnyVersion
func GetSunnyVersion() uintptr {
	return Api.GetSunnyVersion()
}

/*
Free 释放指针
*/
//export Free
func Free(ptr uintptr) {
	public.Free(ptr)
}

/*
CreateSunnyNet 创建Sunny中间件对象,可创建多个
*/
//export CreateSunnyNet
func CreateSunnyNet() int {
	return Api.CreateSunnyNet()
}

/*
ReleaseSunnyNet ReleaseSunnyNet 释放SunnyNet
*/
//export ReleaseSunnyNet
func ReleaseSunnyNet(SunnyContext int) bool {
	return Api.ReleaseSunnyNet(SunnyContext)
}

/*
SunnyNetStart 启动Sunny中间件 成功返回true
*/
//export SunnyNetStart
func SunnyNetStart(SunnyContext int) bool {
	return Api.SunnyNetStart(SunnyContext)
}

/*
SunnyNetSetPort 设置指定端口 Sunny中间件启动之前调用
*/
//export SunnyNetSetPort
func SunnyNetSetPort(SunnyContext, Port int) bool {
	return Api.SunnyNetSetPort(SunnyContext, Port)
}

/*
SunnyNetClose 关闭停止指定Sunny中间件
*/
//export SunnyNetClose
func SunnyNetClose(SunnyContext int) bool {
	return Api.SunnyNetClose(SunnyContext)
}

/*
SunnyNetSetCert 设置自定义证书
*/
//export SunnyNetSetCert
func SunnyNetSetCert(SunnyContext, CertificateManagerId int) bool {
	return Api.SunnyNetSetCert(SunnyContext, CertificateManagerId)
}

/*
SunnyNetInstallCert 安装证书 将证书安装到Windows系统内
*/
//export SunnyNetInstallCert
func SunnyNetInstallCert(SunnyContext int) uintptr {
	return Api.SunnyNetInstallCert(SunnyContext)
}

/*
SunnyNetSetCallback 设置中间件回调地址 httpCallback
*/
//export SunnyNetSetCallback
func SunnyNetSetCallback(SunnyContext, httpCallback, tcpCallback, wsCallback, udpCallback int) bool {
	return Api.SunnyNetSetCallback(SunnyContext, httpCallback, tcpCallback, wsCallback, udpCallback)
}

/*
SunnyNetSocket5AddUser 添加 S5代理需要验证的用户名
*/
//export SunnyNetSocket5AddUser
func SunnyNetSocket5AddUser(SunnyContext int, User, Pass *C.char) bool {
	return Api.SunnyNetSocket5AddUser(SunnyContext, C.GoString(User), C.GoString(Pass))
}

/*
SunnyNetVerifyUser 开启身份验证模式
*/
//export SunnyNetVerifyUser
func SunnyNetVerifyUser(SunnyContext int, open bool) bool {
	return Api.SunnyNetVerifyUser(SunnyContext, open)
}

/*
SunnyNetSocket5DelUser 删除 S5需要验证的用户名
*/
//export SunnyNetSocket5DelUser
func SunnyNetSocket5DelUser(SunnyContext int, User *C.char) bool {
	return Api.SunnyNetSocket5DelUser(SunnyContext, C.GoString(User))
}

/*
SunnyNetGetSocket5User 开启身份验证模式后 获取授权的S5账号,注意UDP请求无法获取到授权的s5账号
*/
//export SunnyNetGetSocket5User
func SunnyNetGetSocket5User(Theology int) uintptr {
	return Api.SunnyNetGetSocket5User(Theology)
}

/*
SunnyNetMustTcp 设置中间件是否开启强制走TCP
*/
//export SunnyNetMustTcp
func SunnyNetMustTcp(SunnyContext int, open bool) {
	Api.SunnyNetMustTcp(SunnyContext, open)
}

/*
CompileProxyRegexp 设置中间件上游代理使用规则
*/
//export CompileProxyRegexp
func CompileProxyRegexp(SunnyContext int, Regexp *C.char) bool {
	return Api.CompileProxyRegexp(SunnyContext, C.GoString(Regexp))
}

/*
SetMustTcpRegexp 设置强制走TCP规则,如果 打开了全部强制走TCP状态,本功能则无效 RulesAllow=false 规则之外走TCP  RulesAllow=true 规则之内走TCP
*/
//export SetMustTcpRegexp
func SetMustTcpRegexp(SunnyContext int, Regexp *C.char, RulesAllow bool) bool {
	return Api.SetMustTcpRegexp(SunnyContext, C.GoString(Regexp), RulesAllow)
}

/*
SunnyNetError 获取中间件启动时的错误信息
*/
//export SunnyNetError
func SunnyNetError(SunnyContext int) uintptr {
	return Api.SunnyNetError(SunnyContext)
}

/*
SetGlobalProxy 设置全局上游代理 仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
*/
//
//export SetGlobalProxy
func SetGlobalProxy(SunnyContext int, ProxyAddress *C.char, outTime int) bool {
	return Api.SetGlobalProxy(SunnyContext, C.GoString(ProxyAddress), outTime)
}

/*
GetRequestProto 获取 HTTPS 请求的协议版本
*/
//export GetRequestProto
func GetRequestProto(MessageId int) uintptr {
	return Api.GetRequestProto(MessageId)
}

/*
GetResponseProto 获取 HTTPS 响应的协议版本
*/
//export GetResponseProto
func GetResponseProto(MessageId int) uintptr {
	return Api.GetResponseProto(MessageId)
}

/*
ExportCert 导出已设置的证书
*/
//export ExportCert
func ExportCert(SunnyContext int) uintptr {
	return Api.ExportCert(SunnyContext)
}

/*
SetHTTPRequestMaxUpdateLength 设置HTTP请求,提交数据,最大的长度
*/
//export SetHTTPRequestMaxUpdateLength
func SetHTTPRequestMaxUpdateLength(SunnyContext int, i int64) bool {
	return Api.SetHTTPRequestMaxUpdateLength(SunnyContext, i)
}

/*
SetIeProxy 设置IE代理 ，Windows 有效
*/
//export SetIeProxy
func SetIeProxy(SunnyContext int) bool {
	return Api.SetIeProxy(SunnyContext)
}

/*
CancelIEProxy  取消设置的IE代理，Windows 有效
*/
//export CancelIEProxy
func CancelIEProxy(SunnyContext int) bool {
	return Api.CancelIEProxy(SunnyContext)
}

/*
SetRequestCookie 修改、设置 HTTP/S当前请求数据中指定Cookie
*/
//export SetRequestCookie
func SetRequestCookie(MessageId int, name, val *C.char) {
	Api.SetRequestCookie(MessageId, C.GoString(name), C.GoString(val))
}

/*
SetRequestAllCookie 修改、设置 HTTP/S当前请求数据中的全部Cookie
*/
//export SetRequestAllCookie
func SetRequestAllCookie(MessageId int, val *C.char) {
	Api.SetRequestAllCookie(MessageId, C.GoString(val))
}

/*
GetRequestCookie 获取 HTTP/S当前请求数据中指定的Cookie
*/
//export GetRequestCookie
func GetRequestCookie(MessageId int, name *C.char) uintptr {
	r := Api.GetRequestCookie(MessageId, C.GoString(name))
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
GetRequestALLCookie 获取 HTTP/S 当前请求全部Cookie
*/
//export GetRequestALLCookie
func GetRequestALLCookie(MessageId int) uintptr {
	r := Api.GetRequestALLCookie(MessageId)
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
DelResponseHeader 删除HTTP/S返回数据中指定的协议头
*/
//export DelResponseHeader
func DelResponseHeader(MessageId int, name *C.char) {
	Api.DelResponseHeader(MessageId, C.GoString(name))
}

/*
DelRequestHeader 删除HTTP/S请求数据中指定的协议头
*/
//export DelRequestHeader
func DelRequestHeader(MessageId int, name *C.char) {
	Api.DelRequestHeader(MessageId, C.GoString(name))
}

/*
SetRequestOutTime 请求设置超时-毫秒
*/
//export SetRequestOutTime
func SetRequestOutTime(MessageId int, times int) {
	Api.SetRequestOutTime(MessageId, times)
}

/*
SetRequestALLHeader SetRequestALLHeader 设置HTTP/ S请求体中的全部协议头
*/
//export SetRequestALLHeader
func SetRequestALLHeader(MessageId int, val *C.char) {
	Api.SetRequestALLHeader(MessageId, C.GoString(val))
}

/*
SetRequestHeader 设置HTTP/S请求体中的协议头
*/
//export SetRequestHeader
func SetRequestHeader(MessageId int, name, val *C.char) {
	Api.SetRequestHeader(MessageId, C.GoString(name), C.GoString(val))
}

/*
RandomRequestCipherSuites RandomRequestCipherSuites 随机设置请求 CipherSuites
*/
//export RandomRequestCipherSuites
func RandomRequestCipherSuites(MessageId int) bool {
	return Api.SetRequestCipherSuites(MessageId)
}

/*
SetRequestHTTP2Config  设置HTTP 2.0 请求指纹配置 (若服务器支持则使用,若服务器不支持,设置了也不会使用)
*/
//export SetRequestHTTP2Config
func SetRequestHTTP2Config(MessageId int, h2Config *C.char) bool {
	return Api.SetRequestHTTP2Config(MessageId, C.GoString(h2Config))
}

/*
SetResponseHeader 修改、设置 HTTP/S当前返回数据中的指定协议头
*/
//export SetResponseHeader
func SetResponseHeader(MessageId int, name *C.char, val *C.char) {
	Api.SetResponseHeader(MessageId, C.GoString(name), C.GoString(val))
}

/*
GetRequestHeader 获取 HTTP/S当前请求数据中的指定协议头
*/
//export GetRequestHeader
func GetRequestHeader(MessageId int, name *C.char) uintptr {
	r := Api.GetRequestHeader(MessageId, C.GoString(name))
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
GetResponseHeader 获取 HTTP/S 当前返回数据中指定的协议头
*/
//export GetResponseHeader
func GetResponseHeader(MessageId int, name *C.char) uintptr {
	r := Api.GetResponseHeader(MessageId, C.GoString(name))
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
GetResponseServerAddress 获取 HTTP/S 相应的服务器地址
*/
//export GetResponseServerAddress
func GetResponseServerAddress(MessageId int) uintptr {
	r := Api.GetResponseServerAddress(MessageId)
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
SetResponseAllHeader 修改、设置 HTTP/S当前返回数据中的全部协议头，例如设置返回两条Cookie 使用本命令设置 使用设置、修改 单条命令无效
*/
//export SetResponseAllHeader
func SetResponseAllHeader(MessageId int, value *C.char) {
	Api.SetResponseAllHeader(MessageId, C.GoString(value))
}

/*
GetResponseAllHeader 获取 HTTP/S 当前返回全部协议头
*/
//export GetResponseAllHeader
func GetResponseAllHeader(MessageId int) uintptr {
	r := Api.GetResponseAllHeader(MessageId)
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
GetRequestAllHeader 获取 HTTP/S 当前请求数据全部协议头
*/
//export GetRequestAllHeader
func GetRequestAllHeader(MessageId int) uintptr {
	r := Api.GetRequestAllHeader(MessageId)
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
SetRequestProxy 设置HTTP/S请求代理，仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
*/
//
//export SetRequestProxy
func SetRequestProxy(MessageId int, ProxyUrl *C.char, outTime int) bool {
	return Api.SetRequestProxy(MessageId, C.GoString(ProxyUrl), outTime)
}

/*
GetResponseStatusCode 获取HTTP/S返回的状态码
*/
//export GetResponseStatusCode
func GetResponseStatusCode(MessageId int) int {
	return Api.GetResponseStatusCode(MessageId)
}

/*
GetRequestClientIp 获取当前HTTP/S请求由哪个IP发起
*/
//export GetRequestClientIp
func GetRequestClientIp(MessageId int) uintptr {
	r := Api.GetRequestClientIp(MessageId)
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
GetResponseStatus 获取HTTP/S返回的状态文本 例如 [200 OK]
*/
//export GetResponseStatus
func GetResponseStatus(MessageId int) uintptr {
	r := Api.GetResponseStatus(MessageId)
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
SetResponseStatus 修改HTTP/S返回的状态码
*/
//export SetResponseStatus
func SetResponseStatus(MessageId, code int) {
	Api.SetResponseStatus(MessageId, code)
}

/*
SetRequestUrl 修改HTTP/S当前请求的URL
*/
//export SetRequestUrl
func SetRequestUrl(MessageId int, URI *C.char) bool {
	return Api.SetRequestUrl(MessageId, C.GoString(URI))
}

/*
GetRequestBodyLen 获取 HTTP/S 当前请求POST提交数据长度
*/
//export GetRequestBodyLen
func GetRequestBodyLen(MessageId int) int {
	return Api.GetRequestBodyLen(MessageId)
}

/*
GetResponseBodyLen 获取 HTTP/S 当前返回  数据长度
*/
//export GetResponseBodyLen
func GetResponseBodyLen(MessageId int) int {
	return Api.GetResponseBodyLen(MessageId)
}

/*
SetResponseData 设置、修改 HTTP/S 当前请求返回数据 如果再发起请求时调用本命令，请求将不会被发送，将会直接返回 data=数据指针  dataLen=数据长度
*/
//export SetResponseData
func SetResponseData(MessageId int, data uintptr, dataLen int) bool {
	return Api.SetResponseData(MessageId, public.CStringToBytes(data, dataLen))
}

/*
SetRequestData 设置、修改 HTTP/S 当前请求POST提交数据  data=数据指针  dataLen=数据长度
*/
//export SetRequestData
func SetRequestData(MessageId int, data uintptr, dataLen int) bool {
	return Api.SetRequestData(MessageId, public.CStringToBytes(data, dataLen))
}

/*
GetRequestBody 获取 HTTP/S 当前POST提交数据 返回 数据指针
*/
//export GetRequestBody
func GetRequestBody(MessageId int) uintptr {
	bs := Api.GetRequestBody(MessageId)
	if bs == nil {
		return 0
	}
	return public.PointerPtr(bs)
}

/*
IsRequestRawBody 此请求是否为原始body 如果是 将无法修改提交的Body，请使用 RawRequestDataToFile 命令来储存到文件
*/
//export IsRequestRawBody
func IsRequestRawBody(MessageId int) bool {
	return Api.IsRequestRawBody(MessageId)
}

/*
RawRequestDataToFile 获取 HTTP/ S 当前POST提交数据原始Data,传入保存文件名路径,例如"c:\1.txt"
*/
//export RawRequestDataToFile
func RawRequestDataToFile(MessageId int, saveFileName uintptr, len int) bool {
	return Api.RawRequestDataToFile(MessageId, string(public.CStringToBytes(saveFileName, len)))
}

/*
GetResponseBody 获取 HTTP/S 当前返回数据  返回 数据指针
*/
//export GetResponseBody
func GetResponseBody(MessageId int) uintptr {
	bs := Api.GetResponseBody(MessageId)
	if bs == nil {
		return 0
	}
	return public.PointerPtr(bs)
}

/*
GetWebsocketBodyLen 获取 WebSocket消息长度
*/
//export GetWebsocketBodyLen
func GetWebsocketBodyLen(MessageId int) int {
	return Api.GetWebsocketBodyLen(MessageId)
}

/*
CloseWebsocket 主动关闭Websocket
*/
//export CloseWebsocket
func CloseWebsocket(Theology int) bool {
	return Api.CloseWebsocket(Theology)
}

/*
GetWebsocketBody 获取 WebSocket消息 返回数据指针
*/
//export GetWebsocketBody
func GetWebsocketBody(MessageId int) uintptr {
	bs := Api.GetWebsocketBody(MessageId)
	if bs == nil {
		return 0
	}
	return public.PointerPtr(bs)
}

/*
SetWebsocketBody 修改 WebSocket消息 data=数据指针  dataLen=数据长度
*/
//export SetWebsocketBody
func SetWebsocketBody(MessageId int, data uintptr, dataLen int) bool {
	return Api.SetWebsocketBody(MessageId, public.CStringToBytes(data, dataLen))
}

/*
SendWebsocketBody 主动向Websocket服务器发送消息 MessageType=WS消息类型 data=数据指针  dataLen=数据长度
*/
//export SendWebsocketBody
func SendWebsocketBody(Theology, MessageType int, data uintptr, dataLen int) bool {
	bs := public.CStringToBytes(data, dataLen)
	return Api.SendWebsocketBody(Theology, MessageType, bs)
}

/*
SendWebsocketClientBody SendWebsocketClientBody 主动向Websocket客户端发送消息 MessageType=WS消息类型 data=数据指针  dataLen=数据长度
*/
//export SendWebsocketClientBody
func SendWebsocketClientBody(Theology, MessageType int, data uintptr, dataLen int) bool {
	bs := public.CStringToBytes(data, dataLen)
	return Api.SendWebsocketClientBody(Theology, MessageType, bs)
}

/*
SetTcpBody 修改 TCP消息数据 MsgType=1 发送的消息 MsgType=2 接收的消息 如果 MsgType和MessageId不匹配，将不会执行操作  data=数据指针  dataLen=数据长度
*/
//export SetTcpBody
func SetTcpBody(MessageId, MsgType int, data uintptr, dataLen int) bool {
	return Api.SetTcpBody(MessageId, MsgType, public.CStringToBytes(data, dataLen))
}

/*
SetTcpAgent 给当前TCP连接设置代理 仅限 TCP回调 即将连接时使用 仅支持S5代理 例如 socket5://admin:123456@127.0.0.1:8888
*/
//
//export SetTcpAgent
func SetTcpAgent(MessageId int, ProxyUrl *C.char, outTime int) bool {
	return Api.SetTcpAgent(MessageId, C.GoString(ProxyUrl), outTime)
}

/*
TcpCloseClient 根据唯一ID关闭指定的TCP连接  唯一ID在回调参数中
*/
//export TcpCloseClient
func TcpCloseClient(theology int) bool {
	return Api.TcpCloseClient(theology)
}

/*
SetTcpConnectionIP 给指定的TCP连接 修改目标连接地址 目标地址必须带端口号 例如 baidu.com:443
*/
//export SetTcpConnectionIP
func SetTcpConnectionIP(MessageId int, address *C.char) bool {
	return Api.SetTcpConnectionIP(MessageId, C.GoString(address))
}

/*
TcpSendMsg 指定的TCP连接 模拟客户端向服务器端主动发送数据
*/
//export TcpSendMsg
func TcpSendMsg(theology int, data uintptr, dataLen int) int {
	return Api.TcpSendMsg(theology, public.CStringToBytes(data, dataLen))
}

/*
TcpSendMsgClient 指定的TCP连接 模拟服务器端向客户端主动发送数据
*/
//export TcpSendMsgClient
func TcpSendMsgClient(theology int, data uintptr, dataLen int) int {
	return Api.TcpSendMsgClient(theology, public.CStringToBytes(data, dataLen))
}

/*
BytesToInt 将Go int的Bytes 转为int
*/
//export BytesToInt
func BytesToInt(data uintptr, dataLen int) int {
	return Api.BytesToInt(data, dataLen)
}

/*
GzipUnCompress Gzip解压缩
*/
//export GzipUnCompress
func GzipUnCompress(data uintptr, dataLen int) uintptr {
	return Api.GzipUnCompress(data, dataLen)
}

/*
BrUnCompress br解压缩
*/
//export BrUnCompress
func BrUnCompress(data uintptr, dataLen int) uintptr {
	return Api.BrUnCompress(data, dataLen)
}

/*
BrCompress br压缩
*/
//export BrCompress
func BrCompress(data uintptr, dataLen int) uintptr {
	return Api.BrCompress(data, dataLen)
}

/*
ZSTDDecompress ZSTD解压缩
*/
//export ZSTDDecompress
func ZSTDDecompress(data uintptr, dataLen int) uintptr {
	return Api.ZSTDDecompress(data, dataLen)
}

/*
ZSTDCompress ZSTD压缩
*/
//export ZSTDCompress
func ZSTDCompress(data uintptr, dataLen int) uintptr {
	return Api.ZSTDCompress(data, dataLen)
}

/*
BrCompress br压缩
*/
//export BrotliCompress
func BrotliCompress(data uintptr, dataLen int) uintptr {
	return Api.BrCompress(data, dataLen)
}

/*
GzipCompress Gzip压缩
*/
//export GzipCompress
func GzipCompress(data uintptr, dataLen int) uintptr {
	return Api.GzipCompress(data, dataLen)
}

/*
ZlibCompress Zlib压缩
*/
//export ZlibCompress
func ZlibCompress(data uintptr, dataLen int) uintptr {
	return Api.ZlibCompress(data, dataLen)
}

/*
ZlibUnCompress Zlib解压缩
*/
//export ZlibUnCompress
func ZlibUnCompress(data uintptr, dataLen int) uintptr {
	return Api.ZlibUnCompress(data, dataLen)
}

/*
DeflateUnCompress Deflate解压缩 (可能等同于zlib解压缩)
*/
//export DeflateUnCompress
func DeflateUnCompress(data uintptr, dataLen int) uintptr {
	return Api.DeflateUnCompress(data, dataLen)
}

/*
DeflateCompress Deflate压缩 (可能等同于zlib压缩)
*/
//export DeflateCompress
func DeflateCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	bx := Api.DeflateCompress(bin)
	if bx == nil {
		return 0
	}
	bx = public.BytesCombine(public.IntToBytes(len(bx)), bx)
	return public.PointerPtr(string(bx))
}

/*
WebpToJpegBytes Webp图片转JEG图片字节数组 SaveQuality=质量(默认75)
*/
//export WebpToJpegBytes
func WebpToJpegBytes(data uintptr, dataLen int, SaveQuality int) uintptr {
	_webp := public.CStringToBytes(data, dataLen)
	bs := Api.WebpToJpegBytes(_webp, SaveQuality)
	bn := public.BytesCombine(public.IntToBytes(len(bs)), bs)
	return public.PointerPtr(string(bn))
}

/*
WebpToPngBytes Webp图片转Png图片字节数组
*/
//export WebpToPngBytes
func WebpToPngBytes(data uintptr, dataLen int) uintptr {
	_webp := public.CStringToBytes(data, dataLen)
	bs := Api.WebpToPngBytes(_webp)
	if bs == nil {
		return 0
	}
	bn := public.BytesCombine(public.IntToBytes(len(bs)), bs)
	return public.PointerPtr(string(bn))
}

/*
WebpToJpeg Webp图片转JEG图片 根据文件名 SaveQuality=质量(默认75)
*/
//export WebpToJpeg
func WebpToJpeg(webpPath, savePath *C.char, SaveQuality int) bool {
	return Api.WebpToJpeg(C.GoString(webpPath), C.GoString(savePath), SaveQuality)
}

/*
WebpToPng Webp图片转Png图片 根据文件名
*/
//export WebpToPng
func WebpToPng(webpPath, savePath *C.char) bool {
	return Api.WebpToPng(C.GoString(webpPath), C.GoString(savePath))
}

/*
OpenDrive 开始进程代理/打开驱动 只允许一个 SunnyNet 使用 [会自动安装所需驱动文件]
IsNfapi 如果为true表示使用NFAPI驱动 如果为false 表示使用Proxifier
*/
//export OpenDrive
func OpenDrive(SunnyContext int, isNf bool) bool {
	return Api.OpenDrive(SunnyContext, isNf)
}

/*
UnDrive 卸载驱动，仅Windows 有效【需要管理权限】执行成功后会立即重启系统,若函数执行后没有重启系统表示没有管理员权限
*/
//export UnDrive
func UnDrive(SunnyContext int) {
	Api.UnDrive(SunnyContext)
}

/*
ProcessAddName 进程代理 添加进程名
*/
//export ProcessAddName
func ProcessAddName(SunnyContext int, Name *C.char) {
	Api.ProcessAddName(SunnyContext, C.GoString(Name))
}

/*
ProcessDelName 进程代理 删除进程名
*/
//export ProcessDelName
func ProcessDelName(SunnyContext int, Name *C.char) {
	Api.ProcessDelName(SunnyContext, C.GoString(Name))
}

/*
ProcessAddPid 进程代理 添加PID
*/
//export ProcessAddPid
func ProcessAddPid(SunnyContext, pid int) {
	Api.ProcessAddPid(SunnyContext, pid)
}

/*
ProcessDelPid 进程代理 删除PID
*/
//export ProcessDelPid
func ProcessDelPid(SunnyContext, pid int) {
	Api.ProcessDelPid(SunnyContext, pid)
}

/*
ProcessCancelAll 进程代理 取消全部已设置的进程名
*/
//export ProcessCancelAll
func ProcessCancelAll(SunnyContext int) {
	Api.ProcessCancelAll(SunnyContext)
}

/*
ProcessALLName 进程代理 设置是否全部进程通过
*/
//export ProcessALLName
func ProcessALLName(SunnyContext int, open, StopNetwork bool) {
	Api.ProcessALLName(SunnyContext, open, StopNetwork)
}

//================================================================================================

/*
GetCommonName 证书管理器 获取证书 CommonName 字段
*/
//export GetCommonName
func GetCommonName(Context int) uintptr {
	return public.PointerPtr(Api.GetCommonName(Context))
}

/*
ExportP12 证书管理器 导出为P12
*/
//export ExportP12
func ExportP12(Context int, path, pass *C.char) bool {
	return Api.ExportP12(Context, C.GoString(path), C.GoString(pass))
}

/*
ExportPub 证书管理器 导出公钥
*/
//export ExportPub
func ExportPub(Context int) uintptr {
	p := Api.ExportPub(Context)
	if p == "" {
		return 0
	}
	return public.PointerPtr(p)
}

/*
ExportKEY 证书管理器 导出私钥
*/
//export ExportKEY
func ExportKEY(Context int) uintptr {
	return public.PointerPtr(Api.ExportKEY(Context))
}

/*
ExportCA 证书管理器 导出证书
*/
//export ExportCA
func ExportCA(Context int) uintptr {
	return public.PointerPtr(Api.ExportCA(Context))
}

/*
CreateCA 证书管理器 创建证书
*/
//export CreateCA
func CreateCA(Context int, Country, Organization, OrganizationalUnit, Province, CommonName, Locality *C.char, bits, NotAfter int) bool {
	return Api.CreateCA(Context, C.GoString(Country), C.GoString(Organization), C.GoString(OrganizationalUnit), C.GoString(Province), C.GoString(CommonName), C.GoString(Locality), bits, NotAfter)
}

/*
AddClientAuth 证书管理器 设置ClientAuth
*/
//export AddClientAuth
func AddClientAuth(Context, val int) bool {
	return Api.AddClientAuth(Context, val)
}

/*
SetCipherSuites SetCipherSuites 证书管理器 设置CipherSuites
*/
//export SetCipherSuites
func SetCipherSuites(Context int, val *C.char) bool {
	return Api.SetCipherSuites(Context, C.GoString(val))
}

/*
AddCertPoolText 证书管理器 设置信任的证书 从 文本
*/
//export AddCertPoolText
func AddCertPoolText(Context int, cer *C.char) bool {
	return Api.AddCertPoolText(Context, C.GoString(cer))
}

/*
AddCertPoolPath 证书管理器 设置信任的证书 从 文件
*/
//export AddCertPoolPath
func AddCertPoolPath(Context int, cer *C.char) bool {
	return Api.AddCertPoolPath(Context, C.GoString(cer))
}

/*
GetServerName 证书管理器 取ServerName
*/
//export GetServerName
func GetServerName(Context int) uintptr {
	return public.PointerPtr(Api.GetServerName(Context))
}

/*
SetServerName 证书管理器 设置ServerName
*/
//export SetServerName
func SetServerName(Context int, name *C.char) bool {
	return Api.SetServerName(Context, C.GoString(name))
}

/*
SetInsecureSkipVerify 证书管理器 设置跳过主机验证
*/
//export SetInsecureSkipVerify
func SetInsecureSkipVerify(Context int, b bool) bool {
	return Api.SetInsecureSkipVerify(Context, b)
}

/*
LoadX509Certificate 证书管理器 载入X509证书
*/
//export LoadX509Certificate
func LoadX509Certificate(Context int, Host, CA, KEY *C.char) bool {
	return Api.LoadX509Certificate(Context, C.GoString(Host), C.GoString(CA), C.GoString(KEY))
}

/*
LoadX509KeyPair 证书管理器 载入X509证书2
*/
//export LoadX509KeyPair
func LoadX509KeyPair(Context int, CaPath, KeyPath *C.char) bool {
	return Api.LoadX509KeyPair(Context, C.GoString(CaPath), C.GoString(KeyPath))
}

/*
LoadP12Certificate 证书管理器 载入p12证书
*/
//export LoadP12Certificate
func LoadP12Certificate(Context int, Name, Password *C.char) bool {
	return Api.LoadP12Certificate(Context, C.GoString(Name), C.GoString(Password))
}

/*
RemoveCertificate 释放 证书管理器 对象
*/
//export RemoveCertificate
func RemoveCertificate(Context int) {
	Api.RemoveCertificate(Context)
}

/*
CreateCertificate 创建 证书管理器 对象
*/
//export CreateCertificate
func CreateCertificate() int {
	return Api.CreateCertificate()
}

//================================================ go map 相关 ==========================================================

/*
KeysWriteStr GoMap 写字符串
*/
//export KeysWriteStr
func KeysWriteStr(KeysHandle int, name *C.char, val uintptr, len int) {
	Api.KeysWriteStr(KeysHandle, C.GoString(name), val, len)
}

/*
KeysGetJson GoMap 转为JSON字符串
*/
//export KeysGetJson
func KeysGetJson(KeysHandle int) uintptr {
	return Api.KeysGetJson(KeysHandle)
}

/*
KeysGetCount GoMap 取数量
*/
//export KeysGetCount
func KeysGetCount(KeysHandle int) int {
	return Api.KeysGetCount(KeysHandle)
}

/*
KeysEmpty GoMap 清空
*/
//export KeysEmpty
func KeysEmpty(KeysHandle int) {
	Api.KeysEmpty(KeysHandle)
}

/*
KeysReadInt GoMap 读整数
*/
//export KeysReadInt
func KeysReadInt(KeysHandle int, name *C.char) int {
	return Api.KeysReadInt(KeysHandle, C.GoString(name))
}

/*
KeysWriteInt GoMap 写整数
*/
//export KeysWriteInt
func KeysWriteInt(KeysHandle int, name *C.char, val int) {
	Api.KeysWriteInt(KeysHandle, C.GoString(name), val)
}

/*
KeysReadLong GoMap 读长整数
*/
//export KeysReadLong
func KeysReadLong(KeysHandle int, name *C.char) int64 {
	return Api.KeysReadLong(KeysHandle, C.GoString(name))
}

/*
KeysWriteLong GoMap 写长整数
*/
//export KeysWriteLong
func KeysWriteLong(KeysHandle int, name *C.char, val int64) {
	Api.KeysWriteLong(KeysHandle, C.GoString(name), val)
}

/*
KeysReadFloat GoMap 读浮点数
*/
//export KeysReadFloat
func KeysReadFloat(KeysHandle int, name *C.char) float64 {
	return Api.KeysReadFloat(KeysHandle, C.GoString(name))
}

/*
KeysWriteFloat GoMap 写浮点数
*/
//export KeysWriteFloat
func KeysWriteFloat(KeysHandle int, name *C.char, val float64) {
	Api.KeysWriteFloat(KeysHandle, C.GoString(name), val)
}

/*
KeysWrite GoMap 写字节数组
*/
//export KeysWrite
func KeysWrite(KeysHandle int, name *C.char, val uintptr, length int) {
	Api.KeysWrite(KeysHandle, C.GoString(name), val, length)
}

/*
KeysRead GoMap 写读字符串/字节数组
*/
//export KeysRead
func KeysRead(KeysHandle int, name *C.char) uintptr {
	return Api.KeysRead(KeysHandle, C.GoString(name))
}

/*
KeysDelete GoMap 删除
*/
//export KeysDelete
func KeysDelete(KeysHandle int, name *C.char) {
	Api.KeysDelete(KeysHandle, C.GoString(name))
}

/*
RemoveKeys GoMap 删除GoMap
*/
//export RemoveKeys
func RemoveKeys(KeysHandle int) {
	Api.RemoveKeys(KeysHandle)
}

/*
CreateKeys GoMap 创建
*/
//export CreateKeys
func CreateKeys() int {
	return Api.CreateKeys()
}

//===================================================== go http Client ================================================

/*
HTTPSetH2Config HTTP 客户端 设置HTTP2指纹
*/
//export HTTPSetH2Config
func HTTPSetH2Config(Context int, config *C.char) bool {
	return Api.SetH2Config(Context, C.GoString(config))
}

/*
HTTPSetRandomTLS HTTP 客户端 设置随机使用TLS指纹
*/
//export HTTPSetRandomTLS
func HTTPSetRandomTLS(Context int, RandomTLS bool) bool {
	return Api.HTTPSetRandomTLS(Context, RandomTLS)
}

/*
HTTPSetRedirect HTTP 客户端 设置重定向
*/
//export HTTPSetRedirect
func HTTPSetRedirect(Context int, Redirect bool) bool {
	return Api.HTTPSetRedirect(Context, Redirect)
}

/*
HTTPGetCode HTTP 客户端 返回响应状态码
*/
//export HTTPGetCode
func HTTPGetCode(Context int) int {
	return Api.HTTPGetCode(Context)
}

/*
HTTPSetCertManager HTTP 客户端 设置证书管理器
*/
//export HTTPSetCertManager
func HTTPSetCertManager(Context, CertManagerContext int) bool {
	return Api.HTTPSetCertManager(Context, CertManagerContext)
}

/*
HTTPGetBody HTTP 客户端 返回响应内容
*/
//export HTTPGetBody
func HTTPGetBody(Context int) uintptr {
	r := Api.HTTPGetBody(Context)
	if r == nil {
		return 0
	}
	return public.PointerPtr(r)
}

/*
HTTPGetHeader HTTP 客户端 返回响应HTTPGetHeader
*/
//export HTTPGetHeader
func HTTPGetHeader(Context int, name *C.char) uintptr {
	s := Api.HTTPGetHeader(Context, C.GoString(name))
	if s == "" {
		return 0
	}
	return public.PointerPtr(s)
}

/*
HTTPGetRequestHeader HTTP 客户端 添加的全部协议头
*/
//export HTTPGetRequestHeader
func HTTPGetRequestHeader(Context int) uintptr {
	s := Api.HTTPGetRequestHeader(Context)
	if s == "" {
		return 0
	}
	return public.PointerPtr(s)
}

/*
HTTPGetHeads HTTP 客户端 返回响应全部Heads
*/
//export HTTPGetHeads
func HTTPGetHeads(Context int) uintptr {
	r := Api.HTTPGetHeads(Context)
	if r == "" {
		return 0
	}
	return public.PointerPtr(r)
}

/*
HTTPGetBodyLen HTTP 客户端 返回响应长度
*/
//export HTTPGetBodyLen
func HTTPGetBodyLen(Context int) int {
	return Api.HTTPGetBodyLen(Context)
}

/*
HTTPSendBin HTTP 客户端 发送Body
*/
//export HTTPSendBin
func HTTPSendBin(Context int, body uintptr, bodyLength int) {
	Api.HTTPSendBin(Context, public.CStringToBytes(body, bodyLength))
}

/*
HTTPSetTimeouts HTTP 客户端 设置超时 毫秒
*/
//export HTTPSetTimeouts
func HTTPSetTimeouts(Context int, t1 int) {
	Api.HTTPSetTimeouts(Context, t1)
}

// HTTPSetServerIP
// HTTP 客户端 设置真实连接IP地址，
//
//export HTTPSetServerIP
func HTTPSetServerIP(Context int, ServerIP *C.char) {
	Api.HTTPSetServerIP(Context, C.GoString(ServerIP))
}

/*
HTTPSetProxyIP HTTP 客户端 设置代理IP 仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
*/
//
//export HTTPSetProxyIP
func HTTPSetProxyIP(Context int, ProxyUrl *C.char) bool {
	return Api.HTTPSetProxyIP(Context, C.GoString(ProxyUrl))
}

/*
HTTPSetHeader HTTP 客户端 设置协议头
*/
//export HTTPSetHeader
func HTTPSetHeader(Context int, name, value *C.char) {
	Api.HTTPSetHeader(Context, C.GoString(name), C.GoString(value))
}

/*
HTTPOpen HTTP 客户端 Open
*/
//export HTTPOpen
func HTTPOpen(Context int, Method, URL *C.char) {
	Api.HTTPOpen(Context, C.GoString(Method), C.GoString(URL))
}

/*
RemoveHTTPClient 释放 HTTP客户端
*/
//export RemoveHTTPClient
func RemoveHTTPClient(Context int) {
	Api.RemoveHTTPClient(Context)
}

/*
CreateHTTPClient 创建 HTTP 客户端
*/
//export CreateHTTPClient
func CreateHTTPClient() int {
	return Api.CreateHTTPClient()
}

//===========================================================================================

/*
JsonToPB JSON格式的protobuf数据转为protobuf二进制数据
*/
//export JsonToPB
func JsonToPB(bin uintptr, binLen int) uintptr {
	b := Api.JsonToPB(string(public.CStringToBytes(bin, binLen)))
	if len(b) < 1 {
		return 0
	}
	c := public.BytesCombine(public.Int64ToBytes(int64(len(b))), b)
	return public.PointerPtr(c)
}

/*
PbToJson protobuf数据转为JSON格式
*/
//export PbToJson
func PbToJson(bin uintptr, binLen int) uintptr {
	n := C.CString(Api.PbToJson(public.CStringToBytes(bin, binLen)))
	return uintptr(unsafe.Pointer(n))
}

//===========================================================================================

/*
QueuePull 队列弹出
*/
//export QueuePull
func QueuePull(name *C.char) uintptr {
	bx := Api.QueuePull(C.GoString(name))
	if bx == nil {
		return 0
	}
	return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(bx)), bx))
}

/*
QueuePush 加入队列
*/
//export QueuePush
func QueuePush(name *C.char, val uintptr, valLen int) {
	Api.QueuePush(C.GoString(name), public.CStringToBytes(val, valLen))
}

/*
QueueLength 取队列长度
*/
//export QueueLength
func QueueLength(name *C.char) int {
	return Api.QueueLength(C.GoString(name))
}

/*
QueueRelease 清空销毁队列
*/
//export QueueRelease
func QueueRelease(name *C.char) {
	Api.QueueRelease(C.GoString(name))
}

/*
QueueIsEmpty 队列是否为空
*/
//export QueueIsEmpty
func QueueIsEmpty(name *C.char) bool {
	return Api.QueueIsEmpty(C.GoString(name))
}

/*
CreateQueue 创建队列
*/
//export CreateQueue
func CreateQueue(name *C.char) {
	Api.CreateQueue(C.GoString(name))
}

//=========================================================================================================

/*
SocketClientWrite TCP客户端 发送数据
*/
//export SocketClientWrite
func SocketClientWrite(Context, OutTimes int, val uintptr, valLen int) int {
	data := public.CStringToBytes(val, valLen)
	return Api.SocketClientWrite(Context, OutTimes, data)
}

/*
SocketClientClose TCP客户端 断开连接
*/
//export SocketClientClose
func SocketClientClose(Context int) {
	Api.SocketClientClose(Context)
}

/*
SocketClientReceive TCP客户端 同步模式下 接收数据
*/
//export SocketClientReceive
func SocketClientReceive(Context, OutTimes int) uintptr {
	bs := Api.SocketClientReceive(Context, OutTimes)
	if bs == nil {
		return 0
	}
	return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(bs)), bs))
}

/*
SocketClientDial TCP客户端 连接
*/
//export SocketClientDial
func SocketClientDial(Context int, addr *C.char, call int, isTls, synchronous bool, ProxyUrl *C.char, CertificateConText int, OutTime int, OutRouterIP *C.char) bool {
	return Api.SocketClientDial(Context, C.GoString(addr), call, nil, isTls, synchronous, C.GoString(ProxyUrl), CertificateConText, OutTime, C.GoString(OutRouterIP))
}

/*
SocketClientSetBufferSize TCP客户端 置缓冲区大小
*/
//export SocketClientSetBufferSize
func SocketClientSetBufferSize(Context, BufferSize int) bool {
	return Api.SocketClientSetBufferSize(Context, BufferSize)
}

/*
SocketClientGetErr TCP客户端 取错误
*/
//export SocketClientGetErr
func SocketClientGetErr(Context int) uintptr {
	return Api.SocketClientGetErr(Context)
}

/*
RemoveSocketClient 释放 TCP客户端
*/
//export RemoveSocketClient
func RemoveSocketClient(Context int) {
	Api.RemoveSocketClient(Context)
}

/*
CreateSocketClient 创建 TCP客户端
*/
//export CreateSocketClient
func CreateSocketClient() int {
	return Api.CreateSocketClient()
}

//==================================================================================================

/*
WebsocketClientReceive Websocket客户端 同步模式下 接收数据 返回数据指针 失败返回0 length=返回数据长度
*/
//export WebsocketClientReceive
func WebsocketClientReceive(Context, OutTimes int) uintptr {
	Buff, messageType := Api.WebsocketClientReceive(Context, OutTimes)
	if Buff == nil {
		return 0
	}
	return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(Buff)), public.BytesCombine(public.IntToBytes(messageType), Buff)))
}

/*
WebsocketReadWrite Websocket客户端  发送数据
*/
//export WebsocketReadWrite
func WebsocketReadWrite(Context int, val uintptr, valLen int, messageType int) bool {
	return Api.WebsocketReadWrite(Context, public.CStringToBytes(val, valLen), messageType)
}

/*
WebsocketClose Websocket客户端 断开
*/
//export WebsocketClose
func WebsocketClose(Context int) {
	Api.WebsocketClose(Context)
}

/*
WebsocketHeartbeat Websocket客户端 心跳设置
*/
//export WebsocketHeartbeat
func WebsocketHeartbeat(Context, HeartbeatTime, call int) {
	Api.WebsocketHeartbeat(Context, HeartbeatTime, call, nil)
}

/*
WebsocketDial Websocket客户端 连接
*/
//export WebsocketDial
func WebsocketDial(Context int, URL, Heads *C.char, call int, synchronous bool, ProxyUrl *C.char, CertificateConText, outTime int, OutRouterIP *C.char) bool {
	return Api.WebsocketDial(Context, C.GoString(URL), C.GoString(Heads), call, nil, synchronous, C.GoString(ProxyUrl), CertificateConText, outTime, C.GoString(OutRouterIP))
}

/*
WebsocketGetErr Websocket客户端 获取错误
*/
//export WebsocketGetErr
func WebsocketGetErr(Context int) uintptr {
	return Api.WebsocketGetErr(Context)
}

/*
RemoveWebsocket 释放 Websocket客户端 对象
*/
//export RemoveWebsocket
func RemoveWebsocket(Context int) {
	Api.RemoveWebsocket(Context)
}

/*
CreateWebsocket 创建 Websocket客户端 对象
*/
//export CreateWebsocket
func CreateWebsocket() int {
	return Api.CreateWebsocket()
}

//==================================================================================================

/*
AddHttpCertificate 创建 Http证书管理器 对象 实现指定Host使用指定证书
*/
//export AddHttpCertificate
func AddHttpCertificate(host *C.char, CertManagerId, Rules int) bool {
	return Api.AddHttpCertificate(C.GoString(host), CertManagerId, uint8(Rules))
}

/*
DelHttpCertificate 删除 Http证书管理器 对象
*/
//export DelHttpCertificate
func DelHttpCertificate(host *C.char) {
	Api.DelHttpCertificate(C.GoString(host))
}

//==================================================================================================

/*
RedisSubscribe Redis 订阅消息
*/
//export RedisSubscribe
func RedisSubscribe(Context int, scribe *C.char, call int, nc bool) bool {
	return Api.RedisSubscribe(Context, C.GoString(scribe), call, nc)
}

/*
RedisDelete Redis 删除
*/
//export RedisDelete
func RedisDelete(Context int, key *C.char) bool {
	return Api.RedisDelete(Context, C.GoString(key))
}

/*
RedisFlushDB Redis 清空当前数据库
*/
//export RedisFlushDB
func RedisFlushDB(Context int) {
	Api.RedisFlushDB(Context)
}

/*
RedisFlushAll Redis 清空redis服务器
*/
//export RedisFlushAll
func RedisFlushAll(Context int) {
	Api.RedisFlushAll(Context)
}

/*
RedisClose Redis 关闭
*/
//export RedisClose
func RedisClose(Context int) {
	Api.RedisClose(Context)
}

/*
RedisGetInt Redis 取整数值
*/
//export RedisGetInt
func RedisGetInt(Context int, key *C.char) int64 {
	return Api.RedisGetInt(Context, C.GoString(key))
}

/*
RedisGetKeys Redis 取指定条件键名
*/
//export RedisGetKeys
func RedisGetKeys(Context int, key *C.char) uintptr {
	bs := Api.RedisGetKeys(Context, C.GoString(key))
	if bs == nil {
		return 0
	}
	return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(bs)), bs))
}

var errorNull = errors.New("")

/*
RedisDo Redis 自定义 执行和查询命令 返回操作结果可能是值 也可能是JSON文本
*/
//export RedisDo
func RedisDo(Context int, args *C.char, error uintptr) uintptr {
	public.WriteErr(errorNull, error)
	p, e := Api.RedisDo(Context, C.GoString(args))
	if e != nil {
		public.WriteErr(e, error)
		return 0
	}
	return public.PointerPtr(p)
}

/*
RedisGetStr Redis 取文本值
*/
//export RedisGetStr
func RedisGetStr(Context int, key *C.char) uintptr {
	s := Api.RedisGetStr(Context, C.GoString(key))
	if s == "" {
		return 0
	}
	return public.PointerPtr(s)
}

/*
RedisGetBytes Redis 取Bytes值
*/
//export RedisGetBytes
func RedisGetBytes(Context int, key *C.char) uintptr {
	p := Api.RedisGetBytes(Context, C.GoString(key))
	if p == nil {
		return 0
	}
	return public.PointerPtr(p)
}

/*
RedisExists Redis 检查指定 key 是否存在
*/
//export RedisExists
func RedisExists(Context int, key *C.char) bool {
	return Api.RedisExists(Context, C.GoString(key))
}

/*
RedisSetNx Redis 设置NX 【如果键名存在返回假】
*/
//export RedisSetNx
func RedisSetNx(Context int, key, val *C.char, expr int) bool {
	return Api.RedisSetNx(Context, C.GoString(key), C.GoString(val), expr)
}

/*
RedisSet Redis 设置值
*/
//export RedisSet
func RedisSet(Context int, key, val *C.char, expr int) bool {
	return Api.RedisSet(Context, C.GoString(key), C.GoString(val), expr)
}

/*
RedisSetBytes Redis 设置Bytes值
*/
//export RedisSetBytes
func RedisSetBytes(Context int, key *C.char, val uintptr, valLen int, expr int) bool {
	data := public.CStringToBytes(val, valLen)
	return Api.RedisSetBytes(Context, C.GoString(key), data, expr)
}

/*
RedisDial Redis 连接
*/
//export RedisDial
func RedisDial(Context int, host, pass *C.char, db, PoolSize, MinIdleCons, DialTimeout, ReadTimeout, WriteTimeout, PoolTimeout, IdleCheckFrequency, IdleTimeout int, error uintptr) bool {
	public.WriteErr(errorNull, error)
	return Api.RedisDial(Context, C.GoString(host), C.GoString(pass), db, PoolSize, MinIdleCons, DialTimeout, ReadTimeout, WriteTimeout, PoolTimeout, IdleCheckFrequency, IdleTimeout, error)
}

/*
RemoveRedis 释放 Redis 对象
*/
//export RemoveRedis
func RemoveRedis(Context int) {
	Api.RemoveRedis(Context)
}

/*
CreateRedis 创建 Redis 对象
*/
//export CreateRedis
func CreateRedis() int {
	return Api.CreateRedis()
}

/*
SetUdpData 设置修改UDP数据
*/
//export SetUdpData
func SetUdpData(MessageId int, val uintptr, valLen int) bool {
	data := public.CStringToBytes(val, valLen)
	return Api.SetUdpData(MessageId, data)
}

/*
GetUdpData 获取UDP数据
*/
//export GetUdpData
func GetUdpData(MessageId int) uintptr {
	bx := Api.GetUdpData(MessageId)
	if len(bx) < 1 {
		return 0
	}
	u := public.PointerPtr(public.BytesCombine(public.IntToBytes(len(bx)), bx))
	return u
}

/*
UdpSendToClient 指定的UDP连接 模拟服务器端向客户端主动发送数据
*/
//export UdpSendToClient
func UdpSendToClient(theology int, data uintptr, dataLen int) bool {
	bs := public.CStringToBytes(data, dataLen)
	return Api.UdpSendToClient(theology, bs)
}

/*
UdpSendToServer 指定的UDP连接 模拟客户端向服务器端主动发送数据
*/
//export UdpSendToServer
func UdpSendToServer(theology int, data uintptr, dataLen int) bool {
	bs := public.CStringToBytes(data, dataLen)
	return Api.UdpSendToServer(theology, bs)
}

// SetScriptCode 加载用户的脚本代码
//
//export SetScriptCode
func SetScriptCode(SunnyContext int, code uintptr, length int) uintptr {
	a := public.CStringToBytes(code, length)
	return public.PointerPtr(Api.SetScriptCode(SunnyContext, string(a)))
}

// SetScriptCall 设置脚本代码的回调函数
//
//export SetScriptCall
func SetScriptCall(SunnyContext int, LOG, SAVE uintptr) {
	Api.SetScriptCall(SunnyContext, LOG, SAVE)
}

/*
SetScriptPage  设置脚本编辑器页面 需不少于8个字符
*/
//export SetScriptPage
func SetScriptPage(SunnyContext int, Page *C.char) uintptr {
	return Api.SetScriptPage(SunnyContext, C.GoString(Page))
}

/*
DisableTCP  禁用TCP 仅对当前SunnyContext有效
*/
//export DisableTCP
func DisableTCP(SunnyContext int, Disable bool) bool {
	return Api.DisableTCP(SunnyContext, Disable)
}

/*
DisableUDP  禁用TCP 仅对当前SunnyContext有效
*/
//export DisableUDP
func DisableUDP(SunnyContext int, Disable bool) bool {
	return Api.DisableUDP(SunnyContext, Disable)
}

/*
SetRandomTLS 是否使用随机TLS指纹 仅对当前SunnyContext有效
*/
//export SetRandomTLS
func SetRandomTLS(SunnyContext int, open bool) bool {
	return Api.SetRandomTLS(SunnyContext, open)
}

/*
SetDnsServer Dns解析服务器 默认:223.5.5.5:853
*/
//export SetDnsServer
func SetDnsServer(ServerName *C.char) {
	dns.SetDnsServer(C.GoString(ServerName))
}

/*
SetOutRouterIP 设置数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11（全局）
*/
//export SetOutRouterIP
func SetOutRouterIP(SunnyContext int, value *C.char) bool {
	return Api.SetOutRouterIP(SunnyContext, C.GoString(value))
}

/*
RequestSetOutRouterIP 设置数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11（TCP/HTTP请求共用这个函数）
*/
//export RequestSetOutRouterIP
func RequestSetOutRouterIP(MessageId int, value *C.char) bool {
	return Api.RequestSetOutRouterIP(MessageId, C.GoString(value))
}

/*
HTTPSetOutRouterIP
HTTP 客户端 设置数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11（TCP/HTTP请求共用这个函数）
*/
//export HTTPSetOutRouterIP
func HTTPSetOutRouterIP(Context int, value *C.char) bool {
	return Api.HTTPSetOutRouterIP(Context, C.GoString(value))
}
