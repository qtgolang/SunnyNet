package public

import (
	"net/url"
	"strings"
)

func ProcessError(err error) string {
	if err == nil {
		return "未知错误"
	}
	er := errorToString(err)
	//fmt.Println(er)
	if !strings.Contains(er, "[SunnyNet]") {
		return "[SunnyNet]" + er
	}
	return er
}
func errorToString(err error) string {
	if err != nil {
		if netErr, ok := err.(*url.Error); ok {
			return errorToString(netErr.Err)
		} else {
			return errReplaceAll(err.Error())
		}
	}
	return "nil error"
}
func errReplaceAll(s string) string {
	if strings.Contains(s, "wsarecv: An existing connection was forcibly closed by the remote host.") {
		return "远程主机强行关闭了现有连接"
	}
	if strings.Contains(s, "The client closes the connection ") {
		return "客户端关闭了连接"
	}
	m := strings.ReplaceAll(s, "read tcp ", "读取连接 ")
	m = strings.ReplaceAll(m, "dial tcp ", "连接到 ")
	m = strings.ReplaceAll(m, "i/o timeout", "超时")
	m = strings.ReplaceAll(m, ": 超时", " 超时")
	m = strings.ReplaceAll(m, "expected declaration, found", "没有找到声明:")
	if strings.Contains(m, "EOF") {
		return "连接已关闭"
	}
	if strings.Contains(m, "Client.Timeout") {
		return "请求超时"
	}
	return m
}
