/*
本类为所有动态库导出函数集合
*/
package main

import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/qtgolang/SunnyNet/Api"
	. "github.com/qtgolang/SunnyNet/JavaApi"
	"github.com/qtgolang/SunnyNet/JavaApi/sig"
	"github.com/qtgolang/SunnyNet/SunnyNet"
	"github.com/qtgolang/SunnyNet/src/Compress"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/SunnyNetUDP"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/public"
)

/*

Go、JNI 和 Java 参数对应表
Go 类型	JNI 类型	Java 类型	说明
bool	jboolean	boolean	1 字节，true 或 false
int8	jbyte	byte	1 字节整数
int16	jshort	short	2 字节整数
int32	jint	int	4 字节整数
int64	jlong	long	8 字节整数
uint8	jbyte	byte	1 字节无符号整数
uint16	jshort	short	2 字节无符号整数
uint32	jint	int	4 字节无符号整数
uint64	jlong	long	8 字节无符号整数
float32	jfloat	float	4 字节浮点数
float64	jdouble	double	8 字节浮点数
string	jstring	String	Java 字符串
[]byte	jbyteArray	byte[]	Java 字节数组
[]int32	jintArray	int[]	Java 整数数组
[]float64	jdoubleArray	double[]	Java 双精度浮点数数组
[]string	jobjectArray	String[]	Java 字符串对象数组
struct	jobject	自定义 Java 对象	Java 对象
*/

/*
Java_com_SunnyNet_api_GetSunnyVersion 获取SunnyNet版本
*/
//export Java_com_SunnyNet_api_GetSunnyVersion
func Java_com_SunnyNet_api_GetSunnyVersion(envObj uintptr, clazz uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(public.SunnyVersion)
}

/*
Java_com_SunnyNet_api_CreateSunnyNet 创建Sunny中间件对象,可创建多个
*/
//export Java_com_SunnyNet_api_CreateSunnyNet
func Java_com_SunnyNet_api_CreateSunnyNet(envObj uintptr, clazz uintptr) int64 {
	id := Api.CreateSunnyNet()
	return int64(id)
}

/*
Java_com_SunnyNet_api_ReleaseSunnyNet  释放SunnyNet
*/
//export Java_com_SunnyNet_api_ReleaseSunnyNet
func Java_com_SunnyNet_api_ReleaseSunnyNet(envObj uintptr, clazz uintptr, SunnyContext int64) bool {
	return Api.ReleaseSunnyNet(int(SunnyContext))
}

/*
Java_com_SunnyNet_api_SunnyNetStart 启动Sunny中间件 成功返回true
*/
//export Java_com_SunnyNet_api_SunnyNetStart
func Java_com_SunnyNet_api_SunnyNetStart(envObj uintptr, clazz uintptr, SunnyContext int64) bool {
	return Api.SunnyNetStart(int(SunnyContext))
}

/*
Java_com_SunnyNet_api_SunnyNetSetPort 设置指定端口 Sunny中间件启动之前调用
*/
//export Java_com_SunnyNet_api_SunnyNetSetPort
func Java_com_SunnyNet_api_SunnyNetSetPort(envObj uintptr, clazz uintptr, SunnyContext, Port int64) bool {
	return Api.SunnyNetSetPort(int(SunnyContext), int(Port))
}

/*
Java_com_SunnyNet_api_SunnyNetClose 关闭停止指定Sunny中间件
*/
//export Java_com_SunnyNet_api_SunnyNetClose
func Java_com_SunnyNet_api_SunnyNetClose(envObj uintptr, clazz uintptr, SunnyContext int64) bool {
	return Api.SunnyNetClose(int(SunnyContext))
}

/*
Java_com_SunnyNet_api_SunnyNetSetCert 设置自定义证书
*/
//export Java_com_SunnyNet_api_SunnyNetSetCert
func Java_com_SunnyNet_api_SunnyNetSetCert(envObj uintptr, clazz uintptr, SunnyContext, CertificateManagerId int64) bool {
	return Api.SunnyNetSetCert(int(SunnyContext), int(CertificateManagerId))
}

/*
Java_com_SunnyNet_api_SunnyNetInstallCert 安装证书 将证书安装到Windows系统内
*/
//export Java_com_SunnyNet_api_SunnyNetInstallCert
func Java_com_SunnyNet_api_SunnyNetInstallCert(envObj uintptr, clazz uintptr, SunnyContext int64) uintptr {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[int(SunnyContext)]
	SunnyNet.SunnyStorageLock.Unlock()
	env := Env(envObj)
	if w == nil {
		return env.NewString("SunnyNet no exist")
	}
	return env.NewString(w.InstallCert())
}

/*
Java_com_SunnyNet_api_SunnyNetSetCallback 设置中间件回调地址 httpCallback
*/
//export Java_com_SunnyNet_api_SunnyNetSetCallback
func Java_com_SunnyNet_api_SunnyNetSetCallback(envObj uintptr, clazz uintptr, SunnyContext int64, Callback uintptr) bool {

	SunnyNet.SunnyStorageLock.Lock()
	s := SunnyNet.SunnyStorage[int(SunnyContext)]
	SunnyNet.SunnyStorageLock.Unlock()
	if s == nil {
		return false
	}

	env := Env(envObj)
	obj := env.NewGlobalRef(Callback)
	cls := env.GetObjectClass(obj)
	FuncSig := "(Lcom/SunnyNet/Internal/HTTPEvent;)V"
	onHTTPCallbackMethodId := env.GetMethodID(cls, "onHTTPCallback", FuncSig)
	if onHTTPCallbackMethodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onHTTPCallback"+FuncSig+"] failed")
		panic("Find Class [onHTTPCallback" + FuncSig + "] failed")
	}
	FuncSig = "(Lcom/SunnyNet/Internal/WebSocketEvent;)V"
	onWebSocketMethodId := env.GetMethodID(cls, "onWebSocketCallback", FuncSig)
	if onWebSocketMethodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onWebSocketCallback"+FuncSig+"] failed")
		panic("Find Class [onWebSocketCallback" + FuncSig + "] failed")
	}
	FuncSig = "(Lcom/SunnyNet/Internal/TCPEvent;)V"
	onTCPMethodId := env.GetMethodID(cls, "onTCPCallback", FuncSig)
	if onTCPMethodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onTCPCallback"+FuncSig+"] failed")
		panic("Find Class [onTCPCallback" + FuncSig + "] failed")
	}
	FuncSig = "(Lcom/SunnyNet/Internal/UDPEvent;)V"
	onUDPMethodId := env.GetMethodID(cls, "onUDPCallback", FuncSig)
	if onUDPMethodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onUDPCallback"+FuncSig+"] failed")
		panic("Find Class [onUDPCallback" + FuncSig + "] failed")
	}
	httpCallback := func(Conn SunnyNet.ConnHTTP) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()

		_Method := _env.NewString(Conn.Method())
		_url := _env.NewString(Conn.URL())
		_er := _env.NewString(Conn.Error())
		HTTPEventClass := aliasToClass("HTTPEvent")
		EventConstructor := _env.GetMethodID(HTTPEventClass, "<init>", fmt.Sprintf("(%s%s%s%s%s%s%s%s)%s", sig.Long, sig.Long, sig.Long, sig.Long, sig.String, sig.String, sig.String, sig.Long, sig.Void))
		EventObj := _env.NewObjectA(HTTPEventClass, EventConstructor, Jvalue(SunnyContext), Jvalue(Conn.Theology()), Jvalue(Conn.MessageId()), Jvalue(Conn.Type()), Jvalue(_Method), Jvalue(_url), Jvalue(_er), Jvalue(Conn.PID()))

		_env.CallVoidMethodA(obj, onHTTPCallbackMethodId, Jvalue(EventObj))
		_env.DeleteLocalRef(EventObj)
		_env.DeleteLocalRef(_Method)
		_env.DeleteLocalRef(_url)
		_env.DeleteLocalRef(_er)
		return
	}

	tcpCallback := func(Conn SunnyNet.ConnTCP) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()
		_LocalAddr := _env.NewString(Conn.LocalAddress())
		_RemoteAddr := _env.NewString(Conn.RemoteAddress())
		_data := _env.NewByteArray(Conn.Body())
		TCPEventClass := aliasToClass("TCPEvent")
		EventConstructor := _env.GetMethodID(TCPEventClass, "<init>", fmt.Sprintf("(%s%s%s%s%s%s%s%s)%s", sig.Long, sig.String, sig.String, sig.Long, sig.Long, sig.Long, sig.Long, sig.ByteArray, sig.Void))
		EventObj := _env.NewObjectA(TCPEventClass, EventConstructor, Jvalue(SunnyContext), Jvalue(_LocalAddr), Jvalue(_RemoteAddr), Jvalue(Conn.Theology()), Jvalue(Conn.MessageId()), Jvalue(Conn.Type()), Jvalue(Conn.PID()), Jvalue(_data))

		_env.CallVoidMethodA(obj, onTCPMethodId, Jvalue(EventObj))
		_env.DeleteLocalRef(EventObj)
		_env.DeleteLocalRef(_LocalAddr)
		_env.DeleteLocalRef(_RemoteAddr)
		_env.DeleteLocalRef(_data)
	}

	wsCallback := func(Conn SunnyNet.ConnWebSocket) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()

		_Method := _env.NewString(Conn.Method())
		_url := _env.NewString(Conn.URL())
		WebSocketEventClass := aliasToClass("WebSocketEvent")
		EventConstructor := _env.GetMethodID(WebSocketEventClass, "<init>", fmt.Sprintf("(%s%s%s%s%s%s%s%s)%s", sig.Long, sig.Long, sig.Long, sig.Long, sig.String, sig.String, sig.Long, sig.Long, sig.Void))
		EventObj := _env.NewObjectA(WebSocketEventClass, EventConstructor, Jvalue(SunnyContext), Jvalue(Conn.Theology()), Jvalue(Conn.MessageId()), Jvalue(Conn.Type()), Jvalue(_Method), Jvalue(_url), Jvalue(Conn.PID()), Jvalue(Conn.MessageType()))

		_env.CallVoidMethodA(obj, onWebSocketMethodId, Jvalue(EventObj))
		_env.DeleteLocalRef(_Method)
		_env.DeleteLocalRef(_url)
		_env.DeleteLocalRef(EventObj)
		return
	}

	udpCallback := func(Conn SunnyNet.ConnUDP) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()
		MessageId := Conn.MessageId()
		SunnyNetUDP.ResetMessage(MessageId, Conn.Body())

		_LocalAddr := _env.NewString(Conn.LocalAddress())
		_RemoteAddr := _env.NewString(Conn.RemoteAddress())
		UDPEventClass := aliasToClass("UDPEvent")
		EventConstructor := _env.GetMethodID(UDPEventClass, "<init>", fmt.Sprintf("(%s%s%s%s%s%s%s)%s", sig.Long, sig.String, sig.String, sig.Long, sig.Long, sig.Long, sig.Long, sig.Void))
		EventObj := _env.NewObjectA(UDPEventClass, EventConstructor, Jvalue(SunnyContext), Jvalue(_LocalAddr), Jvalue(_RemoteAddr), Jvalue(Conn.Theology()), Jvalue(MessageId), Jvalue(Conn.Type()), Jvalue(Conn.PID()))
		_env.CallVoidMethodA(obj, onUDPMethodId, Jvalue(EventObj))
		_env.DeleteLocalRef(EventObj)
		_env.DeleteLocalRef(_LocalAddr)
		_env.DeleteLocalRef(_RemoteAddr)

		Conn.SetBody(SunnyNetUDP.GetMessage(MessageId))
		SunnyNetUDP.DelMessage(MessageId)
		return
	}

	FuncSig = fmt.Sprintf("(%s%s)%s", sig.Long, sig.String, sig.Void)
	onScriptLogMethodId := env.GetMethodID(cls, "onScriptLogCallback", FuncSig)
	if onScriptLogMethodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onScriptLogCallback"+FuncSig+"] failed")
		panic("Find Class [onScriptLogCallback" + FuncSig + "] failed")
	}
	onScriptCodeSaveMethodId := env.GetMethodID(cls, "onScriptCodeSaveCallback", FuncSig)
	if onScriptCodeSaveMethodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onScriptCodeSaveCallback"+FuncSig+"] failed")
		panic("Find Class [onScriptCodeSaveCallback" + FuncSig + "] failed")
	}
	log := func(Context int, info ...any) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()
		_logInfo := _env.NewString(fmt.Sprintf("%v", info))
		_env.CallVoidMethodA(obj, onScriptLogMethodId, Jvalue(Context), Jvalue(_logInfo))
		_env.DeleteLocalRef(_logInfo)
	}
	code := func(Context int, code []byte) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()
		_ScriptCode := _env.NewString(string(code))
		_env.CallVoidMethodA(obj, onScriptCodeSaveMethodId, Jvalue(Context), Jvalue(_ScriptCode))
		_env.DeleteLocalRef(_ScriptCode)
	}
	s.SetScriptCall(log, code)
	s.SetGoCallback(httpCallback, tcpCallback, wsCallback, udpCallback)
	Java_GlobalRef_Add("SunnyNet", obj, int(SunnyContext))
	return true
}

