# Sunny网络中间件 SDK API 参考

## 目录

- [SunnyNet核心类](#sunnynet核心类)
  - [创建和管理SunnyNet实例](#创建和管理sunnynet实例)
  - [证书管理](#证书管理)
  - [回调设置](#回调设置)
  - [代理设置](#代理设置)
  - [TCP/UDP控制](#tcpudp控制)
  - [Socks5代理验证](#socks5代理验证)
  - [进程代理](#进程代理)
  - [网络设置](#网络设置)
  - [脚本支持](#脚本支持)
- [HTTP/HTTPS处理](#httphttps处理)
  - [请求信息获取与修改](#请求信息获取与修改)
  - [响应信息获取与修改](#响应信息获取与修改)
  - [HTTP客户端](#http客户端)
- [WebSocket处理](#websocket处理)
- [TCP处理](#tcp处理)
- [UDP处理](#udp处理)
- [证书管理器](#证书管理器)
- [其他工具函数](#其他工具函数)
  - [数据压缩/解压缩](#数据压缩解压缩)
  - [图片处理](#图片处理)
  - [Protobuf处理](#protobuf处理)
  - [队列操作](#队列操作)
  - [TCP客户端](#tcp客户端)
  - [WebSocket客户端](#websocket客户端)
  - [Redis客户端](#redis客户端)
  - [Go Map操作](#go-map操作)

## SunnyNet核心类

### 创建和管理SunnyNet实例
- `CreateSunnyNet() int` - 创建Sunny中间件对象
- `ReleaseSunnyNet(SunnyContext int) bool` - 释放SunnyNet
- `SunnyNetStart(SunnyContext int) bool` - 启动Sunny中间件
- `SunnyNetClose(SunnyContext int) bool` - 关闭停止指定Sunny中间件
- `SunnyNetSetPort(SunnyContext, Port int) bool` - 设置指定端口
- `SunnyNetError(SunnyContext int) uintptr` - 获取中间件启动时的错误信息

### 证书管理
- `SunnyNetSetCert(SunnyContext, CertificateManagerId int) bool` - 设置自定义证书
- `SunnyNetInstallCert(SunnyContext int) uintptr` - 安装证书到Windows系统内
- `ExportCert(SunnyContext int) uintptr` - 导出已设置的证书

### 回调设置
- `SunnyNetSetCallback(SunnyContext, httpCallback, tcpCallback, wsCallback, udpCallback int) bool` - 设置中间件回调地址

### 代理设置
- `SetGlobalProxy(SunnyContext int, ProxyAddress *C.char, outTime int) bool` - 设置全局上游代理
- `CompileProxyRegexp(SunnyContext int, Regexp *C.char) bool` - 设置中间件上游代理使用规则

### TCP/UDP控制
- `SunnyNetMustTcp(SunnyContext int, open bool)` - 设置中间件是否开启强制走TCP
- `SetMustTcpRegexp(SunnyContext int, Regexp *C.char, RulesAllow bool) bool` - 设置强制走TCP规则
- `DisableTCP(SunnyContext int, Disable bool) bool` - 禁用TCP
- `DisableUDP(SunnyContext int, Disable bool) bool` - 禁用UDP

### Socks5代理验证
- `SunnyNetSocket5AddUser(SunnyContext int, User, Pass *C.char) bool` - 添加S5代理需要验证的用户名
- `SunnyNetVerifyUser(SunnyContext int, open bool) bool` - 开启身份验证模式
- `SunnyNetSocket5DelUser(SunnyContext int, User *C.char) bool` - 删除S5需要验证的用户名
- `SunnyNetGetSocket5User(Theology int) uintptr` - 获取授权的S5账号

### 进程代理
- `OpenDrive(SunnyContext int, devMode int) bool` - 开始进程代理/打开驱动
- `UnDrive(SunnyContext int)` - 卸载驱动
- `ProcessAddName(SunnyContext int, Name *C.char)` - 进程代理添加进程名
- `ProcessDelName(SunnyContext int, Name *C.char)` - 进程代理删除进程名
- `ProcessAddPid(SunnyContext, pid int)` - 进程代理添加PID
- `ProcessDelPid(SunnyContext, pid int)` - 进程代理删除PID
- `ProcessCancelAll(SunnyContext int)` - 进程代理取消全部已设置的进程名
- `ProcessAddBlackName(SunnyContext int, Name *C.char)` - 进程代理添加进程名
- `ProcessDelBlackName(SunnyContext int, Name *C.char)` - 进程代理删除进程名
- `ProcessAddBlackPid(SunnyContext, pid int)` - 进程代理添加PID
- `ProcessDelBlackPid(SunnyContext, pid int)` - 进程代理删除PID
- `ProcessCancelBlackAll(SunnyContext int)` - 进程代理取消全部已设置的进程名
- `ProcessALLName(SunnyContext int, open, StopNetwork bool)` - 进程代理设置是否全部进程通过

### 网络设置
- `SetOutRouterIP(SunnyContext int, value *C.char) bool` - 设置数据出口IP
- `SetIeProxy(SunnyContext int) bool` - 设置IE代理
- `CancelIEProxy(SunnyContext int) bool` - 取消设置的IE代理
- `SetDnsServer(ServerName *C.char)` - 设置Dns解析服务器

### 脚本支持
- `SetScriptCode(SunnyContext int, code uintptr, length int) uintptr` - 加载用户的脚本代码
- `SetScriptCall(SunnyContext int, LOG, SAVE uintptr)` - 设置脚本代码的回调函数
- `SetScriptPage(SunnyContext int, Page *C.char) uintptr` - 设置脚本编辑器页面

## HTTP/HTTPS处理

### 请求信息获取与修改
- `GetRequestProto(MessageId int) uintptr` - 获取HTTPS请求的协议版本
- `SetRequestUrl(MessageId int, URI *C.char) bool` - 修改HTTP/S当前请求的URL
- `GetRequestHeader(MessageId int, name *C.char) uintptr` - 获取HTTP/S当前请求数据中的指定协议头
- `SetRequestHeader(MessageId int, name, val *C.char)` - 设置HTTP/S请求体中的协议头
- `DelRequestHeader(MessageId int, name *C.char)` - 删除HTTP/S请求数据中指定的协议头
- `GetRequestAllHeader(MessageId int) uintptr` - 获取HTTP/S当前请求数据全部协议头
- `SetRequestALLHeader(MessageId int, val *C.char)` - 设置HTTP/S请求体中的全部协议头
- `GetRequestCookie(MessageId int, name *C.char) uintptr` - 获取HTTP/S当前请求数据中指定的Cookie
- `GetRequestALLCookie(MessageId int) uintptr` - 获取HTTP/S当前请求全部Cookie
- `SetRequestCookie(MessageId int, name, val *C.char)` - 修改、设置HTTP/S当前请求数据中指定Cookie
- `SetRequestAllCookie(MessageId int, val *C.char)` - 修改、设置HTTP/S当前请求数据中的全部Cookie
- `GetRequestBodyLen(MessageId int) int` - 获取HTTP/S当前请求POST提交数据长度
- `GetRequestBody(MessageId int) uintptr` - 获取HTTP/S当前POST提交数据
- `SetRequestData(MessageId int, data uintptr, dataLen int) bool` - 设置、修改HTTP/S当前请求POST提交数据
- `IsRequestRawBody(MessageId int) bool` - 此请求是否为原始body
- `RawRequestDataToFile(MessageId int, saveFileName uintptr, len int) bool` - 获取HTTP/S当前POST提交数据原始Data

### 响应信息获取与修改
- `GetResponseProto(MessageId int) uintptr` - 获取HTTPS响应的协议版本
- `GetResponseHeader(MessageId int, name *C.char) uintptr` - 获取HTTP/S当前返回数据中指定的协议头
- `SetResponseHeader(MessageId int, name *C.char, val *C.char)` - 修改、设置HTTP/S当前返回数据中的指定协议头
- `DelResponseHeader(MessageId int, name *C.char)` - 删除HTTP/S返回数据中指定的协议头
- `GetResponseAllHeader(MessageId int) uintptr` - 获取HTTP/S当前返回全部协议头
- `SetResponseAllHeader(MessageId int, value *C.char)` - 修改、设置HTTP/S当前返回数据中的全部协议头
- `GetResponseBodyLen(MessageId int) int` - 获取HTTP/S当前返回数据长度
- `GetResponseBody(MessageId int) uintptr` - 获取HTTP/S当前返回数据
- `SetResponseData(MessageId int, data uintptr, dataLen int) bool` - 设置、修改HTTP/S当前请求返回数据
- `GetResponseStatusCode(MessageId int) int` - 获取HTTP/S返回的状态码
- `GetResponseStatus(MessageId int) uintptr` - 获取HTTP/S返回的状态文本
- `SetResponseStatus(MessageId, code int)` - 修改HTTP/S返回的状态码

### HTTP客户端
- `CreateHTTPClient() int` - 创建HTTP客户端
- `RemoveHTTPClient(Context int)` - 释放HTTP客户端
- `HTTPOpen(Context int, Method, URL *C.char)` - HTTP客户端Open
- `HTTPSetHeader(Context int, name, value *C.char)` - HTTP客户端设置协议头
- `HTTPSetProxyIP(Context int, ProxyUrl *C.char) bool` - HTTP客户端设置代理IP
- `HTTPSetServerIP(Context int, ServerIP *C.char)` - HTTP客户端设置真实连接IP地址
- `HTTPSetTimeouts(Context int, t1 int)` - HTTP客户端设置超时
- `HTTPSendBin(Context int, body uintptr, bodyLength int)` - HTTP客户端发送Body
- `HTTPGetHeads(Context int) uintptr` - HTTP客户端返回响应全部Heads
- `HTTPGetHeader(Context int, name *C.char) uintptr` - HTTP客户端返回响应HTTPGetHeader
- `HTTPGetRequestHeader(Context int) uintptr` - HTTP客户端添加的全部协议头
- `HTTPGetBodyLen(Context int) int` - HTTP客户端返回响应长度
- `HTTPGetBody(Context int) uintptr` - HTTP客户端返回响应内容
- `HTTPGetCode(Context int) int` - HTTP客户端返回响应状态码
- `HTTPSetCertManager(Context, CertManagerContext int) bool` - HTTP客户端设置证书管理器
- `HTTPSetRedirect(Context int, Redirect bool) bool` - HTTP客户端设置重定向
- `HTTPSetRandomTLS(Context int, RandomTLS bool) bool` - HTTP客户端设置随机使用TLS指纹
- `HTTPSetH2Config(Context int, config *C.char) bool` - HTTP客户端设置HTTP2指纹

## WebSocket处理
- `GetWebsocketBodyLen(MessageId int) int` - 获取WebSocket消息长度
- `GetWebsocketBody(MessageId int) uintptr` - 获取WebSocket消息
- `SetWebsocketBody(MessageId int, data uintptr, dataLen int) bool` - 修改WebSocket消息
- `SendWebsocketBody(Theology, MessageType int, data uintptr, dataLen int) bool` - 主动向Websocket服务器发送消息
- `SendWebsocketClientBody(Theology, MessageType int, data uintptr, dataLen int) bool` - 主动向Websocket客户端发送消息
- `CloseWebsocket(Theology int) bool` - 主动关闭Websocket

## TCP处理
- `SetTcpBody(MessageId, MsgType int, data uintptr, dataLen int) bool` - 修改TCP消息数据
- `SetTcpAgent(MessageId int, ProxyUrl *C.char, outTime int) bool` - 给当前TCP连接设置代理
- `TcpCloseClient(theology int) bool` - 根据唯一ID关闭指定的TCP连接
- `SetTcpConnectionIP(MessageId int, address *C.char) bool` - 给指定的TCP连接修改目标连接地址
- `TcpSendMsg(theology int, data uintptr, dataLen int) int` - 指定的TCP连接模拟客户端向服务器端主动发送数据
- `TcpSendMsgClient(theology int, data uintptr, dataLen int) int` - 指定的TCP连接模拟服务器端向客户端主动发送数据

## UDP处理
- `GetUdpData(MessageId int) uintptr` - 获取UDP数据
- `SetUdpData(MessageId int, val uintptr, valLen int) bool` - 设置修改UDP数据
- `UdpSendToClient(theology int, data uintptr, dataLen int) bool` - 指定的UDP连接模拟服务器端向客户端主动发送数据
- `UdpSendToServer(theology int, data uintptr, dataLen int) bool` - 指定的UDP连接模拟客户端向服务器端主动发送数据

## 证书管理器
- `CreateCertificate() int` - 创建证书管理器对象
- `RemoveCertificate(Context int)` - 释放证书管理器对象
- `LoadX509Certificate(Context int, Host, CA, KEY *C.char) bool` - 载入X509证书
- `LoadX509KeyPair(Context int, CaPath, KeyPath *C.char) bool` - 载入X509证书2
- `LoadP12Certificate(Context int, Name, Password *C.char) bool` - 载入p12证书
- `SetInsecureSkipVerify(Context int, b bool) bool` - 设置跳过主机验证
- `SetServerName(Context int, name *C.char) bool` - 设置ServerName
- `GetServerName(Context int) uintptr` - 取ServerName
- `AddCertPoolPath(Context int, cer *C.char) bool` - 设置信任的证书从文件
- `AddCertPoolText(Context int, cer *C.char) bool` - 设置信任的证书从文本
- `SetCipherSuites(Context int, val *C.char) bool` - 设置CipherSuites
- `AddClientAuth(Context, val int) bool` - 设置ClientAuth
- `CreateCA(Context int, Country, Organization, OrganizationalUnit, Province, CommonName, Locality *C.char, bits, NotAfter int) bool` - 创建证书
- `ExportCA(Context int) uintptr` - 导出证书
- `ExportKEY(Context int) uintptr` - 导出私钥
- `ExportPub(Context int) uintptr` - 导出公钥
- `ExportP12(Context int, path, pass *C.char) bool` - 导出为P12
- `GetCommonName(Context int) uintptr` - 获取证书CommonName字段

## 其他工具函数
- `GetSunnyVersion() uintptr` - 获取SunnyNet版本
- `Free(ptr uintptr)` - 释放指针
- `SetHTTPRequestMaxUpdateLength(SunnyContext int, i int64) bool` - 设置HTTP请求提交数据的最大长度
- `RandomRequestCipherSuites(MessageId int) bool` - 随机设置请求CipherSuites
- `SetRequestHTTP2Config(MessageId int, h2Config *C.char) bool` - 设置HTTP 2.0请求指纹配置
- `GetRequestClientIp(MessageId int) uintptr` - 获取当前HTTP/S请求由哪个IP发起
- `GetResponseServerAddress(MessageId int) uintptr` - 获取HTTP/S相应的服务器地址
- `SetRequestOutTime(MessageId int, times int)` - 请求设置超时-毫秒
- `GetMessageNote(MessageId int) uintptr` - 获取请求中的注释
- `BytesToInt(data uintptr, dataLen int) int` - 将Go int的Bytes转为int

### 数据压缩/解压缩
- `GzipUnCompress(data uintptr, dataLen int) uintptr` - Gzip解压缩
- `GzipCompress(data uintptr, dataLen int) uintptr` - Gzip压缩
- `BrUnCompress(data uintptr, dataLen int) uintptr` - br解压缩
- `BrCompress(data uintptr, dataLen int) uintptr` - br压缩
- `BrotliCompress(data uintptr, dataLen int) uintptr` - br压缩(别名)
- `ZSTDDecompress(data uintptr, dataLen int) uintptr` - ZSTD解压缩
- `ZSTDCompress(data uintptr, dataLen int) uintptr` - ZSTD压缩
- `ZlibUnCompress(data uintptr, dataLen int) uintptr` - Zlib解压缩
- `ZlibCompress(data uintptr, dataLen int) uintptr` - Zlib压缩
- `DeflateUnCompress(data uintptr, dataLen int) uintptr` - Deflate解压缩
- `DeflateCompress(data uintptr, dataLen int) uintptr` - Deflate压缩

### 图片处理
- `WebpToJpegBytes(data uintptr, dataLen int, SaveQuality int) uintptr` - Webp图片转JEG图片字节数组
- `WebpToPngBytes(data uintptr, dataLen int) uintptr` - Webp图片转Png图片字节数组
- `WebpToJpeg(webpPath, savePath *C.char, SaveQuality int) bool` - Webp图片转JEG图片根据文件名
- `WebpToPng(webpPath, savePath *C.char) bool` - Webp图片转Png图片根据文件名

### Protobuf处理
- `JsonToPB(bin uintptr, binLen int) uintptr` - JSON格式的protobuf数据转为protobuf二进制数据
- `PbToJson(bin uintptr, binLen int) uintptr` - protobuf数据转为JSON格式

### 队列操作
- `CreateQueue(name *C.char)` - 创建队列
- `QueueIsEmpty(name *C.char) bool` - 队列是否为空
- `QueueRelease(name *C.char)` - 清空销毁队列
- `QueueLength(name *C.char) int` - 取队列长度
- `QueuePush(name *C.char, val uintptr, valLen int)` - 加入队列
- `QueuePull(name *C.char) uintptr` - 队列弹出

### TCP客户端
- `CreateSocketClient() int` - 创建TCP客户端
- `RemoveSocketClient(Context int)` - 释放TCP客户端
- `SocketClientDial(Context int, addr *C.char, call int, isTls, synchronous bool, ProxyUrl *C.char, CertificateConText int, OutTime int, OutRouterIP *C.char) bool` - TCP客户端连接
- `SocketClientReceive(Context, OutTimes int) uintptr` - TCP客户端同步模式下接收数据
- `SocketClientWrite(Context, OutTimes int, val uintptr, valLen int) int` - TCP客户端发送数据
- `SocketClientClose(Context int)` - TCP客户端断开连接
- `SocketClientSetBufferSize(Context, BufferSize int) bool` - TCP客户端置缓冲区大小
- `SocketClientGetErr(Context int) uintptr` - TCP客户端取错误

### WebSocket客户端
- `CreateWebsocket() int` - 创建WebSocket客户端对象
- `RemoveWebsocket(Context int)` - 释放WebSocket客户端对象
- `WebsocketDial(Context int, URL, Heads *C.char, call int, synchronous bool, ProxyUrl *C.char, CertificateConText, outTime int, OutRouterIP *C.char) bool` - Websocket客户端连接
- `WebsocketReceive(Context, OutTimes int) uintptr` - Websocket客户端同步模式下接收数据
- `WebsocketReadWrite(Context int, val uintptr, valLen int, messageType int) bool` - Websocket客户端发送数据
- `WebsocketClose(Context int)` - Websocket客户端断开
- `WebsocketHeartbeat(Context, HeartbeatTime, call int)` - Websocket客户端心跳设置
- `WebsocketGetErr(Context int) uintptr` - Websocket客户端获取错误

### Redis客户端
- `CreateRedis() int` - 创建Redis对象
- `RemoveRedis(Context int)` - 释放Redis对象
- `RedisDial(Context int, host, pass *C.char, db, PoolSize, MinIdleCons, DialTimeout, ReadTimeout, WriteTimeout, PoolTimeout, IdleCheckFrequency, IdleTimeout int, error uintptr) bool` - Redis连接
- `RedisSetBytes(Context int, key *C.char, val uintptr, valLen int, expr int) bool` - Redis设置Bytes值
- `RedisSet(Context int, key, val *C.char, expr int) bool` - Redis设置值
- `RedisSetNx(Context int, key, val *C.char, expr int) bool` - Redis设置NX
- `RedisExists(Context int, key *C.char) bool` - Redis检查指定key是否存在
- `RedisGetBytes(Context int, key *C.char) uintptr` - Redis取Bytes值
- `RedisGetStr(Context int, key *C.char) uintptr` - Redis取文本值
- `RedisDo(Context int, args *C.char, error uintptr) uintptr` - Redis自定义执行和查询命令
- `RedisGetInt(Context int, key *C.char) int64` - Redis取整数值
- `RedisGetKeys(Context int, key *C.char) uintptr` - Redis取指定条件键名
- `RedisDelete(Context int, key *C.char) bool` - Redis删除
- `RedisFlushDB(Context int)` - Redis清空当前数据库
- `RedisFlushAll(Context int)` - Redis清空redis服务器
- `RedisClose(Context int)` - Redis关闭
- `RedisSubscribe(Context int, scribe *C.char, call int, nc bool) bool` - Redis订阅消息

### Go Map操作
- `CreateKeys() int` - GoMap创建
- `RemoveKeys(KeysHandle int)` - GoMap删除GoMap
- `KeysDelete(KeysHandle int, name *C.char)` - GoMap删除
- `KeysRead(KeysHandle int, name *C.char) uintptr` - GoMap读取字符串/字节数组
- `KeysWrite(KeysHandle int, name *C.char, val uintptr, length int)` - GoMap写字节数组
- `KeysWriteFloat(KeysHandle int, name *C.char, val float64)` - GoMap写浮点数
- `KeysReadFloat(KeysHandle int, name *C.char) float64` - GoMap读浮点数
- `KeysWriteLong(KeysHandle int, name *C.char, val int64)` - GoMap写长整数
- `KeysReadLong(KeysHandle int, name *C.char) int64` - GoMap读长整数
- `KeysWriteInt(KeysHandle int, name *C.char, val int)` - GoMap写整数
- `KeysReadInt(KeysHandle int, name *C.char) int` - GoMap读整数
- `KeysEmpty(KeysHandle int)` - GoMap清空
- `KeysGetCount(KeysHandle int) int` - GoMap取数量
- `KeysGetJson(KeysHandle int) uintptr` - GoMap转为JSON字符串
- `KeysWriteStr(KeysHandle int, name *C.char, val uintptr, len int)` - GoMap写字符串