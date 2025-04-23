//go:build windows
// +build windows

package NFapi

import "C"
import (
	"fmt"
	. "github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	net2 "github.com/qtgolang/SunnyNet/src/iphlpapi/net"
	"github.com/qtgolang/SunnyNet/src/public"
	"github.com/shirou/gopsutil/process"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
)

func getTcpInfoPID(tcpInfo string) string {
	connections, _ := net2.Connections("tcp")
	for _, conn := range connections {
		if conn.Laddr.String() == tcpInfo {
			return strconv.Itoa(int(conn.Pid))
		}
	}
	return ""
}

var Api = new(NFApi)

var ProcessPortInt uint16
var SunnyPointer = uintptr(0)
var IsInit = false
var UdpSendReceiveFunc func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte

func threadStart() {

}

func threadEnd() {

}
func GetPid() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	GetCurrentProcessId := kernel32.NewProc("GetCurrentProcessId")
	pid, _, _ := GetCurrentProcessId.Call()
	return strconv.Itoa(int(pid))
}

var ExePid, _ = strconv.Atoi(GetPid())

func getIPV6Lan() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		ipv6 := regexp.MustCompile(`(\w+:){7}\w+`).FindString(addr.String())
		if strings.Count(ipv6, ":") == 7 {
			return ipv6
		}
	}
	return ""
}
func isLocalNetRequest(pConnInfo *NF_TCP_CONN_INFO) bool {
	if strings.Contains(pConnInfo.RemoteAddress.String(), "127.0.0.1") || strings.Contains(pConnInfo.RemoteAddress.String(), "[::1]") {
		if strings.Contains(pConnInfo.LocalAddress.String(), "0.0.0.0") {
			__localNetInfo := fmt.Sprintf("127.0.0.1:%d", int(pConnInfo.RemoteAddress.GetPort()))
			__pid := getTcpInfoPID(__localNetInfo)
			__ProcessId := strconv.Itoa(int(pConnInfo.ProcessId.Get()))
			if __pid == __ProcessId {
				return true
			}
			__localNetInfo = fmt.Sprintf("[::1]:%d", int(pConnInfo.RemoteAddress.GetPort()))
			__pid = getTcpInfoPID(__localNetInfo)
			__ProcessId = strconv.Itoa(int(pConnInfo.ProcessId.Get()))
			if __pid == __ProcessId {
				return true
			}
		}
	}
	return false
}

// 实现 tcpConnectRequest 函数，用于处理 TCP 连接请求
func tcpConnectRequest(id uint64, pConnInfo *NF_TCP_CONN_INFO) {
	if pConnInfo == nil {
		return
	}
	// 如果 ProcessPortInt 等于 0，则直接返回
	if ProcessPortInt == 0 {
		return
	}
	// 如果进程 ID 等于 ExePid，则直接返回
	if pConnInfo.ProcessId.Get() == uint32(ExePid) {
		return
	}
	// 获取进程名，并检查是否在代理名单中
	_, _, ProcessName := Api.NfgetProcessNameA(pConnInfo.ProcessId.Get())
	if ProcessName == "" {
		_pid := int32(pConnInfo.ProcessId.Get())
		arr, e := process.Processes()
		if e == nil {
			for _, v := range arr {
				if v.Pid == _pid {
					ProcessName, _ = v.Name()
					break
				}
			}
		}
	}
	Lock.Lock()
	if HookProcess == false {
		if Name[strings.ToLower(ProcessName)] == false {
			if Pid[pConnInfo.ProcessId.Get()] == false {
				Lock.Unlock()
				_, _ = Api.NfTcpDisableFiltering(id)
				return
			}
		}
	}
	Lock.Unlock()
	if IsFilterRequests(ProcessName, pConnInfo.RemoteAddress.String()) {
		return
	}
	if isLocalNetRequest(pConnInfo) {
		return
	}
	// 如果连接是 IPv6 的，则将连接的远程地址改为本地 IPv6 地址，并保存到代理列表中
	if pConnInfo.RemoteAddress.IsIpv6() {
		_, IP := pConnInfo.RemoteAddress.GetIP()
		p4 := IP.To4()
		if len(p4) != net.IPv4len {

			//这里是IPV6
			Process := &ProcessInfo{Pid: strconv.Itoa(int(pConnInfo.ProcessId.Get())), RemoteAddress: IP.String(), RemotePort: pConnInfo.RemoteAddress.GetPort(), Id: id, V6: true}
			Lock.Lock()
			Proxy[pConnInfo.LocalAddress.GetPort()] = Process
			Lock.Unlock()
			pConnInfo.RemoteAddress.SetIP(false, net.ParseIP(getIPV6Lan()))
			pConnInfo.RemoteAddress.SetPort(ProcessPortInt)
			return
		}
		//这里实际上还是IPV4
		Process := &ProcessInfo{Pid: strconv.Itoa(int(pConnInfo.ProcessId.Get())), RemoteAddress: p4.String(), RemotePort: pConnInfo.RemoteAddress.GetPort(), Id: id}

		pConnInfo.RemoteAddress.Data2[12] = 127
		pConnInfo.RemoteAddress.Data2[13] = 0
		pConnInfo.RemoteAddress.Data2[14] = 0
		pConnInfo.RemoteAddress.Data2[15] = 1
		var Port UINT16
		Port.BigEndianSet(ProcessPortInt)
		pConnInfo.RemoteAddress.Port = Port
		Lock.Lock()
		Proxy[pConnInfo.LocalAddress.GetPort()] = Process
		Lock.Unlock()
		return
	}
	// 如果连接是 IPv4 的，则将连接的远程地址改为本地 IPv4 地址，并保存到代理列表中
	_, i := pConnInfo.RemoteAddress.GetIP()
	Process := &ProcessInfo{Pid: strconv.Itoa(int(pConnInfo.ProcessId.Get())), RemoteAddress: i.String(), RemotePort: pConnInfo.RemoteAddress.GetPort(), Id: id}
	Lock.Lock()
	Proxy[pConnInfo.LocalAddress.GetPort()] = Process
	Lock.Unlock()
	pConnInfo.RemoteAddress.SetIP(true, net.ParseIP("127.0.0.1"))
	pConnInfo.RemoteAddress.SetPort(ProcessPortInt)
	return
}