/*
Java_com_SunnyNet_api_SunnyNetSocket5AddUser 添加 S5代理需要验证的用户名
*/
//export Java_com_SunnyNet_api_SunnyNetSocket5AddUser
func Java_com_SunnyNet_api_SunnyNetSocket5AddUser(envObj uintptr, clazz uintptr, SunnyContext int64, User, Pass uintptr) bool {
	env := Env(envObj)
	return Api.SunnyNetSocket5AddUser(int(SunnyContext), env.GetString(User), env.GetString(Pass))
}

/*
Java_com_SunnyNet_api_SunnyNetVerifyUser 开启身份验证模式
*/
//export Java_com_SunnyNet_api_SunnyNetVerifyUser
func Java_com_SunnyNet_api_SunnyNetVerifyUser(envObj uintptr, clazz uintptr, SunnyContext int64, open bool) bool {
	return Api.SunnyNetVerifyUser(int(SunnyContext), open)
}

/*
Java_com_SunnyNet_api_SunnyNetSocket5DelUser 删除 S5需要验证的用户名
*/
//export Java_com_SunnyNet_api_SunnyNetSocket5DelUser
func Java_com_SunnyNet_api_SunnyNetSocket5DelUser(envObj uintptr, clazz uintptr, SunnyContext int64, User uintptr) bool {
	env := Env(envObj)
	return Api.SunnyNetSocket5DelUser(int(SunnyContext), env.GetString(User))
}

/*
Java_com_SunnyNet_api_SunnyNetGetSocket5User 开启身份验证模式后 获取授权的S5账号,注意UDP请求无法获取到授权的s5账号
*/
//export Java_com_SunnyNet_api_SunnyNetGetSocket5User
func Java_com_SunnyNet_api_SunnyNetGetSocket5User(envObj uintptr, clazz uintptr, Theology int64) uintptr {
	env := Env(envObj)
	return env.NewString(SunnyNet.GetSocket5User(int(Theology)))
}

/*
Java_com_SunnyNet_api_SunnyNetMustTcp 设置中间件是否开启强制走TCP
*/
//export Java_com_SunnyNet_api_SunnyNetMustTcp
func Java_com_SunnyNet_api_SunnyNetMustTcp(envObj uintptr, clazz uintptr, SunnyContext int64, open bool) {
	Api.SunnyNetMustTcp(int(SunnyContext), open)
}

/*
Java_com_SunnyNet_api_CompileProxyRegexp 设置中间件上游代理使用规则
*/
//export Java_com_SunnyNet_api_CompileProxyRegexp
func Java_com_SunnyNet_api_CompileProxyRegexp(envObj uintptr, clazz uintptr, SunnyContext int64, Regexp uintptr) bool {
	env := Env(envObj)
	return Api.CompileProxyRegexp(int(SunnyContext), env.GetString(Regexp))
}

/*
Java_com_SunnyNet_api_SetMustTcpRegexp 设置强制走TCP规则,如果 打开了全部强制走TCP状态,本功能则无效
*/
//export Java_com_SunnyNet_api_SetMustTcpRegexp
func Java_com_SunnyNet_api_SetMustTcpRegexp(envObj uintptr, clazz uintptr, SunnyContext int64, Regexp uintptr, RulesAllow bool) bool {
	env := Env(envObj)
	return Api.SetMustTcpRegexp(int(SunnyContext), env.GetString(Regexp), RulesAllow)
}

/*
Java_com_SunnyNet_api_SunnyNetError 获取中间件启动时的错误信息
*/
//export Java_com_SunnyNet_api_SunnyNetError
func Java_com_SunnyNet_api_SunnyNetError(envObj uintptr, clazz uintptr, SunnyContext int64) uintptr {
	//return Api.SunnyNetError(int(SunnyContext))
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[int(SunnyContext)]
	SunnyNet.SunnyStorageLock.Unlock()
	env := Env(envObj)
	if w == nil {
		return env.NewString("")
	}
	if w.Error == nil {
		return env.NewString("")
	}
	return env.NewString(w.Error.Error())

}

/*
Java_com_SunnyNet_api_SetGlobalProxy 设置全局上游代理 仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
*/
//
//export Java_com_SunnyNet_api_SetGlobalProxy
func Java_com_SunnyNet_api_SetGlobalProxy(envObj uintptr, clazz uintptr, SunnyContext int64, ProxyAddress uintptr, outTime int64) bool {
	env := Env(envObj)
	return Api.SetGlobalProxy(int(SunnyContext), env.GetString(ProxyAddress), int(outTime))
}

/*
Java_com_SunnyNet_api_GetRequestProto 获取 HTTPS 请求的协议版本
*/
//export Java_com_SunnyNet_api_GetRequestProto
func Java_com_SunnyNet_api_GetRequestProto(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	//return Api.GetRequestProto(int(MessageId))
	env := Env(envObj)
	k, ok := SunnyNet.GetSceneProxyRequest(int(MessageId))
	if ok == false {
		return env.NewString("")
	}
	if k == nil {
		return env.NewString("")
	}
	if k.Request == nil {
		return env.NewString("")
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	return env.NewString(k.Request.Proto)

}

/*
Java_com_SunnyNet_api_GetResponseProto 获取 HTTPS 响应的协议版本
*/
//export Java_com_SunnyNet_api_GetResponseProto
func Java_com_SunnyNet_api_GetResponseProto(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	//return Api.GetResponseProto(int(MessageId))
	env := Env(envObj)
	k, ok := SunnyNet.GetSceneProxyRequest(int(MessageId))
	if ok == false {
		return env.NewString("")
	}
	if k == nil {
		return env.NewString("")
	}
	if k.Response.Response == nil {
		return env.NewString("")
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	return env.NewString(k.Response.Proto)
}

/*
Java_com_SunnyNet_api_ExportCert 导出已设置的证书
*/
//export Java_com_SunnyNet_api_ExportCert
func Java_com_SunnyNet_api_ExportCert(envObj uintptr, clazz uintptr, SunnyContext int64) uintptr {
	//	return Api.ExportCert(int(SunnyContext))
	env := Env(envObj)
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[int(SunnyContext)]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		return env.NewString(string(w.ExportCert()))
	}
	return env.NewString("")
}

/*
Java_com_SunnyNet_api_SetHTTPRequestMaxUpdateLength 设置HTTP请求,提交数据,最大的长度
*/
//export Java_com_SunnyNet_api_SetHTTPRequestMaxUpdateLength
func Java_com_SunnyNet_api_SetHTTPRequestMaxUpdateLength(envObj uintptr, clazz uintptr, SunnyContext, i int64) bool {
	return Api.SetHTTPRequestMaxUpdateLength(int(SunnyContext), i)
}

/*
Java_com_SunnyNet_api_CancelIEProxy 取消设置的IE代理
*/
//export Java_com_SunnyNet_api_CancelIEProxy
func Java_com_SunnyNet_api_CancelIEProxy(envObj uintptr, clazz uintptr, SunnyContext int64) bool {
	return Api.CancelIEProxy(int(SunnyContext))
}

/*
Java_com_SunnyNet_api_SetIeProxy 设置IE代理
*/
//export Java_com_SunnyNet_api_SetIeProxy
func Java_com_SunnyNet_api_SetIeProxy(envObj uintptr, clazz uintptr, SunnyContext int64) bool {
	return Api.SetIeProxy(int(SunnyContext))
}

/*
Java_com_SunnyNet_api_SetRequestCookie 修改、设置 HTTP/S当前请求数据中指定Cookie
*/
//export Java_com_SunnyNet_api_SetRequestCookie
func Java_com_SunnyNet_api_SetRequestCookie(envObj uintptr, clazz uintptr, MessageId int64, name, val uintptr) {
	env := Env(envObj)
	Api.SetRequestCookie(int(MessageId), env.GetString(name), env.GetString(val))
}

/*
Java_com_SunnyNet_api_SetRequestAllCookie 修改、设置 HTTP/S当前请求数据中的全部Cookie
*/
//export Java_com_SunnyNet_api_SetRequestAllCookie
func Java_com_SunnyNet_api_SetRequestAllCookie(envObj uintptr, clazz uintptr, MessageId int64, val uintptr) {
	env := Env(envObj)
	Api.SetRequestAllCookie(int(MessageId), env.GetString(val))
}

/*
Java_com_SunnyNet_api_GetRequestCookie 获取 HTTP/S当前请求数据中指定的Cookie
*/
//export Java_com_SunnyNet_api_GetRequestCookie
func Java_com_SunnyNet_api_GetRequestCookie(envObj uintptr, clazz uintptr, MessageId int64, name uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetRequestCookie(int(MessageId), env.GetString(name)))
}

/*
Java_com_SunnyNet_api_GetRequestALLCookie 获取 HTTP/S 当前请求全部Cookie
*/
//export Java_com_SunnyNet_api_GetRequestALLCookie
func Java_com_SunnyNet_api_GetRequestALLCookie(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetRequestALLCookie(int(MessageId)))
}

/*
Java_com_SunnyNet_api_DelResponseHeader 删除HTTP/S返回数据中指定的协议头
*/
//export Java_com_SunnyNet_api_DelResponseHeader
func Java_com_SunnyNet_api_DelResponseHeader(envObj uintptr, clazz uintptr, MessageId int64, name uintptr) {
	env := Env(envObj)
	Api.DelResponseHeader(int(MessageId), env.GetString(name))
}

/*
Java_com_SunnyNet_api_DelRequestHeader 删除HTTP/S请求数据中指定的协议头
*/
//export Java_com_SunnyNet_api_DelRequestHeader
func Java_com_SunnyNet_api_DelRequestHeader(envObj uintptr, clazz uintptr, MessageId int64, name uintptr) {
	env := Env(envObj)
	Api.DelRequestHeader(int(MessageId), env.GetString(name))
}

