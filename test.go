package main

import "C"
import (
	"fmt"
	"github.com/qtgolang/SunnyNet/SunnyNet"
	"github.com/qtgolang/SunnyNet/src/encoding/hex"
	"github.com/qtgolang/SunnyNet/src/public"
	"log"
)

func Test() {
	var Sunny = SunnyNet.NewSunny()
	/*
		//载入自定义证书
		cert := SunnyNet.NewCertManager()
		ok := cert.LoadP12Certificate("C:\\Users\\Qin\\Desktop\\Cert\\ca6afc5aa40fcbd3.p12", "GXjc75IRAO0T")
		fmt.Println("载入P12:", ok)
		fmt.Println("证书名称：", cert.GetCommonName())

		//给指定域名使用这个证书
		Sunny.AddHttpCertificate("api.vlightv.com", cert, SunnyNet.HTTPCertRules_Request)

	*/

	/*
		log := func(Context int, info ...any) {
			fmt.Println("x脚本日志", fmt.Sprintf("%v", info))
		}
		save := func(Context int, code []byte) {
			//在这里将code代码 储存到文件，下次启动时，载入恢复
		}
		Sunny.SetScriptCall(log, save)
		//载入上次保存的脚本代码
		Sunny.SetScriptCode(string(GoScriptCode.DefaultCode))
	*/

	/*
		//设置全局上游代理
		Sunny.SetGlobalProxy("socket://192.168.31.1:4321", 60000)

		//指定IP或域名不使用全局的上游代理
		Sunny.CompileProxyRegexp("127.0.0.1;[::1];192.168.*;*.baidu.com")
	*/

	/*
		//开启强制走TCP,开启后 https 将不会解密 直接转发数据流量
		Sunny.MustTcp(true)
	*/
	/*
		//禁止TCP，所有TCP流量将直接断开连接
		Sunny.DisableTCP(true)
	*/

	/*
		//设置强制走TCP规则，使用这个函数后 就不要使用 Sunny.MustTcp(true) 否则这个函数无效
		Sunny.SetMustTcpRegexp("tpstelemetry.tencent.com", true)
	*/
	/*
		//使用驱动抓包 (两个驱动各有特点自行尝试,哪个能用/好用 用哪个)
		Sunny.OpenDrive(true)  // 使用 NFAPI 驱动
		Sunny.OpenDrive(false) // 使用 Proxifier 驱动 不支持32位操作系统，不支持UDP数据捕获

		Sunny.ProcessAddName("gamemon.des") //添加指定进程名称
		Sunny.ProcessDelName("gamemon.des") //删除已添加的指定进程名称
		Sunny.ProcessAddPid(1122)		    //添加指定进程PID
		Sunny.ProcessDelPid(1122)		    //删除已添加的指定进程PID
		Sunny.ProcessCancelAll()			//删除已添加的所有进程名称/PID
		Sunny.ProcessALLName(true, false)	//捕获全部进程开始后，添加进程名称-PID无效
	*/
	//Sunny.SetMustTcpRegexp("124.221.161.122", true)
	//Sunny.SetGlobalProxy("socket://127.0.0.1:2026", 60000)
	//Sunny.SetOutRouterIP("192.168.31.154")
	//Sunny.SetMustTcpRegexp("shopr-cnlive.mcoc-cdn.cn", false)
	//Sunny.MustTcp(true)
	//设置回调地址
	Sunny.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback)
	Port := 2025
	Sunny.SetPort(Port).Start()
	//fmt.Println(Sunny.SetIEProxy())
	err := Sunny.Error
	if err != nil {
		panic(err)
	}
	fmt.Println("Run Port=", Port)
	//阻止程序退出
	select {}
}
func HttpCallback(Conn SunnyNet.ConnHTTP) {
	switch Conn.Type() {
	case public.HttpSendRequest: //发起请求
		fmt.Println("发起请求", Conn.Proto())
		//Conn.SetResponseBody([]byte("123456"))
		//直接响应,不让其发送请求
		//Conn.StopRequest(200, "Hello Word")
		return
	case public.HttpResponseOK: //请求完成
		bs := Conn.GetResponseBody()
		log.Println("请求完成", Conn.GetResponseProto(), Conn.URL(), len(bs), Conn.GetResponseHeader())
		return
	case public.HttpRequestFail: //请求错误
		//fmt.Println(time.Now(), Conn.URL(), Conn.Error())
		return
	}
}
func WSCallback(Conn SunnyNet.ConnWebSocket) {
	return
	switch Conn.Type() {
	case public.WebsocketConnectionOK: //连接成功
		log.Println("PID", Conn.PID(), "Websocket 连接成功:", Conn.URL())
		return
	case public.WebsocketUserSend: //发送数据
		if Conn.MessageType() < 5 {
			log.Println("PID", Conn.PID(), "Websocket 发送数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
		}
		return
	case public.WebsocketServerSend: //收到数据
		if Conn.MessageType() < 5 {
			log.Println("PID", Conn.PID(), "Websocket 收到数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
		}
		return
	case public.WebsocketDisconnect: //连接关闭
		log.Println("PID", Conn.PID(), "Websocket 连接关闭", Conn.URL())
		return
	default:
		return
	}
}
func TcpCallback(Conn SunnyNet.ConnTCP) {
	return
	switch Conn.Type() {
	case public.SunnyNetMsgTypeTCPAboutToConnect: //即将连接
		mode := string(Conn.Body())
		log.Println("PID", Conn.PID(), "TCP 即将连接到:", mode, Conn.LocalAddress(), "->", Conn.RemoteAddress())
		//修改目标连接地址
		//Conn.SetNewAddress("8.8.8.8:8080")
		return
	case public.SunnyNetMsgTypeTCPConnectOK: //连接成功
		log.Println("PID", Conn.PID(), "TCP 连接到:", Conn.LocalAddress(), "->", Conn.RemoteAddress(), "成功")
		return
	case public.SunnyNetMsgTypeTCPClose: //连接关闭
		log.Println("PID", Conn.PID(), "TCP 断开连接:", Conn.LocalAddress(), "->", Conn.RemoteAddress())
		return
	case public.SunnyNetMsgTypeTCPClientSend: //客户端发送数据
		log.Println("PID", Conn.PID(), "TCP 发送数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
		return
	case public.SunnyNetMsgTypeTCPClientReceive: //客户端收到数据

		log.Println("PID", Conn.PID(), "收到数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
		return
	default:
		return
	}
}
func UdpCallback(Conn SunnyNet.ConnUDP) {

	switch Conn.Type() {
	case public.SunnyNetUDPTypeSend: //客户端向服务器端发送数据

		log.Println("PID", Conn.PID(), "发送UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
		//修改发送的数据
		//Conn.SetBody([]byte("Hello Word"))

		return
	case public.SunnyNetUDPTypeReceive: //服务器端向客户端发送数据
		log.Println("PID", Conn.PID(), "接收UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
		//修改响应的数据
		//Conn.SetBody([]byte("Hello Word"))
		return
	case public.SunnyNetUDPTypeClosed: //关闭会话
		log.Println("PID", Conn.PID(), "关闭UDP", Conn.LocalAddress(), Conn.RemoteAddress())
		return
	}

}