func tcpConnected(id uint64, pConnInfo *NF_TCP_CONN_INFO) {
	return
}

func tcpClosed(id uint64, pConnInfo *NF_TCP_CONN_INFO) {
	if pConnInfo == nil {
		return
	}
	Lock.Lock()
	delete(Proxy, pConnInfo.LocalAddress.GetPort())
	Lock.Unlock()
	return
}

func tcpReceive(id uint64, buf *byte, len int32) {
	//_, _ = Api.NfTcpPostReceive(id, buf, len)
	return
}

func tcpSend(id uint64, buf *byte, len int32) {
	//_, _ = Api.NfTcpPostSend(id, buf, len)
	return
}

func tcpCanReceive(id uint64) {

	return
}

func tcpCanSend(id uint64) {

	return
}

// 实现 isEmpower 函数，用于检查是否有权限发送 UDP 数据
func isEmpower(id uint64) (bool, SockaddrInx, uint32, NF_UDP_CONN_INFO) {
	// 获取 UDP 连接信息
	var pConnInfo NF_UDP_CONN_INFO
	Api.NfGetUDPConnInfo(id, &pConnInfo)

	// 如果 ProcessPortInt 等于 0，则直接返回 false，并将进程 ID 和本地地址返回
	if ProcessPortInt == 0 {
		return false, pConnInfo.LocalAddress, pConnInfo.ProcessId.Get(), pConnInfo
	}
	// 如果进程 ID 等于 ExePid，则直接返回 false，并将进程 ID 和本地地址返回
	if pConnInfo.ProcessId.Get() == uint32(ExePid) {
		return false, pConnInfo.LocalAddress, pConnInfo.ProcessId.Get(), pConnInfo
	}

	// 获取进程名，并检查是否在代理名单中
	_, _, ProcessName := Api.NfgetProcessNameA(pConnInfo.ProcessId.Get())
	Lock.Lock()
	if HookProcess == false {
		if Name[strings.ToLower(ProcessName)] == false {
			if Pid[pConnInfo.ProcessId.Get()] == false {
				Lock.Unlock()
				Api.NfTcpDisableFiltering(id)
				return false, pConnInfo.LocalAddress, pConnInfo.ProcessId.Get(), pConnInfo
			}
		}
	}
	Lock.Unlock()

	// 如果有权限，则返回 true，并将本地地址和进程 ID 返回
	return true, pConnInfo.LocalAddress, pConnInfo.ProcessId.Get(), pConnInfo
}

func udpCreated(id uint64, pConnInfo *NF_UDP_CONN_INFO) {
}
func udpConnectRequest(id uint64, pConnReq *NF_UDP_CONN_REQUEST) {
}