/*
Java_com_SunnyNet_api_SetRequestOutTime 请求设置超时-毫秒
*/
//export Java_com_SunnyNet_api_SetRequestOutTime
func Java_com_SunnyNet_api_SetRequestOutTime(envObj uintptr, clazz uintptr, MessageId int64, times int64) {
	Api.SetRequestOutTime(int(MessageId), int(times))
}

/*
Java_com_SunnyNet_api_SetRequestALLHeader 设置HTTP/ S请求体中的全部协议头
*/
//export Java_com_SunnyNet_api_SetRequestALLHeader
func Java_com_SunnyNet_api_SetRequestALLHeader(envObj uintptr, clazz uintptr, MessageId int64, val uintptr) {
	env := Env(envObj)
	Api.SetRequestALLHeader(int(MessageId), env.GetString(val))
}

/*
Java_com_SunnyNet_api_SetRequestHeader 设置HTTP/S请求体中的协议头
*/
//export Java_com_SunnyNet_api_SetRequestHeader
func Java_com_SunnyNet_api_SetRequestHeader(envObj uintptr, clazz uintptr, MessageId int64, name, val uintptr) {
	env := Env(envObj)
	Api.SetRequestHeader(int(MessageId), env.GetString(name), env.GetString(val))
}

/*
Java_com_SunnyNet_api_RandomRequestCipherSuites 随机设置请求 CipherSuites
*/
//export Java_com_SunnyNet_api_RandomRequestCipherSuites
func Java_com_SunnyNet_api_RandomRequestCipherSuites(envObj uintptr, clazz uintptr, MessageId int64) bool {
	return Api.SetRequestCipherSuites(int(MessageId))
}

/*
Java_com_SunnyNet_api_SetRequestHTTP2Config  设置HTTP 2.0 请求指纹配置 (若服务器支持则使用,若服务器不支持,设置了也不会使用),如果强制请求发送时使用HTTP/1.1 请填入参数 http/1.1
*/
//export Java_com_SunnyNet_api_SetRequestHTTP2Config
func Java_com_SunnyNet_api_SetRequestHTTP2Config(envObj uintptr, clazz uintptr, MessageId int64, h2Config uintptr) bool {
	env := Env(envObj)
	return Api.SetRequestHTTP2Config(int(MessageId), env.GetString(h2Config))
}

/*
Java_com_SunnyNet_api_SetResponseHeader 修改、设置 HTTP/S当前返回数据中的指定协议头
*/
//export Java_com_SunnyNet_api_SetResponseHeader
func Java_com_SunnyNet_api_SetResponseHeader(envObj uintptr, clazz uintptr, MessageId int64, name uintptr, val uintptr) {
	env := Env(envObj)
	Api.SetResponseHeader(int(MessageId), env.GetString(name), env.GetString(val))
}

/*
Java_com_SunnyNet_api_GetRequestHeader 获取 HTTP/S当前请求数据中的指定协议头
*/
//export Java_com_SunnyNet_api_GetRequestHeader
func Java_com_SunnyNet_api_GetRequestHeader(envObj uintptr, clazz uintptr, MessageId int64, name uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetRequestHeader(int(MessageId), env.GetString(name)))
}

/*
Java_com_SunnyNet_api_GetResponseHeader 获取 HTTP/S 当前返回数据中指定的协议头
*/
//export Java_com_SunnyNet_api_GetResponseHeader
func Java_com_SunnyNet_api_GetResponseHeader(envObj uintptr, clazz uintptr, MessageId int64, name uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetResponseHeader(int(MessageId), env.GetString(name)))
}

/*
Java_com_SunnyNet_api_GetResponseServerAddress 获取 HTTP/S 相应的服务器地址
*/
//export Java_com_SunnyNet_api_GetResponseServerAddress
func Java_com_SunnyNet_api_GetResponseServerAddress(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetResponseServerAddress(int(MessageId)))
}

/*
Java_com_SunnyNet_api_SetResponseAllHeader 修改、设置 HTTP/S当前返回数据中的全部协议头，例如设置返回两条Cookie 使用本命令设置 使用设置、修改 单条命令无效
*/
//export Java_com_SunnyNet_api_SetResponseAllHeader
func Java_com_SunnyNet_api_SetResponseAllHeader(envObj uintptr, clazz uintptr, MessageId int64, value uintptr) {
	env := Env(envObj)
	Api.SetResponseAllHeader(int(MessageId), env.GetString(value))
}

/*
Java_com_SunnyNet_api_GetResponseAllHeader 获取 HTTP/S 当前响应全部协议头
*/
//export Java_com_SunnyNet_api_GetResponseAllHeader
func Java_com_SunnyNet_api_GetResponseAllHeader(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetResponseAllHeader(int(MessageId)))
}

/*
Java_com_SunnyNet_api_GetRequestAllHeader 获取 HTTP/S 当前请求数据全部协议头
*/
//export Java_com_SunnyNet_api_GetRequestAllHeader
func Java_com_SunnyNet_api_GetRequestAllHeader(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	r := Api.GetRequestAllHeader(int(MessageId))
	return env.NewString(r)
}

/*
ava_com_SunnyNet_api_GetMessageNote 获取请求中的注释,由脚本代码中设置
*/
//export ava_com_SunnyNet_api_GetMessageNote
func ava_com_SunnyNet_api_GetMessageNote(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetMessageNote(int(MessageId)))
}

/*
Java_com_SunnyNet_api_SetRequestProxy 设置HTTP/S请求代理，仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
*/
//
//export Java_com_SunnyNet_api_SetRequestProxy
func Java_com_SunnyNet_api_SetRequestProxy(envObj uintptr, clazz uintptr, MessageId int64, ProxyUrl uintptr, outTime int) bool {
	env := Env(envObj)
	return Api.SetRequestProxy(int(MessageId), env.GetString(ProxyUrl), outTime)
}

/*
Java_com_SunnyNet_api_GetResponseStatusCode 获取HTTP/S返回的状态码
*/
//export Java_com_SunnyNet_api_GetResponseStatusCode
func Java_com_SunnyNet_api_GetResponseStatusCode(envObj uintptr, clazz uintptr, MessageId int64) int64 {
	return int64(Api.GetResponseStatusCode(int(MessageId)))
}

/*
Java_com_SunnyNet_api_GetRequestClientIp 获取当前HTTP/S请求由哪个IP发起
*/
//export Java_com_SunnyNet_api_GetRequestClientIp
func Java_com_SunnyNet_api_GetRequestClientIp(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetRequestClientIp(int(MessageId)))
}

/*
Java_com_SunnyNet_api_GetResponseStatus 获取HTTP/S返回的状态文本 例如 [200 OK]
*/
//export Java_com_SunnyNet_api_GetResponseStatus
func Java_com_SunnyNet_api_GetResponseStatus(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetResponseStatus(int(MessageId)))
}

/*
Java_com_SunnyNet_api_SetResponseStatus 修改HTTP/S返回的状态码
*/
//export Java_com_SunnyNet_api_SetResponseStatus
func Java_com_SunnyNet_api_SetResponseStatus(envObj uintptr, clazz uintptr, MessageId, code int64) {
	Api.SetResponseStatus(int(MessageId), int(code))
}

/*
Java_com_SunnyNet_api_SetRequestUrl 修改HTTP/S当前请求的URL
*/
//export Java_com_SunnyNet_api_SetRequestUrl
func Java_com_SunnyNet_api_SetRequestUrl(envObj uintptr, clazz uintptr, MessageId int64, URI uintptr) bool {
	env := Env(envObj)
	return Api.SetRequestUrl(int(MessageId), env.GetString(URI))
}

/*
Java_com_SunnyNet_api_SetResponseData 设置、修改 HTTP/S 当前请求返回数据 如果再发起请求时调用本命令，请求将不会被发送，将会直接返回 data=数据
*/
//export Java_com_SunnyNet_api_SetResponseData
func Java_com_SunnyNet_api_SetResponseData(envObj uintptr, clazz uintptr, MessageId int64, data uintptr) bool {
	env := Env(envObj)
	return Api.SetResponseData(int(MessageId), env.GetBytes(data))
}

/*
Java_com_SunnyNet_api_SetRequestData 设置、修改 HTTP/S 当前请求POST提交数据  data=数据
*/
//export Java_com_SunnyNet_api_SetRequestData
func Java_com_SunnyNet_api_SetRequestData(envObj uintptr, clazz uintptr, MessageId int64, data uintptr) bool {
	env := Env(envObj)
	return Api.SetRequestData(int(MessageId), env.GetBytes(data))
}

/*
Java_com_SunnyNet_api_GetRequestBody 获取 HTTP/S 当前POST提交数据 返回 数据指针
*/
//export Java_com_SunnyNet_api_GetRequestBody
func Java_com_SunnyNet_api_GetRequestBody(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.GetRequestBody(int(MessageId)))
}

/*
Java_com_SunnyNet_api_IsRequestRawBody 此请求是否为原始body 如果是 将无法修改提交的Body，请使用 RawRequestDataToFile 命令来储存到文件
*/
//export Java_com_SunnyNet_api_IsRequestRawBody
func Java_com_SunnyNet_api_IsRequestRawBody(envObj uintptr, clazz uintptr, MessageId int64) bool {
	return Api.IsRequestRawBody(int(MessageId))
}

/*
Java_com_SunnyNet_api_RawRequestDataToFile 获取 HTTP/ S 当前POST提交数据原始Data,传入保存文件名路径,例如"c:\1.txt"
*/
//export Java_com_SunnyNet_api_RawRequestDataToFile
func Java_com_SunnyNet_api_RawRequestDataToFile(envObj uintptr, clazz uintptr, MessageId int64, saveFileName uintptr) bool {
	env := Env(envObj)
	return Api.RawRequestDataToFile(int(MessageId), env.GetString(saveFileName))
}

/*
Java_com_SunnyNet_api_GetResponseBody 获取 HTTP/S 当前返回数据
*/
//export Java_com_SunnyNet_api_GetResponseBody
func Java_com_SunnyNet_api_GetResponseBody(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.GetResponseBody(int(MessageId)))
}

/*
Java_com_SunnyNet_api_CloseWebsocket 主动关闭Websocket
*/
//export Java_com_SunnyNet_api_CloseWebsocket
func Java_com_SunnyNet_api_CloseWebsocket(envObj uintptr, clazz uintptr, Theology int64) bool {
	return Api.CloseWebsocket(int(Theology))
}

/*
Java_com_SunnyNet_api_GetWebsocketBody 获取 WebSocket消息
*/
//export Java_com_SunnyNet_api_GetWebsocketBody
func Java_com_SunnyNet_api_GetWebsocketBody(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.GetWebsocketBody(int(MessageId)))
}

/*
Java_com_SunnyNet_api_SetWebsocketBody 修改 WebSocket消息 data=数据
*/
//export Java_com_SunnyNet_api_SetWebsocketBody
func Java_com_SunnyNet_api_SetWebsocketBody(envObj uintptr, clazz uintptr, MessageId int64, data uintptr) bool {
	env := Env(envObj)
	return Api.SetWebsocketBody(int(MessageId), env.GetBytes(data))
}

