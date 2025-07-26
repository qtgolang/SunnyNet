
<div style="text-align: center;"><h3><a style="color: red;">请注意:由于本仓库历史记录太大</a></div></h3></center>
<div style="text-align: center;"><h3><a style="color: red;">本仓库于 2025-04-24 删除重建</a></div></h3></center>
 
#  <center><h3>Sunny网络中间件</center></h3></center>

---

> Sunny网络中间件 和 Fiddler 类似。 是可跨平台的网络分析组件
 ```log 
 可用于HTTP/HTTPS/WS/WSS/TCP/UDP网络分析 为二次开发量身制作
 
 支持 获取/修改 HTTP/HTTPS/WS/WSS/TCP/TLS-TCP/UDP 发送及返回数据
 
 支持 对 HTTP/HTTPS/WS/WSS 指定连接使用指定代理
 
 支持 对 HTTP/HTTPS/WS/WSS/TCP/TLS-TCP 链接重定向
 
 支持 gzip, deflate, br, zstd 解码
 
 支持 WS/WSS/TCP/TLS-TCP/UDP 主动发送数据 
 
```

---
* # 由于代码主要是做DLL使用,部分功能未封装给Go使用，请自行探索！
* # 如需支持Win7系统
* # 请使用Go1.21以下版本编译,例如 go 1.20.4版本 
* # <a href="https://github.com/jmeubank/tdm-gcc/releases/download/v10.3.0-tdm64-2/tdm64-gcc-10.3.0-2.exe">编译请使用 TDM-GCC</a>
<div style="text-align: center;"><h2><a style="color: red;">BUG 反馈</a></h2></div>
<div style="text-align: center;"><h3>QQ群:</h3></div>
<div style="text-align: center;"><h3>一群：751406884</h3></div>
<div style="text-align: center;"><h3>二群：545120699</h3></div>
<div style="text-align: center;"><h3>三群：170902713</h3></div>
<div style="text-align: center;"><h3>四群：616787804</h3></div>
<div style="text-align: center;"><h3>网址：<a href="https://esunny.vip/">https://esunny.vip/</a></h3></div>

---

### <center><h3>各语言,示例文件以及抓包工具 下载地址 </center>
<div style="text-align: center;"><h3>https://wwxa.lanzouu.com/b02p4aet8j</h3></div>
<div style="text-align: center;"><h3>密码:4h7r</h3></div>
<div style="text-align: center;"><h3></h3></div>


---
- > GoLang使用示例代码

```golang
package main

import (
	"github.com/qtgolang/SunnyNet/SunnyNet"
	"github.com/qtgolang/SunnyNet/src/public"
	"time"
	"log"
	"fmt"
)
func main() {
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
	//设置回调地址
	Sunny.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback)
	Port := 2025
	Sunny.SetPort(Port).Start()
	err := Sunny.Error
	if err != nil {
		panic(err)
	}
	fmt.Println("Run Port=", Port)
	//阻止程序退出
	select {}
}

func HttpCallback(Conn SunnyNet.ConnHTTP) {

	if Conn.Type() == public.HttpSendRequest {
		//fmt.Println(Conn.URL())
		//发起请求

		//直接响应,不让其发送请求
		//Conn.StopRequest(200, "Hello Word")

	} else if Conn.Type() == public.HttpResponseOK {
		//请求完成
		//log.Println("Call", Conn.URL())
	} else if Conn.Type() == public.HttpRequestFail {
		//请求错误
		/*	fmt.Println(Conn.Request.URL.String(), Conn.GetError())
		 */
	}
}
func WSCallback(Conn SunnyNet.ConnWebSocket) {
	log.Println("WebSocket", Conn.URL())
}
func TcpCallback(Conn SunnyNet.ConnTCP) {

	if Conn.Type() == public.SunnyNetMsgTypeTCPAboutToConnect {
		//即将连接
		mode := string(Conn.Body())
		log.Println("PID", Conn.PID(), "TCP 即将连接到:", mode, Conn.LocalAddress(), "->", Conn.RemoteAddress())
		//修改目标连接地址
		//Conn.SetNewAddress("8.8.8.8:8080")
		return
	}

	if Conn.Type() == public.SunnyNetMsgTypeTCPConnectOK {
		log.Println("PID", Conn.PID(), "TCP 连接到:", Conn.LocalAddress(), "->", Conn.RemoteAddress(), "成功")
		return
	}

	if Conn.Type() == public.SunnyNetMsgTypeTCPClose {
		log.Println("PID", Conn.PID(), "TCP 断开连接:", Conn.LocalAddress(), "->", Conn.RemoteAddress())
		return
	}
	if Conn.Type() == public.SunnyNetMsgTypeTCPClientSend {
		log.Println("PID", Conn.PID(), "发送数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
		return
	}
	if Conn.Type() == public.SunnyNetMsgTypeTCPClientReceive {
		log.Println("PID", Conn.PID(), "收到数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
		return
	}
}
func UdpCallback(Conn SunnyNet.ConnUDP) {

	if Conn.Type() == public.SunnyNetUDPTypeSend {
		//客户端向服务器端发送数据
		log.Println("PID", Conn.PID(), "发送UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
		//修改发送的数据
		//Conn.SetBody([]byte("Hello Word"))

		return
	}
	if Conn.Type() == public.SunnyNetUDPTypeReceive {
		//服务器端向客户端发送数据
		log.Println("PID", Conn.PID(), "接收UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
		//修改响应的数据
		//Conn.SetBody([]byte("Hello Word"))
		return
	}
	if Conn.Type() == public.SunnyNetUDPTypeClosed {

		log.Println("PID", Conn.PID(), "关闭UDP", Conn.LocalAddress(), Conn.RemoteAddress())
		return
	}

}
```