func udpClosed(id uint64, pConnInfo *NF_UDP_CONN_INFO) {
	if pConnInfo == nil {
		return
	}
	tid := NfIdGetTid(id)
	if tid < 1 {
		return
	}
	if UdpSendReceiveFunc != nil {
		o := NfTidGetObj(tid)
		if o != nil {
			UdpSendReceiveFunc(public.SunnyNetUDPTypeClosed, o.Theoni, pConnInfo.ProcessId.Get(), pConnInfo.LocalAddress.String(), o.Send.RemoteAddress.String(), nil)
		}
	}
	NfDelTid(tid)
	return
}

func udpReceive(id uint64, RemoteAddress *SockaddrInx, buf []byte, options *NF_UDP_OPTIONS) {
	if RemoteAddress == nil {
		return
	}
	if UdpSendReceiveFunc == nil || ProcessPortInt == 0 {
		_, _ = Api.NfUdpPostReceive(id, RemoteAddress, buf, options)
		return
	}
	_, LocalAddress, Pid, pConnInfo := isEmpower(id)
	k := pConnInfo.LocalAddress.String() + RemoteAddress.String()
	o := UdpSenders.GetObj(k)
	if o == nil {
		_, _ = Api.NfUdpPostReceive(id, RemoteAddress, buf, options)
		return
	}
	UdpLock.Lock()
	if o.Receive == nil {
		o.Receive = &NfSend{Id: id, RemoteAddress: RemoteAddress.Clone(), options: options.Clone()}
	}
	UdpLock.Unlock()
	bs := UdpSendReceiveFunc(public.SunnyNetUDPTypeReceive, o.Theoni, Pid, LocalAddress.String(), RemoteAddress.String(), buf)
	if len(bs) > 0 {
		_, _ = Api.NfUdpPostReceive(id, RemoteAddress, bs, options)
	}
	return
}

// 实现 udpSend 函数，用于发送 UDP 数据
func udpSend(id uint64, RemoteAddress *SockaddrInx, buf []byte, options *NF_UDP_OPTIONS) {
	if RemoteAddress == nil {
		return
	}
	if UdpSendReceiveFunc == nil || ProcessPortInt == 0 {
		Api.NfUdpPostSend(id, RemoteAddress, buf, options)
		return
	}
	// 检查授权，并调用相应的 PID
	ok, LocalAddress, Pid, pConnInfo := isEmpower(id)
	if !ok {
		k := RemoteAddress.String() + pConnInfo.LocalAddress.String()
		o := UdpSenders.GetObj(k)
		if o == nil {
			Api.NfUdpPostSend(id, RemoteAddress, buf, options)
			return
		}
		UdpLock.Lock()
		if o.Receive == nil {
			o.Receive = &NfSend{Id: id, RemoteAddress: RemoteAddress.Clone(), options: options.Clone()}
		}
		UdpLock.Unlock()
		//这里因为是接收 所以 RemoteAddress 是本地地址 而 LocalAddress 是远程地址
		bs := UdpSendReceiveFunc(public.SunnyNetUDPTypeReceive, o.Theoni, Pid, RemoteAddress.String(), LocalAddress.String(), buf)
		if len(bs) > 0 {
			_, _ = Api.NfUdpPostSend(id, RemoteAddress, bs, options)
		}
		return
	}

	// 生成唯一键值并获取连接
	k := LocalAddress.String() + RemoteAddress.String()
	o := UdpSenders.GetObj(k)
	// 如果连接不存在，则新建连接并添加到连接池中
	if o == nil {
		Tid := atomic.AddInt64(&public.Theology, 1)
		UdpSenders.Add(k, nil, Tid, &NfSend{Id: id, RemoteAddress: RemoteAddress.Clone(), options: options.Clone()}, nil, nil, nil, nil)
		NfAddTid(id, Tid, k)
		bs := UdpSendReceiveFunc(public.SunnyNetUDPTypeSend, Tid, Pid, LocalAddress.String(), RemoteAddress.String(), buf)
		if len(bs) > 0 {
			_, _ = Api.NfUdpPostSend(id, RemoteAddress, bs, options)
		}
	} else {
		// 如果连接已建立，则发送数据
		bs := UdpSendReceiveFunc(public.SunnyNetUDPTypeSend, o.Theoni, Pid, LocalAddress.String(), RemoteAddress.String(), buf)
		if len(bs) > 0 {
			_, _ = Api.NfUdpPostSend(id, RemoteAddress, bs, options)
		}
	}
}

func udpCanReceive(id uint64) {
	return
}

func udpCanSend(id uint64) {
	return
}