/*
Java_com_SunnyNet_api_SendWebsocketBody 主动向Websocket服务器发送消息 MessageType=WS消息类型 data=数据指针  dataLen=数据长度
*/
//export Java_com_SunnyNet_api_SendWebsocketBody
func Java_com_SunnyNet_api_SendWebsocketBody(envObj uintptr, clazz uintptr, Theology, MessageType int64, data uintptr) bool {
	env := Env(envObj)
	return Api.SendWebsocketBody(int(Theology), int(MessageType), env.GetBytes(data))
}

/*
Java_com_SunnyNet_api_SendWebsocketClientBody  主动向Websocket客户端发送消息 MessageType=WS消息类型 data=数据指针  dataLen=数据长度
*/
//export Java_com_SunnyNet_api_SendWebsocketClientBody
func Java_com_SunnyNet_api_SendWebsocketClientBody(envObj uintptr, clazz uintptr, Theology, MessageType int64, data uintptr) bool {
	env := Env(envObj)
	return Api.SendWebsocketClientBody(int(Theology), int(MessageType), env.GetBytes(data))
}

/*
Java_com_SunnyNet_api_SetTcpBody 修改 TCP消息数据 MsgType=1 发送的消息 MsgType=2 接收的消息 如果 MsgType和MessageId不匹配，将不会执行操作  data=数据指针  dataLen=数据长度
*/
//export Java_com_SunnyNet_api_SetTcpBody
func Java_com_SunnyNet_api_SetTcpBody(envObj uintptr, clazz uintptr, MessageId, MsgType int64, data uintptr) bool {
	env := Env(envObj)
	return Api.SetTcpBody(int(MessageId), int(MsgType), env.GetBytes(data))
}

/*
Java_com_SunnyNet_api_SetTcpAgent 给当前TCP连接设置代理 仅限 TCP回调 即将连接时使用 仅支持S5代理 例如 socket5://admin:123456@127.0.0.1:8888
*/
//
//export Java_com_SunnyNet_api_SetTcpAgent
func Java_com_SunnyNet_api_SetTcpAgent(envObj uintptr, clazz uintptr, MessageId int64, ProxyUrl uintptr, outTime int) bool {
	env := Env(envObj)
	return Api.SetTcpAgent(int(MessageId), env.GetString(ProxyUrl), outTime)
}

/*
Java_com_SunnyNet_api_TcpCloseClient 根据唯一ID关闭指定的TCP连接  唯一ID在回调参数中
*/
//export Java_com_SunnyNet_api_TcpCloseClient
func Java_com_SunnyNet_api_TcpCloseClient(envObj uintptr, clazz uintptr, theology int64) bool {
	return Api.TcpCloseClient(int(theology))
}

/*
Java_com_SunnyNet_api_SetTcpConnectionIP 给指定的TCP连接 修改目标连接地址 目标地址必须带端口号 例如 baidu.com:443
*/
//export Java_com_SunnyNet_api_SetTcpConnectionIP
func Java_com_SunnyNet_api_SetTcpConnectionIP(envObj uintptr, clazz uintptr, MessageId int64, address uintptr) bool {
	env := Env(envObj)
	return Api.SetTcpConnectionIP(int(MessageId), env.GetString(address))
}

/*
Java_com_SunnyNet_api_TcpSendMsg 指定的TCP连接 模拟客户端向服务器端主动发送数据
*/
//export Java_com_SunnyNet_api_TcpSendMsg
func Java_com_SunnyNet_api_TcpSendMsg(envObj uintptr, clazz uintptr, theology int64, data uintptr) bool {
	env := Env(envObj)
	return Api.TcpSendMsg(int(theology), env.GetBytes(data)) > 0
}

/*
Java_com_SunnyNet_api_TcpSendMsgClient 指定的TCP连接 模拟服务器端向客户端主动发送数据
*/
//export Java_com_SunnyNet_api_TcpSendMsgClient
func Java_com_SunnyNet_api_TcpSendMsgClient(envObj uintptr, clazz uintptr, theology int64, data uintptr) bool {
	env := Env(envObj)
	return Api.TcpSendMsgClient(int(theology), env.GetBytes(data)) > 0
}

/*
Java_com_SunnyNet_api_GzipUnCompress Gzip解压缩
*/
//export Java_com_SunnyNet_api_GzipUnCompress
func Java_com_SunnyNet_api_GzipUnCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.GzipUnCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_BrUnCompress br解压缩
*/
//export Java_com_SunnyNet_api_BrUnCompress
func Java_com_SunnyNet_api_BrUnCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.BrUnCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_BrCompress br压缩
*/
//export Java_com_SunnyNet_api_BrCompress
func Java_com_SunnyNet_api_BrCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.BrCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_ZSTDDecompress ZSTD解压缩
*/
//export Java_com_SunnyNet_api_ZSTDDecompress
func Java_com_SunnyNet_api_ZSTDDecompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.ZSTDDecompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_ZSTDCompress ZSTD压缩
*/
//export Java_com_SunnyNet_api_ZSTDCompress
func Java_com_SunnyNet_api_ZSTDCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.ZSTDCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_GzipCompress Gzip压缩
*/
//export Java_com_SunnyNet_api_GzipCompress
func Java_com_SunnyNet_api_GzipCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.GzipCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_ZlibCompress Zlib压缩
*/
//export Java_com_SunnyNet_api_ZlibCompress
func Java_com_SunnyNet_api_ZlibCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.ZlibCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_ZlibUnCompress Zlib解压缩
*/
//export Java_com_SunnyNet_api_ZlibUnCompress
func Java_com_SunnyNet_api_ZlibUnCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.ZlibUnCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_DeflateUnCompress Deflate解压缩 (可能等同于zlib解压缩)
*/
//export Java_com_SunnyNet_api_DeflateUnCompress
func Java_com_SunnyNet_api_DeflateUnCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.DeflateUnCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_DeflateCompress Deflate压缩 (可能等同于zlib压缩)
*/
//export Java_com_SunnyNet_api_DeflateCompress
func Java_com_SunnyNet_api_DeflateCompress(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Compress.DeflateCompress(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_WebpToJpegBytes Webp图片转JEG图片字节数组 SaveQuality=质量(默认75)
*/
//export Java_com_SunnyNet_api_WebpToJpegBytes
func Java_com_SunnyNet_api_WebpToJpegBytes(envObj uintptr, clazz uintptr, data uintptr, SaveQuality int64) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.WebpToJpegBytes(env.GetBytes(data), int(SaveQuality)))
}

/*
Java_com_SunnyNet_api_WebpToPngBytes Webp图片转Png图片字节数组
*/
//export Java_com_SunnyNet_api_WebpToPngBytes
func Java_com_SunnyNet_api_WebpToPngBytes(envObj uintptr, clazz uintptr, data uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.WebpToPngBytes(env.GetBytes(data)))
}

/*
Java_com_SunnyNet_api_WebpToJpeg Webp图片转JEG图片 根据文件名 SaveQuality=质量(默认75)
*/
//export Java_com_SunnyNet_api_WebpToJpeg
func Java_com_SunnyNet_api_WebpToJpeg(envObj uintptr, clazz uintptr, webpPath, savePath uintptr, SaveQuality int64) bool {
	env := Env(envObj)
	return Api.WebpToJpeg(env.GetString(webpPath), env.GetString(savePath), int(SaveQuality))
}

/*
Java_com_SunnyNet_api_WebpToPng Webp图片转Png图片 根据文件名
*/
//export Java_com_SunnyNet_api_WebpToPng
func Java_com_SunnyNet_api_WebpToPng(envObj uintptr, clazz uintptr, webpPath, savePath uintptr) bool {
	env := Env(envObj)
	return Api.WebpToPng(env.GetString(webpPath), env.GetString(savePath))
}

/*
Java_com_SunnyNet_api_OpenDrive 开启进程代理/打开驱动
*/
//export Java_com_SunnyNet_api_OpenDrive
func Java_com_SunnyNet_api_OpenDrive(envObj uintptr, clazz uintptr, SunnyContext int64, devMode int64) bool {
	return Api.OpenDrive(int(SunnyContext), int(devMode))
}

/*
Java_com_SunnyNet_api_UnDrive 卸载驱动，仅Windows 有效【需要管理权限】执行成功后会立即重启系统,若函数执行后没有重启系统表示没有管理员权限
*/
//export Java_com_SunnyNet_api_UnDrive
func Java_com_SunnyNet_api_UnDrive(envObj uintptr, clazz uintptr, SunnyContext int64) {

	Api.UnDrive(int(SunnyContext))
}

/*
Java_com_SunnyNet_api_ProcessAddName 进程代理 添加进程名
*/
//export Java_com_SunnyNet_api_ProcessAddName
func Java_com_SunnyNet_api_ProcessAddName(envObj uintptr, clazz uintptr, SunnyContext int64, Name uintptr) {
	env := Env(envObj)
	Api.ProcessAddName(int(SunnyContext), env.GetString(Name))
}

/*
Java_com_SunnyNet_api_ProcessDelName 进程代理 删除进程名
*/
//export Java_com_SunnyNet_api_ProcessDelName
func Java_com_SunnyNet_api_ProcessDelName(envObj uintptr, clazz uintptr, SunnyContext int64, Name uintptr) {
	env := Env(envObj)
	Api.ProcessDelName(int(SunnyContext), env.GetString(Name))
}

/*
Java_com_SunnyNet_api_ProcessAddPid 进程代理 添加PID
*/
//export Java_com_SunnyNet_api_ProcessAddPid
func Java_com_SunnyNet_api_ProcessAddPid(envObj uintptr, clazz uintptr, SunnyContext, pid int64) {
	Api.ProcessAddPid(int(SunnyContext), int(pid))
}

/*
Java_com_SunnyNet_api_ProcessDelPid 进程代理 删除PID
*/
//export Java_com_SunnyNet_api_ProcessDelPid
func Java_com_SunnyNet_api_ProcessDelPid(envObj uintptr, clazz uintptr, SunnyContext, pid int64) {

	Api.ProcessDelPid(int(SunnyContext), int(pid))
}

/*
Java_com_SunnyNet_api_ProcessCancelAll 进程代理 取消全部已设置的进程名
*/
//export Java_com_SunnyNet_api_ProcessCancelAll
func Java_com_SunnyNet_api_ProcessCancelAll(envObj uintptr, clazz uintptr, SunnyContext int64) {

	Api.ProcessCancelAll(int(SunnyContext))
}

/*
Java_com_SunnyNet_api_ProcessALLName 进程代理 设置是否全部进程通过
*/
//export Java_com_SunnyNet_api_ProcessALLName
func Java_com_SunnyNet_api_ProcessALLName(envObj uintptr, clazz uintptr, SunnyContext int64, open, StopNetwork bool) {

	Api.ProcessALLName(int(SunnyContext), open, StopNetwork)
}

//================================================================================================

/*
Java_com_SunnyNet_api_GetCommonName 证书管理器 获取证书 CommonName 字段
*/
//export Java_com_SunnyNet_api_GetCommonName
func Java_com_SunnyNet_api_GetCommonName(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetCommonName(int(Context)))
}

/*
Java_com_SunnyNet_api_ExportP12 证书管理器 导出为P12
*/
//export Java_com_SunnyNet_api_ExportP12
func Java_com_SunnyNet_api_ExportP12(envObj uintptr, clazz uintptr, Context int64, path, pass uintptr) bool {
	env := Env(envObj)
	return Api.ExportP12(int(Context), env.GetString(path), env.GetString(pass))
}

/*
Java_com_SunnyNet_api_ExportPub 证书管理器 导出公钥
*/
//export Java_com_SunnyNet_api_ExportPub
func Java_com_SunnyNet_api_ExportPub(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.ExportPub(int(Context)))
}

