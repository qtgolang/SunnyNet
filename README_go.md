# SunnyNet Go语言使用示例

本文档提供了SunnyNet网络中间件在Go语言环境下的各种使用场景示例，基于test.go文件中的实际代码示例。

## 目录

1. [基础HTTP代理](#基础http代理)
2. [双向认证处理](#双向认证处理)
3. [WebSocket处理](#websocket处理)
4. [TCP连接处理](#tcp连接处理)
5. [UDP数据处理](#udp数据处理)
6. [证书管理](#证书管理)
7. [代理设置](#代理设置)
8. [驱动模式](#驱动模式)
9. [DNS解析模式](#dns解析模式)
10. [其他功能](#其他功能)

## 基础HTTP代理

### 简单HTTP代理服务器

```go
package main

import (
    "fmt"
    "log"
    "github.com/qtgolang/SunnyNet/SunnyNet"
    "github.com/qtgolang/SunnyNet/src/public"
)

func main() {
    var Sunny = SunnyNet.NewSunny()
    
    // 设置回调地址
    Sunny.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback)
    
    Port := 2021
    Sunny.SetPort(Port).Start()
    
    err := Sunny.Error
    if err != nil {
        panic(err)
        return
    }
    
    fmt.Println("Run Port=", Port)
    
    // 阻止程序退出
    select {}
}

func HttpCallback(Conn SunnyNet.ConnHTTP) {
    switch Conn.Type() {
    case public.HttpSendRequest: // 发起请求
        fmt.Println("发起请求", Conn.URL(), Conn.Proto(), Conn.GetProcessName())
        // 直接响应,不让其发送请求
        // Conn.StopRequest(200, "Hello Word")
        return
    case public.HttpResponseOK: // 请求完成
        bs := Conn.GetResponseBody()
        log.Println("请求完成", Conn.GetResponseProto(), Conn.URL(), len(bs), Conn.GetResponseHeader())
        return
    case public.HttpRequestFail: // 请求错误
        // fmt.Println(time.Now(), Conn.URL(), Conn.Error())
        return
    }
}

func WSCallback(Conn SunnyNet.ConnWebSocket) {
    switch Conn.Type() {
    case public.WebsocketConnectionOK: // 连接成功
        log.Println("PID", Conn.PID(), "Websocket 连接成功:", Conn.URL())
        return
    case public.WebsocketUserSend: // 发送数据 
        log.Println("PID", Conn.PID(), "Websocket 发送数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        return
    case public.WebsocketServerSend: // 收到数据
        log.Println("PID", Conn.PID(), "Websocket 收到数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        return
    case public.WebsocketDisconnect: // 连接关闭
        log.Println("PID", Conn.PID(), "Websocket 连接关闭", Conn.URL())
        return
    default:
        return
    }
}

func TcpCallback(Conn SunnyNet.ConnTCP) {
    switch Conn.Type() {
    case public.SunnyNetMsgTypeTCPAboutToConnect: // 即将连接
        mode := string(Conn.Body())
        log.Println("PID", Conn.PID(), "TCP 即将连接到:", mode, Conn.LocalAddress(), "->", Conn.RemoteAddress())
        // 修改目标连接地址
        // Conn.SetNewAddress("8.8.8.8:8080")
        return
    case public.SunnyNetMsgTypeTCPConnectOK: // 连接成功
        // log.Println("PID", Conn.PID(), "TCP 连接到:", Conn.LocalAddress(), "->", Conn.RemoteAddress(), "成功")
        return
    case public.SunnyNetMsgTypeTCPClose: // 连接关闭
        // log.Println("PID", Conn.PID(), "TCP 断开连接:", Conn.LocalAddress(), "->", Conn.RemoteAddress())
        return
    case public.SunnyNetMsgTypeTCPClientSend: // 客户端发送数据
        // log.Println("PID", Conn.PID(), "TCP 发送数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), len(Conn.Body()))
        return
    case public.SunnyNetMsgTypeTCPClientReceive: // 客户端收到数据
        // log.Println("PID", Conn.PID(), "收到数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), len(Conn.Body()))
        return
    default:
        return
    }
}

func UdpCallback(Conn SunnyNet.ConnUDP) {
    switch Conn.Type() {
    case public.SunnyNetUDPTypeSend: // 客户端向服务器端发送数据
        log.Println("PID", Conn.PID(), Conn.GetProcessName(), "发送UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        // 修改发送的数据
        // Conn.SetBody([]byte("Hello Word"))
        return
    case public.SunnyNetUDPTypeReceive: // 服务器端向客户端发送数据
        log.Println("PID", Conn.PID(), Conn.GetProcessName(), "接收UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        // 修改响应的数据
        // Conn.SetBody([]byte("Hello Word"))
        return
    case public.SunnyNetUDPTypeClosed: // 关闭会话
        log.Println("PID", Conn.PID(), Conn.GetProcessName(), "关闭UDP", Conn.LocalAddress(), Conn.RemoteAddress())
        return
    }
}
```
 
### 双向认证处理

```go
package main

import (
    "fmt"
    "log"
    "github.com/qtgolang/SunnyNet/SunnyNet"
    "github.com/qtgolang/SunnyNet/src/public"
)

func main() {
    var Sunny = SunnyNet.NewSunny() 
    // 载入自定义证书 
    cert := SunnyNet.NewCertManager()
    ok := cert.LoadP12Certificate("./test.p12", "password")
    if ok {
        fmt.Println("证书名称：", cert.GetCommonName())
        // 给指定域名使用这个证书
        // SunnyNet.HTTPCertRules_ResponseAndRequest 按需调整 SunnyNet.HTTPCertRules_Request、SunnyNet.HTTPCertRules_Response
        Sunny.AddHttpCertificate("api.test.com", cert, SunnyNet.HTTPCertRules_ResponseAndRequest)
    }else{
        panic("载入P12证书失败")
    } 
    // 设置回调地址
    Sunny.SetGoCallback(HttpCallback, nil, nil, nil)
    
    Port := 2021
    Sunny.SetPort(Port).Start()
    
    err := Sunny.Error
    if err != nil {
        panic(err)
        return
    }
    
    fmt.Println("Run Port=", Port)
    
    // 阻止程序退出
    select {}
}

func HttpCallback(Conn SunnyNet.ConnHTTP) {
    switch Conn.Type() {
    case public.HttpSendRequest: // 发起请求
        fmt.Println("发起请求", Conn.URL(), Conn.Proto(), Conn.GetProcessName())
        // 直接响应,不让其发送请求
        // Conn.StopRequest(200, "Hello Word")
        return
    case public.HttpResponseOK: // 请求完成
        bs := Conn.GetResponseBody()
        log.Println("请求完成", Conn.GetResponseProto(), Conn.URL(), len(bs), Conn.GetResponseHeader())
        return
    case public.HttpRequestFail: // 请求错误
        // fmt.Println(time.Now(), Conn.URL(), Conn.Error())
        return
    }
}

```
 

## WebSocket处理

### WebSocket连接监控

```go
func WSCallback(Conn SunnyNet.ConnWebSocket) {
    switch Conn.Type() {
    case public.WebsocketConnectionOK: // 连接成功
        log.Println("PID", Conn.PID(), "Websocket 连接成功:", Conn.URL())
        // 连接建立时可以主动发送消息
        // Conn.SendToClient(1, []byte("Hello from proxy!"))
        return
    case public.WebsocketUserSend: // 发送数据
        log.Println("PID", Conn.PID(), "Websocket 发送数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        // 可以修改发送的消息
        // if Conn.MessageType() == 1 { // 文本消息
        //     modifiedMsg := []byte("modified: " + string(Conn.Body()))
        //     Conn.SetBody(modifiedMsg)
        // }
		// 拦截此消息,不让此消息发出
		// Conn.SetBody(nil)
		// 主动向客户端发送数据(可将Conn储存到全局随时调用)
		// Conn.SendToClient([]byte("模拟服务端发送到客户端的消息"))
		// 主动向服务端发送数据(可将Conn储存到全局随时调用)
		// Conn.SendToServer([]byte("模拟客户端发送到服务端的消息"))
        return
    case public.WebsocketServerSend: // 收到数据
        log.Println("PID", Conn.PID(), "Websocket 收到数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        // 可以修改接收的消息
        // if Conn.MessageType() == 1 { // 文本消息
        //     modifiedMsg := []byte("modified: " + string(Conn.Body()))
        //     Conn.SetBody(modifiedMsg)
        // }
		// 拦截此消息,不让此消息发出
		// Conn.SetBody(nil)
		// 主动向客户端发送数据(可将Conn储存到全局随时调用)
		// Conn.SendToClient([]byte("模拟服务端发送到客户端的消息"))
		// 主动向服务端发送数据(可将Conn储存到全局随时调用)
		// Conn.SendToServer([]byte("模拟客户端发送到服务端的消息"))
        return
    case public.WebsocketDisconnect: // 连接关闭
        log.Println("PID", Conn.PID(), "Websocket 连接关闭", Conn.URL())
        return
    default:
        return
    }
}
```

## TCP连接处理

### TCP连接监控

```go
func TcpCallback(Conn SunnyNet.ConnTCP) {
    switch Conn.Type() {
    case public.SunnyNetMsgTypeTCPAboutToConnect: // 即将连接
        mode := string(Conn.Body())
        log.Println("PID", Conn.PID(), "TCP 即将连接到:", mode, Conn.LocalAddress(), "->", Conn.RemoteAddress())
        // 修改目标连接地址
        // Conn.SetNewAddress("8.8.8.8:8080")
        // 设置TCP代理
        // Conn.SetAgent("socks5://user:pass@127.0.0.1:1080", 30)
        return
    case public.SunnyNetMsgTypeTCPConnectOK: // 连接成功
        // log.Println("PID", Conn.PID(), "TCP 连接到:", Conn.LocalAddress(), "->", Conn.RemoteAddress(), "成功")
        return
    case public.SunnyNetMsgTypeTCPClose: // 连接关闭
        // log.Println("PID", Conn.PID(), "TCP 断开连接:", Conn.LocalAddress(), "->", Conn.RemoteAddress())
        return
    case public.SunnyNetMsgTypeTCPClientSend: // 客户端发送数据
        // log.Println("PID", Conn.PID(), "TCP 发送数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), len(Conn.Body()))
        // 修改发送的数据
        // Conn.SetBody([]byte("modified data"))
        // 拦截此消息,不让此消息发出
        // Conn.SetBody(nil)
        // 主动向客户端发送数据(可将Conn储存到全局随时调用)
        // Conn.SendToClient([]byte("模拟服务端发送到客户端的消息"))
        // 主动向服务端发送数据(可将Conn储存到全局随时调用)
        // Conn.SendToServer([]byte("模拟客户端发送到服务端的消息"))
        return
    case public.SunnyNetMsgTypeTCPClientReceive: // 客户端收到数据
        // log.Println("PID", Conn.PID(), "收到数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), len(Conn.Body()))
        // 修改接收的数据
        // Conn.SetBody([]byte("modified response"))
        // 拦截此消息,不让此消息发出
        // Conn.SetBody(nil)
        // 主动向客户端发送数据(可将Conn储存到全局随时调用)
        // Conn.SendToClient([]byte("模拟服务端发送到客户端的消息"))
        // 主动向服务端发送数据(可将Conn储存到全局随时调用)
        // Conn.SendToServer([]byte("模拟客户端发送到服务端的消息"))
        return
    default:
        return
    }
}
```

## UDP数据处理

### UDP监控

```go
func UdpCallback(Conn SunnyNet.ConnUDP) {
    switch Conn.Type() {
    case public.SunnyNetUDPTypeSend: // 客户端向服务器端发送数据
        log.Println("PID", Conn.PID(), Conn.GetProcessName(), "发送UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        // 修改发送的数据
        // Conn.SetBody([]byte("Hello Word"))
        // 拦截此消息,不让此消息发出
        // Conn.SetBody(nil)
        // 主动向客户端发送数据(可将Conn储存到全局随时调用)
        // Conn.SendToClient([]byte("模拟服务端发送到客户端的消息"))
        // 主动向服务端发送数据(可将Conn储存到全局随时调用)
        // Conn.SendToServer([]byte("模拟客户端发送到服务端的消息"))
        return
    case public.SunnyNetUDPTypeReceive: // 服务器端向客户端发送数据
        log.Println("PID", Conn.PID(), Conn.GetProcessName(), "接收UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        // 修改响应的数据
        // Conn.SetBody([]byte("Hello Word"))
		// 拦截此消息,不让此消息发出
		// Conn.SetBody(nil)
        // 主动向客户端发送数据(可将Conn储存到全局随时调用)
        // Conn.SendToClient([]byte("模拟服务端发送到客户端的消息"))
		// 主动向服务端发送数据(可将Conn储存到全局随时调用)
		// Conn.SendToServer([]byte("模拟客户端发送到服务端的消息"))
        return
    case public.SunnyNetUDPTypeClosed: // 关闭会话
        log.Println("PID", Conn.PID(), Conn.GetProcessName(), "关闭UDP", Conn.LocalAddress(), Conn.RemoteAddress())
        return
    }
}
```

## 证书管理

### 创建自定义证书

```go  

var Sunny = SunnyNet.NewSunny()

Cert := SunnyNet.NewCertManager()

//从文件载入证书并且设置为SunnyNet根证书
if Cert.LoadP12Certificate("./SunnyNet.p12", "1234567890") {
    Sunny.SetCert(Cert.Context())
}

success := Cert.CreateCA(
    "CN",              // Country
    "My Organization", // Organization
    "IT",              // Organizational Unit
    "Beijing",         // Province
    "My Custom CA",    // Common Name
    "Beijing",         // Locality
    2048,              // bits
    365,               // NotAfter (days)
    )

if success {
    fmt.Println("CA证书创建成功")
    // 导出证书
    caPem := Cert.ExportCA()
    keyPem := Cert.ExportKEY()

    fmt.Println("CA证书:", caPem)
    fmt.Println("私钥:", keyPem)
    //导出到文件
	Cert.ExportP12("./SunnyNet.p12", "1234567890")
}

```

## 代理设置

### 全局上游代理

```go
// 设置全局上游代理
Sunny.SetGlobalProxy("socket://192.168.31.1:4321", 60000)

// 指定IP或域名不使用全局的上游代理
Sunny.CompileProxyRegexp("127.0.0.1;[::1];192.168.*;*.baidu.com")
 
```

### 强制TCP模式

```go
// 开启强制走TCP,开启后 https 将不会解密 直接转发数据流量
Sunny.MustTcp(true)

// 或者设置强制走TCP规则，使用这个函数后 就不要使用 Sunny.MustTcp(true) 否则这个函数无效
Sunny.SetMustTcpRegexp("tpstelemetry.tencent.com", true)

// 设置TCP规则模式 (true: 规则内走TCP, false: 规则外走TCP)
Sunny.SetMustTcpRegexp("*.google.com;*.github.com", true)
```

### 禁用TCP/UDP

```go
// 禁止TCP，所有TCP流量将直接断开连接
Sunny.DisableTCP(true)

// 禁止UDP，所有UDP流量将直接断开连接
Sunny.DisableUDP(true)
```

## 驱动模式

### 使用驱动抓包

```go
// 使用驱动抓包 (两个驱动各有特点自行尝试,哪个能用/好用 用哪个)
// 0=Proxifier,1=NFAPI,2=Tun
success := Sunny.OpenDrive(2)
if success {
    fmt.Println("驱动启动成功")
} else {
    fmt.Println("驱动启动失败")
}

// 添加指定进程名称进行抓包
Sunny.ProcessAddName("chrome.exe")
Sunny.ProcessAddName("firefox.exe")

// 删除已添加的指定进程名称
Sunny.ProcessDelName("firefox.exe")

// 添加指定进程PID进行抓包
Sunny.ProcessAddPid(1234)

// 删除已添加的指定进程PID
Sunny.ProcessDelPid(1234)

// 删除已添加的所有进程名称/PID
Sunny.ProcessCancelAll()

// 捕获全部进程开始后，添加进程名称-PID无效
Sunny.ProcessALLName(true, false)

// 设置数据出口IP
Sunny.SetOutRouterIP("192.168.1.100")
```

## DNS解析模式

### DNS使用的3种模式

```go
/*
DNS使用的3种模式
当你设置了全局上游代理，或对请求单独设置了代理的情况下，使用DNS模式，视情况而定,来选择设置
情况1.你没有使用全局上游代理，也没有对请求单独设置代理,这种情况下，没什么好说的，无论你设置的是那种模式，都只会使用本机的DNS进行解析!
情况2.你使用了全局上游代理或请求单独设置代理,这种情况下，你设置以下3种模式会有区别
     1.使用     本地解析    模式，你要访问的目标地址，通过你本地DNS解析出的IP，可能会被服务器拒绝连接。这时候你需要尝试 远程解析/远程服务器解析
     2.使用     远程解析    模式，你所使用的代理服务器可能存在无法解析的情况。这时你应该尝试 远程服务器解析
     3.使用  远程服务器解析  模式，远程服务器解析 会使用你设置的代理，连接到远程DNS服务器进行查询并且解析，可能会导致首次访问变慢

     Sunny.SetDnsServer("local")             //本地解析
     Sunny.SetDnsServer("remote")            //远程解析
     Sunny.SetDnsServer("223.5.5.5:853")     //远程服务器解析
*/

// 设置DNS解析模式示例
Sunny.SetDnsServer("223.5.5.5:853")  // 使用远程DNS服务器解析
```

## 其他功能

### 设置出口IP

```go
// 设置数据出口IP
Sunny.SetOutRouterIP("192.168.31.154")

// 为请求单独设置出口IP(在回调中设置)
// Conn.SetOutRouterIP("192.168.31.154")

```

### 脚本功能

```go
log := func(Context int, info ...any) {
    fmt.Println("x脚本日志", fmt.Sprintf("%v", info))
}
save := func(Context int, code []byte) {
    // 在这里将code代码 储存到文件，下次启动时，载入恢复
}
Sunny.SetScriptCall(log, save)
// 载入上次保存的脚本代码
Sunny.SetScriptCode(string(GoScriptCode.DefaultCode))

// 设置脚本页面
Sunny.SetScriptPage("MyScriptEditor")
//设置后 使用浏览器访问 http://127.0.0.1:{prot}/MyScriptEditor 来使用脚本
```

### 数据压缩/解压缩

```go
// Gzip压缩
originalData := []byte("Hello, World!")
//压缩 仅支持 gzip、br、deflate、zstd、zlib 算法
compressedData := Compress.CompressAuto("gzip", originalData)
if compressedData != nil {
     fmt.Printf("压缩后长度: %d\n", len(compressedData))
     // 解压缩 仅支持 gzip、br、deflate、zstd、zlib 算法
     decompressedData := Compress.UnCompressAuto("gzip", compressedData)
     if decompressedData != nil {
          fmt.Printf("解压后数据: %s\n", string(decompressedData))
     } 
}
```

以上示例展示了SunnyNet在各种场景下的使用方法，开发者可以根据具体需求选择合适的示例进行参考和修改。[有任何Bug和示例错误可以随时反馈](README.md)
