# Sunny网络中间件 SDK 文档

<div align="center">
  <svg xmlns="http://www.w3.org/2000/svg" width="210" height="20" role="img" aria-label="Platform: Windows | Linux | macOS"><title>Platform: Windows | Linux | macOS</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="210" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="57" height="20" fill="#555"/><rect x="57" width="153" height="20" fill="#007ec6"/><rect width="210" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="295" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="470">Platform</text><text x="295" y="140" transform="scale(.1)" fill="#fff" textLength="470">Platform</text><text aria-hidden="true" x="1325" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="1430">Windows | Linux | macOS</text><text x="1325" y="140" transform="scale(.1)" fill="#fff" textLength="1430">Windows | Linux | macOS</text></g></svg>
  <svg xmlns="http://www.w3.org/2000/svg" width="78" height="20" role="img" aria-label="Go: >=1.16"><title>Go: >=1.16</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="78" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="25" height="20" fill="#555"/><rect x="25" width="53" height="20" fill="#97ca00"/><rect width="78" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="135" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="150">Go</text><text x="135" y="140" transform="scale(.1)" fill="#fff" textLength="150">Go</text><text aria-hidden="true" x="505" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="430">>=1.16</text><text x="505" y="140" transform="scale(.1)" fill="#fff" textLength="430">>=1.16</text></g></svg>
  <svg xmlns="http://www.w3.org/2000/svg" width="82" height="20" role="img" aria-label="License: MIT"><title>License: MIT</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="82" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="51" height="20" fill="#555"/><rect x="51" width="31" height="20" fill="#dfb317"/><rect width="82" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="265" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="410">License</text><text x="265" y="140" transform="scale(.1)" fill="#fff" textLength="410">License</text><text aria-hidden="true" x="655" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="210">MIT</text><text x="655" y="140" transform="scale(.1)" fill="#fff" textLength="210">MIT</text></g></svg>
</div>

<div align="center">
  <h3>跨平台网络分析组件 SDK</h3>
  <p>类似 Fiddler 的网络中间件，支持 HTTP/HTTPS/WS/WSS/TCP/UDP 网络分析</p>
</div>

## 📌 重要通知

<div align="center">
  <h3><span style="color: red;">请注意: 由于本仓库历史记录太大</span></h3>
  <h3><span style="color: red;">本仓库于 2025-04-24 删除重建</span></h3>
</div>

## 🌟 项目简介

Sunny网络中间件是一个功能强大的跨平台网络分析组件，专为二次开发而设计。它提供了完整的网络流量捕获和修改功能，支持多种协议类型。

## 🚀 主要特性

- ✅ **多协议支持**: HTTP/HTTPS/WS/WSS/TCP/UDP 网络分析
- ✅ **数据获取与修改**: 可获取和修改所有协议的发送及返回数据
- ✅ **代理设置**: 可为指定连接设置独立代理
- ✅ **连接重定向**: 支持 HTTP/HTTPS/WS/WSS/TCP/TLS-TCP 链接重定向
- ✅ **数据解码**: 支持 gzip, deflate, br, zstd 解码
- ✅ **主动发送**: 支持 WS/WSS/TCP/TLS-TCP/UDP 主动发送数据
- ✅ **跨平台**: 支持 Windows、Linux 和 macOS
- ✅ **脚本支持**: 支持通过Go脚本自定义处理逻辑

## 🚦 多驱动支持

| 驱动名称 | 平台 | 127.0.0.1捕获 | 内网捕获 | 兼容性 |
|---------|------|-------------|------|-------|
| Netfilter | Windows | ✅ | ✅    | 一般 |
| Proxifier | Windows | ✅ | ✅    | 一般 |
| Tun(WinDivert) | Windows | ❌ | ✅    | 较好 |
| Tun(VPN) | Android | ✅ | ✅    | 较好 |
| Tun(utun) | MacOs | ❌ | ❌    | 较好 |
| Tun(tun) | Linux | ❌ | ❌    | 较好 |

## 📚 SDK API 参考



有关Go语言环境下使用SunnyNet的详细示例，请参考 [Go语言使用示例](README_go.md) 文档。



完整的API参考文档请查看 [API参考文档](README_api.md)。


## ⚙️ 使用说明

### 系统要求

- Windows 7 及以上版本（使用 Go 1.21 以下版本编译）
- Windows 10/11 推荐（支持最新 Go 版本）
- Linux / macOS 最新稳定版本
 
 
## 🛠 编译说明

### Windows 编译步骤

1. 安装 [TDM-GCC](https://github.com/jmeubank/tdm-gcc/releases/download/v10.3.0-tdm64-2/tdm64-gcc-10.3.0-2.exe)
2. 进入到 SunnyNet 目录
3. 执行命令 `.\BuildLibrary\BuildALL.bat`

### Linux 编译步骤

1. 确保已安装 GCC 工具链
2. 进入到 SunnyNet 目录
3. 执行命令 `.\BuildLibrary\Linux64.sh`
4. 或 执行命令 `.\BuildLibrary\Linux32.sh`

### macOS 编译步骤

1. 确保已安装 GCC 工具链
2. 进入到 SunnyNet 目录
3. 执行命令 `.\BuildLibrary\MacOs_amd64.sh`
4. 或 执行命令 `.\BuildLibrary\MacOs_arm64.sh`
 

## 📨 BUG 反馈与技术支持 
<p>项目网站: <a href="https://esunny.vip/">https://esunny.vip/</a></p>
 
<p><strong>QQ群:</strong></p>

<ul>

  <li>一群：751406884</li>

  <li>二群：545120699</li>

  <li>三群：170902713</li>

  <li>四群：1070797457</li>

</ul>

## 📦 下载资源



<p>各语言示例文件以及抓包工具下载地址:</p>

<p>🔗 <a href="https://wwxa.lanzouu.com/b02p4aet8j">https://wwxa.lanzouu.com/b02p4aet8j</a></p>

<p><strong>密码:</strong> 4h7r</p>

## ⚠️ 注意事项

1. 如需支持 Win7 系统，请使用 Go 1.21 以下版本编译，例如 go 1.20.4 版本
2. <a href="https://github.com/jmeubank/tdm-gcc/releases/download/v10.3.0-tdm64-2/tdm64-gcc-10.3.0-2.exe">编译请使用 TDM-GCC</a>