/*
Java_com_SunnyNet_api_ExportKEY 证书管理器 导出私钥
*/
//export Java_com_SunnyNet_api_ExportKEY
func Java_com_SunnyNet_api_ExportKEY(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.ExportKEY(int(Context)))
}

/*
Java_com_SunnyNet_api_ExportCA 证书管理器 导出证书
*/
//export Java_com_SunnyNet_api_ExportCA
func Java_com_SunnyNet_api_ExportCA(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.ExportCA(int(Context)))
}

/*
Java_com_SunnyNet_api_CreateCA 证书管理器 创建证书
*/
//export Java_com_SunnyNet_api_CreateCA
func Java_com_SunnyNet_api_CreateCA(envObj uintptr, clazz uintptr, Context int64, Country, Organization, OrganizationalUnit, Province, CommonName, Locality uintptr, bits, NotAfter int64) bool {
	env := Env(envObj)
	return Api.CreateCA(int(Context), env.GetString(Country), env.GetString(Organization), env.GetString(OrganizationalUnit), env.GetString(Province), env.GetString(CommonName), env.GetString(Locality), int(bits), int(NotAfter))
}

/*
Java_com_SunnyNet_api_AddClientAuth 证书管理器 设置ClientAuth
*/
//export Java_com_SunnyNet_api_AddClientAuth
func Java_com_SunnyNet_api_AddClientAuth(envObj uintptr, clazz uintptr, Context, val int64) bool {
	//env := Env(envObj)
	return Api.AddClientAuth(int(Context), int(val))
}

/*
Java_com_SunnyNet_api_SetCipherSuites   证书管理器 设置CipherSuites
*/
//export Java_com_SunnyNet_api_SetCipherSuites
func Java_com_SunnyNet_api_SetCipherSuites(envObj uintptr, clazz uintptr, Context int64, val uintptr) bool {
	env := Env(envObj)
	return Api.SetCipherSuites(int(Context), env.GetString(val))
}

/*
Java_com_SunnyNet_api_AddCertPoolText 证书管理器 设置信任的证书 从 文本
*/
//export Java_com_SunnyNet_api_AddCertPoolText
func Java_com_SunnyNet_api_AddCertPoolText(envObj uintptr, clazz uintptr, Context int64, cer uintptr) bool {
	env := Env(envObj)
	return Api.AddCertPoolText(int(Context), env.GetString(cer))
}

/*
Java_com_SunnyNet_api_AddCertPoolPath 证书管理器 设置信任的证书 从 文件
*/
//export Java_com_SunnyNet_api_AddCertPoolPath
func Java_com_SunnyNet_api_AddCertPoolPath(envObj uintptr, clazz uintptr, Context int64, cer uintptr) bool {
	env := Env(envObj)
	return Api.AddCertPoolPath(int(Context), env.GetString(cer))
}

/*
Java_com_SunnyNet_api_GetServerName 证书管理器 取ServerName
*/
//export Java_com_SunnyNet_api_GetServerName
func Java_com_SunnyNet_api_GetServerName(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.GetServerName(int(Context)))
}

/*
Java_com_SunnyNet_api_SetServerName 证书管理器 设置ServerName
*/
//export Java_com_SunnyNet_api_SetServerName
func Java_com_SunnyNet_api_SetServerName(envObj uintptr, clazz uintptr, Context int64, name uintptr) bool {
	env := Env(envObj)
	return Api.SetServerName(int(Context), env.GetString(name))
}

/*
Java_com_SunnyNet_api_SetInsecureSkipVerify 证书管理器 设置跳过主机验证
*/
//export Java_com_SunnyNet_api_SetInsecureSkipVerify
func Java_com_SunnyNet_api_SetInsecureSkipVerify(envObj uintptr, clazz uintptr, Context int64, b bool) bool {
	//env := Env(envObj)
	return Api.SetInsecureSkipVerify(int(Context), b)
}

/*
Java_com_SunnyNet_api_LoadX509Certificate 证书管理器 载入X509证书
*/
//export Java_com_SunnyNet_api_LoadX509Certificate
func Java_com_SunnyNet_api_LoadX509Certificate(envObj uintptr, clazz uintptr, Context int64, Host, CA, KEY uintptr) bool {
	env := Env(envObj)
	return Api.LoadX509Certificate(int(Context), env.GetString(Host), env.GetString(CA), env.GetString(KEY))
}

/*
Java_com_SunnyNet_api_LoadX509KeyPair 证书管理器 载入X509证书2
*/
//export Java_com_SunnyNet_api_LoadX509KeyPair
func Java_com_SunnyNet_api_LoadX509KeyPair(envObj uintptr, clazz uintptr, Context int64, CaPath, KeyPath uintptr) bool {
	env := Env(envObj)
	return Api.LoadX509KeyPair(int(Context), env.GetString(CaPath), env.GetString(KeyPath))
}

/*
Java_com_SunnyNet_api_LoadP12Certificate 证书管理器 载入p12证书
*/
//export Java_com_SunnyNet_api_LoadP12Certificate
func Java_com_SunnyNet_api_LoadP12Certificate(envObj uintptr, clazz uintptr, Context int64, Name, Password uintptr) bool {
	env := Env(envObj)
	return Api.LoadP12Certificate(int(Context), env.GetString(Name), env.GetString(Password))
}

/*
Java_com_SunnyNet_api_RemoveCertificate 释放 证书管理器 对象
*/
//export Java_com_SunnyNet_api_RemoveCertificate
func Java_com_SunnyNet_api_RemoveCertificate(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RemoveCertificate(int(Context))
}

/*
Java_com_SunnyNet_api_CreateCertificate 创建 证书管理器 对象
*/
//export Java_com_SunnyNet_api_CreateCertificate
func Java_com_SunnyNet_api_CreateCertificate(envObj uintptr, clazz uintptr) int64 {
	//env := Env(envObj)
	return int64(Api.CreateCertificate())
}

//===================================================== go http Client ================================================

/*
Java_com_SunnyNet_api_HTTPSetH2Config HTTP 客户端 设置HTTP2指纹
*/
//export Java_com_SunnyNet_api_HTTPSetH2Config
func Java_com_SunnyNet_api_HTTPSetH2Config(envObj uintptr, clazz uintptr, Context int64, config uintptr) bool {
	env := Env(envObj)
	return Api.SetH2Config(int(Context), env.GetString(config))
}

/*
Java_com_SunnyNet_api_HTTPSetRandomTLS HTTP 客户端 设置随机使用TLS指纹
*/
//export Java_com_SunnyNet_api_HTTPSetRandomTLS
func Java_com_SunnyNet_api_HTTPSetRandomTLS(envObj uintptr, clazz uintptr, Context int64, RandomTLS bool) bool {
	//env := Env(envObj)
	return Api.HTTPSetRandomTLS(int(Context), RandomTLS)
}

/*
Java_com_SunnyNet_api_HTTPSetRedirect HTTP 客户端 设置重定向
*/
//export Java_com_SunnyNet_api_HTTPSetRedirect
func Java_com_SunnyNet_api_HTTPSetRedirect(envObj uintptr, clazz uintptr, Context int64, Redirect bool) bool {
	//env := Env(envObj)
	return Api.HTTPSetRedirect(int(Context), Redirect)
}

/*
Java_com_SunnyNet_api_HTTPGetCode HTTP 客户端 返回响应状态码
*/
//export Java_com_SunnyNet_api_HTTPGetCode
func Java_com_SunnyNet_api_HTTPGetCode(envObj uintptr, clazz uintptr, Context int64) int64 {
	//env := Env(envObj)
	return int64(Api.HTTPGetCode(int(Context)))
}

/*
Java_com_SunnyNet_api_HTTPSetCertManager HTTP 客户端 设置证书管理器
*/
//export Java_com_SunnyNet_api_HTTPSetCertManager
func Java_com_SunnyNet_api_HTTPSetCertManager(envObj uintptr, clazz uintptr, Context, CertManagerContext int64) bool {
	//env := Env(envObj)
	return Api.HTTPSetCertManager(int(Context), int(CertManagerContext))
}

/*
Java_com_SunnyNet_api_HTTPGetBody HTTP 客户端 返回响应内容
*/
//export Java_com_SunnyNet_api_HTTPGetBody
func Java_com_SunnyNet_api_HTTPGetBody(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.HTTPGetBody(int(Context)))
}

/*
Java_com_SunnyNet_api_HTTPGetRequestHeader HTTP 客户端 添加的全部协议头
*/
//export Java_com_SunnyNet_api_HTTPGetRequestHeader
func Java_com_SunnyNet_api_HTTPGetRequestHeader(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.HTTPGetRequestHeader(int(Context)))
}

/*
Java_com_SunnyNet_api_HTTPGetHeader HTTP 客户端 返回响应HTTPGetHeader
*/
//export Java_com_SunnyNet_api_HTTPGetHeader
func Java_com_SunnyNet_api_HTTPGetHeader(envObj uintptr, clazz uintptr, Context int64, name uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(Api.HTTPGetHeader(int(Context), env.GetString(name)))
}

/*
Java_com_SunnyNet_api_HTTPGetHeads HTTP 客户端 返回响应全部Heads
*/
//export Java_com_SunnyNet_api_HTTPGetHeads
func Java_com_SunnyNet_api_HTTPGetHeads(envObj uintptr, clazz uintptr, Context int64) uintptr {
	env := Env(envObj)
	return env.NewString(Api.HTTPGetHeads(int(Context)))
}

/*
Java_com_SunnyNet_api_HTTPGetBodyLen HTTP 客户端 返回响应长度
*/
//export Java_com_SunnyNet_api_HTTPGetBodyLen
func Java_com_SunnyNet_api_HTTPGetBodyLen(envObj uintptr, clazz uintptr, Context int64) int64 {
	//env := Env(envObj)
	return int64(Api.HTTPGetBodyLen(int(Context)))
}

/*
Java_com_SunnyNet_api_HTTPSendBin HTTP 客户端 发送Body
*/
//export Java_com_SunnyNet_api_HTTPSendBin
func Java_com_SunnyNet_api_HTTPSendBin(envObj uintptr, clazz uintptr, Context int64, body uintptr) {
	env := Env(envObj)
	Api.HTTPSendBin(int(Context), env.GetBytes(body))
}

/*
Java_com_SunnyNet_api_HTTPSetTimeouts HTTP 客户端 设置超时 毫秒
*/
//export Java_com_SunnyNet_api_HTTPSetTimeouts
func Java_com_SunnyNet_api_HTTPSetTimeouts(envObj uintptr, clazz uintptr, Context int64, t1 int64) {
	//env := Env(envObj)
	Api.HTTPSetTimeouts(int(Context), int(t1))
}

// Java_com_SunnyNet_api_HTTPSetServerIP
// HTTP 客户端 设置真实连接IP地址，
//
//export Java_com_SunnyNet_api_HTTPSetServerIP
func Java_com_SunnyNet_api_HTTPSetServerIP(envObj uintptr, clazz uintptr, Context int64, ServerIP uintptr) {
	env := Env(envObj)
	Api.HTTPSetServerIP(int(Context), env.GetString(ServerIP))
}

/*
Java_com_SunnyNet_api_HTTPSetProxyIP HTTP 客户端 设置代理IP 仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
*/
//
//export Java_com_SunnyNet_api_HTTPSetProxyIP
func Java_com_SunnyNet_api_HTTPSetProxyIP(envObj uintptr, clazz uintptr, Context int64, ProxyUrl uintptr) bool {
	env := Env(envObj)
	return Api.HTTPSetProxyIP(int(Context), env.GetString(ProxyUrl))
}

/*
Java_com_SunnyNet_api_HTTPSetHeader HTTP 客户端 设置协议头
*/
//export Java_com_SunnyNet_api_HTTPSetHeader
func Java_com_SunnyNet_api_HTTPSetHeader(envObj uintptr, clazz uintptr, Context int64, name, value uintptr) {
	env := Env(envObj)
	Api.HTTPSetHeader(int(Context), env.GetString(name), env.GetString(value))
}

/*
Java_com_SunnyNet_api_HTTPOpen HTTP 客户端 Open
*/
//export Java_com_SunnyNet_api_HTTPOpen
func Java_com_SunnyNet_api_HTTPOpen(envObj uintptr, clazz uintptr, Context int64, Method, URL uintptr) {
	env := Env(envObj)
	Api.HTTPOpen(int(Context), env.GetString(Method), env.GetString(URL))
}

/*
Java_com_SunnyNet_api_RemoveHTTPClient 释放 HTTP客户端
*/
//export Java_com_SunnyNet_api_RemoveHTTPClient
func Java_com_SunnyNet_api_RemoveHTTPClient(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RemoveHTTPClient(int(Context))
}

/*
Java_com_SunnyNet_api_CreateHTTPClient 创建 HTTP 客户端
*/
//export Java_com_SunnyNet_api_CreateHTTPClient
func Java_com_SunnyNet_api_CreateHTTPClient(envObj uintptr, clazz uintptr) int64 {
	//env := Env(envObj)
	return int64(Api.CreateHTTPClient())
}

//===========================================================================================

/*
Java_com_SunnyNet_api_JsonToPB JSON格式的protobuf数据转为protobuf二进制数据
*/
//export Java_com_SunnyNet_api_JsonToPB
func Java_com_SunnyNet_api_JsonToPB(envObj uintptr, clazz uintptr, bin uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.JsonToPB(env.GetString(bin)))
}

/*
Java_com_SunnyNet_api_PbToJson protobuf数据转为JSON格式
*/
//export Java_com_SunnyNet_api_PbToJson
func Java_com_SunnyNet_api_PbToJson(envObj uintptr, clazz uintptr, bin uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(Api.PbToJson(env.GetBytes(bin)))
}

//===========================================================================================

/*
Java_com_SunnyNet_api_QueuePull 队列弹出
*/
//export Java_com_SunnyNet_api_QueuePull
func Java_com_SunnyNet_api_QueuePull(envObj uintptr, clazz uintptr, name uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.QueuePull(env.GetString(name)))
}

/*
Java_com_SunnyNet_api_QueuePush 加入队列
*/
//export Java_com_SunnyNet_api_QueuePush
func Java_com_SunnyNet_api_QueuePush(envObj uintptr, clazz uintptr, name uintptr, val uintptr) {
	env := Env(envObj)
	Api.QueuePush(env.GetString(name), env.GetBytes(val))
}

/*
Java_com_SunnyNet_api_QueueLength 取队列长度
*/
//export Java_com_SunnyNet_api_QueueLength
func Java_com_SunnyNet_api_QueueLength(envObj uintptr, clazz uintptr, name uintptr) int64 {
	env := Env(envObj)
	return int64(Api.QueueLength(env.GetString(name)))
}

/*
Java_com_SunnyNet_api_QueueRelease 清空销毁队列
*/
//export Java_com_SunnyNet_api_QueueRelease
func Java_com_SunnyNet_api_QueueRelease(envObj uintptr, clazz uintptr, name uintptr) {
	env := Env(envObj)
	Api.QueueRelease(env.GetString(name))
}

/*
Java_com_SunnyNet_api_QueueIsEmpty 队列是否为空
*/
//export Java_com_SunnyNet_api_QueueIsEmpty
func Java_com_SunnyNet_api_QueueIsEmpty(envObj uintptr, clazz uintptr, name uintptr) bool {
	env := Env(envObj)
	return Api.QueueIsEmpty(env.GetString(name))
}

/*
Java_com_SunnyNet_api_CreateQueue 创建队列
*/
//export Java_com_SunnyNet_api_CreateQueue
func Java_com_SunnyNet_api_CreateQueue(envObj uintptr, clazz uintptr, name uintptr) {
	env := Env(envObj)
	Api.CreateQueue(env.GetString(name))
}

//=========================================================================================================

/*
Java_com_SunnyNet_api_SocketClientWrite TCP客户端 发送数据
*/
//export Java_com_SunnyNet_api_SocketClientWrite
func Java_com_SunnyNet_api_SocketClientWrite(envObj uintptr, clazz uintptr, Context, OutTimes int64, val uintptr) bool {
	env := Env(envObj)
	data := env.GetBytes(val)
	return Api.SocketClientWrite(int(Context), int(OutTimes), data) > 0
}

/*
Java_com_SunnyNet_api_SocketClientClose TCP客户端 断开连接
*/
//export Java_com_SunnyNet_api_SocketClientClose
func Java_com_SunnyNet_api_SocketClientClose(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.SocketClientClose(int(Context))
}

/*
Java_com_SunnyNet_api_SocketClientReceive TCP客户端 同步模式下 接收数据
*/
//export Java_com_SunnyNet_api_SocketClientReceive
func Java_com_SunnyNet_api_SocketClientReceive(envObj uintptr, clazz uintptr, Context, OutTimes int64) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.SocketClientReceive(int(Context), int(OutTimes)))
}

/*
Java_com_SunnyNet_api_SocketClientDial TCP客户端 连接
*/
//export Java_com_SunnyNet_api_SocketClientDial
func Java_com_SunnyNet_api_SocketClientDial(envObj uintptr, clazz uintptr, Context int64, addr uintptr, call uintptr, isTls, synchronous bool, ProxyUrl uintptr, CertificateContext int64, OutTime int64, OutRouterIP uintptr) bool {
	env := Env(envObj)
	if synchronous {
		return Api.SocketClientDial(int(Context), env.GetString(addr), 0, nil, isTls, true, env.GetString(ProxyUrl), int(CertificateContext), int(OutTime), env.GetString(OutRouterIP))
	}
	obj := env.NewGlobalRef(call)
	cls := env.GetObjectClass(obj)
	methodId := env.GetMethodID(cls, "onCallback", "(JJ[B)V")
	if methodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onCallback(JJ[B)V] failed")
		panic("Find Class [onCallback(JJ[B)V] failed")
	}
	f := func(Context, types int, bs []byte) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()
		_obj := Jvalue(_env.NewByteArray(bs))
		_env.CallVoidMethodA(obj, methodId, Jvalue(Context), Jvalue(types), _obj)
		_env.DeleteLocalRef(Jobject(_obj))
	}
	Java_GlobalRef_Add("SocketClient", obj, int(Context))
	return Api.SocketClientDial(int(Context), env.GetString(addr), 0, f, isTls, false, env.GetString(ProxyUrl), int(CertificateContext), int(OutTime), env.GetString(OutRouterIP))
}

/*
Java_com_SunnyNet_api_SocketClientSetBufferSize TCP客户端 置缓冲区大小
*/
//export Java_com_SunnyNet_api_SocketClientSetBufferSize
func Java_com_SunnyNet_api_SocketClientSetBufferSize(envObj uintptr, clazz uintptr, Context, BufferSize int64) bool {
	//env := Env(envObj)
	return Api.SocketClientSetBufferSize(int(Context), int(BufferSize))
}

/*
Java_com_SunnyNet_api_SocketClientGetErr TCP客户端 取错误
*/
//export Java_com_SunnyNet_api_SocketClientGetErr
func Java_com_SunnyNet_api_SocketClientGetErr(envObj uintptr, clazz uintptr, Context int64) uintptr {
	//env := Env(envObj)
	return Api.SocketClientGetErr(int(Context))
}

/*
Java_com_SunnyNet_api_RemoveSocketClient 释放 TCP客户端
*/
//export Java_com_SunnyNet_api_RemoveSocketClient
func Java_com_SunnyNet_api_RemoveSocketClient(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RemoveSocketClient(int(Context))
}

/*
Java_com_SunnyNet_api_CreateSocketClient 创建 TCP客户端
*/
//export Java_com_SunnyNet_api_CreateSocketClient
func Java_com_SunnyNet_api_CreateSocketClient(envObj uintptr, clazz uintptr) int64 {
	//env := Env(envObj)
	return int64(Api.CreateSocketClient())
}

//==================================================================================================

/*
Java_com_SunnyNet_api_WebsocketClientReceive Websocket客户端 同步模式下 接收数据 返回数据指针 失败返回0 length=返回数据长度
*/
//export Java_com_SunnyNet_api_WebsocketClientReceive
func Java_com_SunnyNet_api_WebsocketClientReceive(envObj uintptr, clazz uintptr, Context, OutTimes int64) uintptr {
	env := Env(envObj)
	Buff, messageType := Api.WebsocketClientReceive(int(Context), int(OutTimes))
	class := "com/SunnyNet/WebsocketResult"
	fun := "<init>"
	sig := "([BJ)V"
	// 获取 RedisRet 类的引用
	redisRetClass := env.FindClass(class)
	if redisRetClass == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class ["+class+"] failed")
		panic("Find Class [" + class + "] failed")
	}
	// 获取构造函数的ID
	constructor := env.GetMethodID(redisRetClass, fun, sig)
	if constructor == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Func ["+class+";"+fun+sig+"] failed")
		panic("Find Func [" + class + ";" + fun + sig + "] failed")
	}
	val := env.NewByteArray(Buff)
	// 创建 RedisRet 对象
	redisRetObject := env.NewObjectA(redisRetClass, constructor, Jvalue(val), Jvalue(messageType))
	// 释放局部引用
	env.DeleteLocalRef(val)
	return redisRetObject
}

/*
Java_com_SunnyNet_api_WebsocketReadWrite Websocket客户端  发送数据
*/
//export Java_com_SunnyNet_api_WebsocketReadWrite
func Java_com_SunnyNet_api_WebsocketReadWrite(envObj uintptr, clazz uintptr, Context int64, val uintptr, messageType int64) bool {
	env := Env(envObj)
	return Api.WebsocketReadWrite(int(Context), env.GetBytes(val), int(messageType))
}

/*
Java_com_SunnyNet_api_WebsocketClose Websocket客户端 断开
*/
//export Java_com_SunnyNet_api_WebsocketClose
func Java_com_SunnyNet_api_WebsocketClose(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.WebsocketClose(int(Context))
}

/*
Java_com_SunnyNet_api_WebsocketHeartbeat Websocket客户端 心跳设置
*/
//export Java_com_SunnyNet_api_WebsocketHeartbeat
func Java_com_SunnyNet_api_WebsocketHeartbeat(envObj uintptr, clazz uintptr, Context int64, HeartbeatTime int64, call uintptr) {
	env := Env(envObj)
	if call != 0 {
		obj := env.NewGlobalRef(call)
		if obj != 0 {
			cls := env.GetObjectClass(obj)
			if cls != 0 {
				methodId := env.GetMethodID(cls, "onHeartbeatCallback", "(J)V")
				if methodId != 0 {
					Api.WebsocketHeartbeat(int(Context), int(HeartbeatTime), 0, func(_Context int) {
						runtime.LockOSThread()
						defer runtime.UnlockOSThread()
						_env, ret := GlobalVM.AttachCurrentThread()
						if ret != JNI_OK {
							return
						}
						defer GlobalVM.DetachCurrentThread()
						_env.CallVoidMethodA(obj, methodId, Jvalue(_Context))
						return
					})
					return
				}

			}
		}
	}
	Api.WebsocketHeartbeat(int(Context), 0, 0, nil)
}

/*
Java_com_SunnyNet_api_WebsocketDial Websocket客户端 连接
*/
//export Java_com_SunnyNet_api_WebsocketDial
func Java_com_SunnyNet_api_WebsocketDial(envObj uintptr, clazz uintptr, Context int64, URL, Heads uintptr, call uintptr, synchronous bool, ProxyUrl uintptr, CertificateConText, outTime int64, OutRouterIP uintptr) bool {
	env := Env(envObj)
	if !synchronous {
		obj := env.NewGlobalRef(call)
		cls := env.GetObjectClass(obj)
		methodId := env.GetMethodID(cls, "onCallback", "(JJ[BJ)V")
		if methodId == 0 {
			env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onCallback(JJ[BJ)V] failed")
			panic("Find Class [onCallback(JJ[BJ)V] failed")
		}
		f := func(Context, types int, bs []byte, messageType int) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			_env, ret := GlobalVM.AttachCurrentThread()
			if ret != JNI_OK {
				return
			}
			defer GlobalVM.DetachCurrentThread()
			_obj := Jvalue(_env.NewByteArray(bs))
			_env.CallVoidMethodA(obj, methodId, Jvalue(Context), Jvalue(types), _obj, Jvalue(messageType))
			_env.DeleteLocalRef(Jobject(_obj))
			return
		}
		Java_GlobalRef_Add("websocket", obj, int(Context))
		return Api.WebsocketDial(int(Context), env.GetString(URL), env.GetString(Heads), 0, f, false, env.GetString(ProxyUrl), int(CertificateConText), int(outTime), env.GetString(OutRouterIP))
	}
	return Api.WebsocketDial(int(Context), env.GetString(URL), env.GetString(Heads), 0, nil, true, env.GetString(ProxyUrl), int(CertificateConText), int(outTime), env.GetString(OutRouterIP))
}

/*
Java_com_SunnyNet_api_WebsocketGetErr Websocket客户端 获取错误
*/
//export Java_com_SunnyNet_api_WebsocketGetErr
func Java_com_SunnyNet_api_WebsocketGetErr(envObj uintptr, clazz uintptr, Context int64) uintptr {
	//env := Env(envObj)
	return Api.WebsocketGetErr(int(Context))
}

/*
Java_com_SunnyNet_api_RemoveWebsocket 释放 Websocket客户端 对象
*/
//export Java_com_SunnyNet_api_RemoveWebsocket
func Java_com_SunnyNet_api_RemoveWebsocket(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RemoveWebsocket(int(Context))
}

/*
Java_com_SunnyNet_api_CreateWebsocket 创建 Websocket客户端 对象
*/
//export Java_com_SunnyNet_api_CreateWebsocket
func Java_com_SunnyNet_api_CreateWebsocket(envObj uintptr, clazz uintptr) int64 {
	//env := Env(envObj)
	return int64(Api.CreateWebsocket())
}

//==================================================================================================

/*
Java_com_SunnyNet_api_AddHttpCertificate 创建 Http证书管理器 对象 实现指定Host使用指定证书
*/
//export Java_com_SunnyNet_api_AddHttpCertificate
func Java_com_SunnyNet_api_AddHttpCertificate(envObj uintptr, clazz uintptr, host uintptr, CertManagerId, Rules int64) bool {
	env := Env(envObj)
	return Api.AddHttpCertificate(env.GetString(host), int(CertManagerId), uint8(Rules))
}

/*
Java_com_SunnyNet_api_DelHttpCertificate 删除 Http证书管理器 对象
*/
//export Java_com_SunnyNet_api_DelHttpCertificate
func Java_com_SunnyNet_api_DelHttpCertificate(envObj uintptr, clazz uintptr, host uintptr) {
	env := Env(envObj)
	Api.DelHttpCertificate(env.GetString(host))
}

//==================================================================================================

/*
Java_com_SunnyNet_api_RedisSubscribe Redis 订阅消息
*/
//export Java_com_SunnyNet_api_RedisSubscribe
func Java_com_SunnyNet_api_RedisSubscribe(envObj uintptr, clazz uintptr, Context int64, scribe uintptr, call uintptr) {
	env := Env(envObj)
	obj := env.NewGlobalRef(call)
	cls := env.GetObjectClass(obj)
	methodId := env.GetMethodID(cls, "onCallback", "(Ljava/lang/String;)V")
	if methodId == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class [onCallback(Ljava/lang/String;)V] failed")
		panic("Find Class [onCallback(Ljava/lang/String;)V] failed")
	}
	f := func(message string) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		_env, ret := GlobalVM.AttachCurrentThread()
		if ret != JNI_OK {
			return
		}
		defer GlobalVM.DetachCurrentThread()
		_obj := _env.NewString(message)
		_env.CallVoidMethodA(obj, methodId, Jvalue(_obj))
		_env.DeleteLocalRef(_obj)
	}
	Java_GlobalRef_Add("Redis", obj, int(Context))
	Api.RedisSubscribeGo(int(Context), env.GetString(scribe), f)
}

/*
Java_com_SunnyNet_api_RedisDelete Redis 删除
*/
//export Java_com_SunnyNet_api_RedisDelete
func Java_com_SunnyNet_api_RedisDelete(envObj uintptr, clazz uintptr, Context int64, key uintptr) bool {
	env := Env(envObj)
	return Api.RedisDelete(int(Context), env.GetString(key))
}

/*
Java_com_SunnyNet_api_RedisFlushDB Redis 清空当前数据库
*/
//export Java_com_SunnyNet_api_RedisFlushDB
func Java_com_SunnyNet_api_RedisFlushDB(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RedisFlushDB(int(Context))
}

/*
Java_com_SunnyNet_api_RedisFlushAll Redis 清空redis服务器
*/
//export Java_com_SunnyNet_api_RedisFlushAll
func Java_com_SunnyNet_api_RedisFlushAll(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RedisFlushAll(int(Context))
}

/*
Java_com_SunnyNet_api_RedisClose Redis 关闭
*/
//export Java_com_SunnyNet_api_RedisClose
func Java_com_SunnyNet_api_RedisClose(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RedisClose(int(Context))
}

/*
Java_com_SunnyNet_api_RedisGetInt Redis 取整数值
*/
//export Java_com_SunnyNet_api_RedisGetInt
func Java_com_SunnyNet_api_RedisGetInt(envObj uintptr, clazz uintptr, Context int64, key uintptr) int64 {
	env := Env(envObj)
	return Api.RedisGetInt(int(Context), env.GetString(key))
}

/*
Java_com_SunnyNet_api_RedisGetKeys Redis 取指定条件键名
*/
//export Java_com_SunnyNet_api_RedisGetKeys
func Java_com_SunnyNet_api_RedisGetKeys(envObj uintptr, clazz uintptr, Context int64, key uintptr) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.RedisGetKeys(int(Context), env.GetString(key)))
}

/*
Java_com_SunnyNet_api_RedisDo Redis 自定义 执行和查询命令 返回操作结果可能是值 也可能是JSON文本
*/
//export Java_com_SunnyNet_api_RedisDo
func Java_com_SunnyNet_api_RedisDo(envObj uintptr, clazz uintptr, Context int64, args uintptr) uintptr {
	env := Env(envObj)
	p, e := Api.RedisDo(int(Context), env.GetString(args))
	if e != nil {
		return javaNewRedisResultClass(env, "com/SunnyNet/RedisResult", "<init>", "(ZLjava/lang/String;Ljava/lang/String;)V", false, "", e.Error())
	}
	return javaNewRedisResultClass(env, "com/SunnyNet/RedisResult", "<init>", "(ZLjava/lang/String;Ljava/lang/String;)V", true, string(p), "")

}

/*
Java_com_SunnyNet_api_RedisGetStr Redis 取文本值
*/
//export Java_com_SunnyNet_api_RedisGetStr
func Java_com_SunnyNet_api_RedisGetStr(envObj uintptr, clazz uintptr, Context int64, key uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(Api.RedisGetStr(int(Context), env.GetString(key)))
}

/*
Java_com_SunnyNet_api_RedisGetBytes Redis 取Bytes值
*/
//export Java_com_SunnyNet_api_RedisGetBytes
func Java_com_SunnyNet_api_RedisGetBytes(envObj uintptr, clazz uintptr, Context int64, key uintptr) uintptr {
	env := Env(envObj)
	bs := Api.RedisGetBytes(int(Context), env.GetString(key))
	return env.NewByteArray(bs)
}

/*
Java_com_SunnyNet_api_RedisExists Redis 检查指定 key 是否存在
*/
//export Java_com_SunnyNet_api_RedisExists
func Java_com_SunnyNet_api_RedisExists(envObj uintptr, clazz uintptr, Context int64, key uintptr) bool {
	env := Env(envObj)
	return Api.RedisExists(int(Context), env.GetString(key))
}

/*
Java_com_SunnyNet_api_RedisSetNx Redis 设置NX 【如果键名存在返回假】
*/
//export Java_com_SunnyNet_api_RedisSetNx
func Java_com_SunnyNet_api_RedisSetNx(envObj uintptr, clazz uintptr, Context int64, key, val uintptr, expr int) bool {
	env := Env(envObj)
	return Api.RedisSetNx(int(Context), env.GetString(key), env.GetString(val), expr)
}

/*
Java_com_SunnyNet_api_RedisSet Redis 设置值
*/
//export Java_com_SunnyNet_api_RedisSet
func Java_com_SunnyNet_api_RedisSet(envObj uintptr, clazz uintptr, Context int64, key, val uintptr, expr int64) bool {
	env := Env(envObj)
	return Api.RedisSet(int(Context), env.GetString(key), env.GetString(val), int(expr))
}

/*
Java_com_SunnyNet_api_RedisSetBytes Redis 设置Bytes值
*/
//export Java_com_SunnyNet_api_RedisSetBytes
func Java_com_SunnyNet_api_RedisSetBytes(envObj uintptr, clazz uintptr, Context int64, key uintptr, val uintptr, expr int64) bool {
	env := Env(envObj)
	data := env.GetBytes(val)
	return Api.RedisSetBytes(int(Context), env.GetString(key), data, int(expr))
}

func javaNewRedisResultClass(env Env, class, fun, sig string, ok bool, value, errorValue string) uintptr {
	// 获取 RedisRet 类的引用
	redisRetClass := env.FindClass(class)
	if redisRetClass == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Class ["+class+"] failed")
		panic("Find Class [" + class + "] failed")
	}
	// 获取构造函数的ID
	constructor := env.GetMethodID(redisRetClass, fun, sig)
	if constructor == 0 {
		env.ThrowNew(env.FindClass("java/lang/RuntimeException"), "Find Func ["+class+";"+fun+sig+"] failed")
		panic("Find Func [" + class + ";" + fun + sig + "] failed")
	}
	success := JNI_TRUE
	if !ok {
		success = JNI_FALSE
	}
	val := env.NewString(value)
	err := env.NewString(errorValue)
	// 创建 RedisRet 对象
	redisRetObject := env.NewObjectA(redisRetClass, constructor, Jvalue(success), Jvalue(val), Jvalue(err))
	// 释放局部引用
	env.DeleteLocalRef(val)
	env.DeleteLocalRef(err)
	return redisRetObject
}

/*
Java_com_SunnyNet_api_RedisDial Redis 连接
*/
//export Java_com_SunnyNet_api_RedisDial
func Java_com_SunnyNet_api_RedisDial(envObj uintptr, clazz uintptr, Context int64, host, pass uintptr, db, PoolSize, MinIdleCons, DialTimeout, ReadTimeout, WriteTimeout, PoolTimeout, IdleCheckFrequency, IdleTimeout int64) uintptr {
	env := Env(envObj)
	err := make([]byte, 256)
	p := uintptr(unsafe.Pointer(&err[0]))
	public.WriteErr(errorNull, p)
	if Api.RedisDial(int(Context), env.GetString(host), env.GetString(pass), int(db), int(PoolSize), int(MinIdleCons), int(DialTimeout), int(ReadTimeout), int(WriteTimeout), int(PoolTimeout), int(IdleCheckFrequency), int(IdleTimeout), p) {
		return javaNewRedisResultClass(env, "com/SunnyNet/RedisResult", "<init>", "(ZLjava/lang/String;Ljava/lang/String;)V", true, "", "")
	}
	return javaNewRedisResultClass(env, "com/SunnyNet/RedisResult", "<init>", "(ZLjava/lang/String;Ljava/lang/String;)V", false, "", public.BytesToCString(p))
}

/*
Java_com_SunnyNet_api_RemoveRedis 释放 Redis 对象
*/
//export Java_com_SunnyNet_api_RemoveRedis
func Java_com_SunnyNet_api_RemoveRedis(envObj uintptr, clazz uintptr, Context int64) {
	//env := Env(envObj)
	Api.RemoveRedis(int(Context))
}

/*
Java_com_SunnyNet_api_CreateRedis 创建 Redis 对象
*/
//export Java_com_SunnyNet_api_CreateRedis
func Java_com_SunnyNet_api_CreateRedis(envObj uintptr, clazz uintptr) int64 {
	//env := Env(envObj)
	return int64(Api.CreateRedis())
}

/*
Java_com_SunnyNet_api_SetUdpData 设置修改UDP数据
*/
//export Java_com_SunnyNet_api_SetUdpData
func Java_com_SunnyNet_api_SetUdpData(envObj uintptr, clazz uintptr, MessageId int64, data uintptr) bool {
	env := Env(envObj)
	bs := env.GetBytes(data)
	return Api.SetUdpData(int(MessageId), bs)
}

/*
Java_com_SunnyNet_api_GetUdpData 获取UDP数据
*/
//export Java_com_SunnyNet_api_GetUdpData
func Java_com_SunnyNet_api_GetUdpData(envObj uintptr, clazz uintptr, MessageId int64) uintptr {
	env := Env(envObj)
	return env.NewByteArray(Api.GetUdpData(int(MessageId)))
}

/*
Java_com_SunnyNet_api_UdpSendToClient 指定的UDP连接 模拟服务器端向客户端主动发送数据
*/
//export Java_com_SunnyNet_api_UdpSendToClient
func Java_com_SunnyNet_api_UdpSendToClient(envObj uintptr, clazz uintptr, theology int64, data uintptr) bool {
	env := Env(envObj)
	bs := env.GetBytes(data)
	return Api.UdpSendToClient(int(theology), bs)
}

/*
Java_com_SunnyNet_api_UdpSendToServer 指定的UDP连接 模拟客户端向服务器端主动发送数据
*/
//export Java_com_SunnyNet_api_UdpSendToServer
func Java_com_SunnyNet_api_UdpSendToServer(envObj uintptr, clazz uintptr, theology int64, data uintptr) bool {
	env := Env(envObj)
	bs := env.GetBytes(data)
	return Api.UdpSendToServer(int(theology), bs)
}

// Java_com_SunnyNet_api_SetScriptCode 加载用户的脚本代码
//
//export Java_com_SunnyNet_api_SetScriptCode
func Java_com_SunnyNet_api_SetScriptCode(envObj uintptr, clazz uintptr, SunnyContext int64, code uintptr) uintptr {
	env := Env(envObj)
	return env.NewString(Api.SetScriptCode(int(SunnyContext), env.GetString(code)))
}

/*
Java_com_SunnyNet_api_SetScriptPage  设置脚本编辑器页面 需不少于8个字符
*/
//export Java_com_SunnyNet_api_SetScriptPage
func Java_com_SunnyNet_api_SetScriptPage(envObj uintptr, clazz uintptr, SunnyContext int64, Page uintptr) uintptr {
	env := Env(envObj)
	//return Api.SetScriptPage(int(SunnyContext), env.GetString(Page))
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[int(SunnyContext)]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return env.NewString("")
	}
	return env.NewString(w.SetScriptPage(env.GetString(Page)))
}

/*
Java_com_SunnyNet_api_DisableTCP  禁用TCP 仅对当前SunnyContext有效
*/
//export Java_com_SunnyNet_api_DisableTCP
func Java_com_SunnyNet_api_DisableTCP(envObj uintptr, clazz uintptr, SunnyContext int64, Disable bool) bool {
	//env := Env(envObj)
	return Api.DisableTCP(int(SunnyContext), Disable)
}

/*
Java_com_SunnyNet_api_DisableUDP  禁用TCP 仅对当前SunnyContext有效
*/
//export Java_com_SunnyNet_api_DisableUDP
func Java_com_SunnyNet_api_DisableUDP(envObj uintptr, clazz uintptr, SunnyContext int64, Disable bool) bool {
	//env := Env(envObj)
	return Api.DisableUDP(int(SunnyContext), Disable)
}

/*
Java_com_SunnyNet_api_SetRandomTLS 是否使用随机TLS指纹 仅对当前SunnyContext有效
*/
//export Java_com_SunnyNet_api_SetRandomTLS
func Java_com_SunnyNet_api_SetRandomTLS(envObj uintptr, clazz uintptr, SunnyContext int64, open bool) bool {
	//env := Env(envObj)
	return Api.SetRandomTLS(int(SunnyContext), open)
}

/*
Java_com_SunnyNet_api_SetDnsServer Dns解析服务器 默认:223.5.5.5:853
*/
//export Java_com_SunnyNet_api_SetDnsServer
func Java_com_SunnyNet_api_SetDnsServer(envObj uintptr, clazz uintptr, ServerName uintptr) {
	env := Env(envObj)
	dns.SetDnsServer(env.GetString(ServerName))
}

/*
Java_com_SunnyNet_api_SetOutRouterIP 设置数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11（全局）
*/
//export Java_com_SunnyNet_api_SetOutRouterIP
func Java_com_SunnyNet_api_SetOutRouterIP(envObj uintptr, clazz uintptr, SunnyContext int64, value uintptr) bool {
	env := Env(envObj)
	return Api.SetOutRouterIP(int(SunnyContext), env.GetString(value))
}

/*
Java_com_SunnyNet_api_RequestSetOutRouterIP 设置数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11（TCP/HTTP请求共用这个函数）
*/
//export Java_com_SunnyNet_api_RequestSetOutRouterIP
func Java_com_SunnyNet_api_RequestSetOutRouterIP(envObj uintptr, clazz uintptr, MessageId int64, value uintptr) bool {
	env := Env(envObj)
	return Api.RequestSetOutRouterIP(int(MessageId), env.GetString(value))
}

/*
Java_com_SunnyNet_api_HTTPSetOutRouterIP 设置数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11（TCP/HTTP请求共用这个函数）
*/
//export Java_com_SunnyNet_api_HTTPSetOutRouterIP
func Java_com_SunnyNet_api_HTTPSetOutRouterIP(envObj uintptr, clazz uintptr, MessageId int64, value uintptr) bool {
	env := Env(envObj)
	return Api.HTTPSetOutRouterIP(int(MessageId), env.GetString(value))
}

//export Java_com_SunnyNet_api_OnTunSetFd
func Java_com_SunnyNet_api_OnTunSetFd(JavaVM uintptr, reserved uintptr, fd int64) {
	tun.SetFd(int(fd))
}

type _GlobalRef struct {
	obj     uintptr
	Type    string
	Context int
}

var ___Java_GlobalRef_lock sync.Mutex

var ___Java_GlobalRef_map = make(map[int]_GlobalRef)
var ___Java_GlobalRef_index int = 0

func Java_GlobalRef_Add(Type string, obj uintptr, Context int) {
	___Java_GlobalRef_lock.Lock()
	defer ___Java_GlobalRef_lock.Unlock()
	___Java_GlobalRef_index++
	___Java_GlobalRef_map[___Java_GlobalRef_index] = _GlobalRef{obj: obj, Type: Type, Context: Context}
}

func goJavaInit() {
	// 固定在一个线程上
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if GlobalVM == 0 {
		return
	}
	env, ok := GlobalVM.AttachCurrentThread()
	if ok != JNI_OK {
		return
	}
	defer GlobalVM.DetachCurrentThread() // 退出时再 detach

	for {
		time.Sleep(10 * time.Second)

		for key, v := range ___Java_GlobalRef_map {
			switch v.Type {
			case "SocketClient":
				if Api.LoadSocketContext(v.Context) == nil {
					env.DeleteGlobalRef(v.obj)
					delete(___Java_GlobalRef_map, key)
				}
			case "Redis":
				if Api.LoadRedisContext(v.Context) == nil {
					env.DeleteGlobalRef(v.obj)
					delete(___Java_GlobalRef_map, key)
				}
			case "websocket":
				if Api.LoadWebSocketContext(v.Context) == nil {
					env.DeleteGlobalRef(v.obj)
					delete(___Java_GlobalRef_map, key)
				}
			case "SunnyNet":
				SunnyNet.SunnyStorageLock.Lock()
				w := SunnyNet.SunnyStorage[v.Context]
				SunnyNet.SunnyStorageLock.Unlock()
				if w == nil {
					env.DeleteGlobalRef(v.obj)
					delete(___Java_GlobalRef_map, key)
				}
			}
		}
	}
}

var _classList = make(map[string]Jclass)
var _classLock sync.Mutex

func aliasToClass(ClassAlias string) Jclass {
	_classLock.Lock()
	defer _classLock.Unlock()
	return _classList[ClassAlias]
}

// classInit 因为 FindClass 在新线程里会失败，因为新线程没有应用类加载器。 所以全局缓存
func classInit(env Env) {
	_classLock.Lock()
	defer _classLock.Unlock()

	names := []struct {
		alias, path string
	}{
		{"HTTPEvent", "com/SunnyNet/Internal/HTTPEvent"},
		{"TCPEvent", "com/SunnyNet/Internal/TCPEvent"},
		{"WebSocketEvent", "com/SunnyNet/Internal/WebSocketEvent"},
		{"UDPEvent", "com/SunnyNet/Internal/UDPEvent"},
	}

	for _, n := range names {
		local := env.FindClass(n.path)
		if local == 0 {
			panic("FindClass [" + n.path + "] failed")
		}
		_classList[n.alias] = env.NewGlobalRef(local)
	}
}

//export JNI_OnLoad
func JNI_OnLoad(JavaVM uintptr, reserved uintptr) int {
	GlobalVM = VM(JavaVM)
	env, ret := GlobalVM.GetEnv(JNI_VERSION_1_6)
	if ret != JNI_OK {
		return 0
	}
	go goJavaInit()
	classInit(env)
	return JNI_VERSION_1_6
}
