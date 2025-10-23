package CrossCompiled

import (
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun"
	Tun2 "github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/Tun"
	"github.com/qtgolang/SunnyNet/src/iphlpapi/net"
	"github.com/shirou/gopsutil/process"
	"os"
	"strconv"
)

type NFAPI struct {
	TCP   Tun2.TcpFunc
	UDP   Tun2.UdpFunc
	Sunny Tun2.Interface
}
type Pr struct {
	TCP   Tun2.TcpFunc
	UDP   Tun2.UdpFunc
	Sunny Tun2.Interface
}
type Tun struct {
	TCP   Tun2.TcpFunc
	UDP   Tun2.UdpFunc
	Sunny Tun2.Interface
}

func (t Tun) Install() bool {
	return tun.Install()
}

func (t Tun) IsRun() bool {
	return tun.IsRun()
}

func (t Tun) SetHandle() bool {
	tun.SetHandle(t.TCP, t.UDP, t.Sunny)
	return true
}

func (t Tun) Run() bool {
	return tun.Run()
}

func (t Tun) Close() bool {
	return tun.Close()
}

func (t Tun) Name() string {
	return tun.Name()
}

func (t Tun) UnInstall() bool {
	return tun.UnInstall()
}

const DrvPr = 0
const DrvNF = 1
const DrvTun = 2

// GetTcpInfoPID 用于获取指定 TCP 连接信息的 PID
func GetTcpInfoPID(tcpInfo string, SunnyPort int) string {
	connections, _ := net.Connections("tcp")
	for _, conn := range connections {
		if conn.Laddr.String() == tcpInfo {
			return strconv.Itoa(int(conn.Pid))
		}
	}
	return ""
}

// GetPidName 用于获取指定 PID 的进程名称
func GetPidName(pid int32) string {
	p, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, err := p.Name()
	if err != nil {
		return ""
	}
	return name
}

var myPid = int32(os.Getpid())

// IsLoopRequest 是否环路请求
func IsLoopRequest(Port string, SunnyPort int) bool {
	p, _ := strconv.Atoi(Port)
	if p == 0 {
		return false
	}
	_ConnPort := uint32(p)
	_SunnyPort := uint32(SunnyPort)
	connections, _ := net.ConnectionsPid("tcp", myPid)
	for _, conn := range connections {
		if conn.Raddr.Port == _SunnyPort {
			if conn.Laddr.Port == _ConnPort {
				return true
			}
		}

	}
	return false
}

func LoopRemotePort(Srt string) uint32 {
	connections, _ := net.ConnectionsPid("tcp", myPid)
	for _, conn := range connections {
		if conn.Laddr.String() == Srt {
			return conn.Raddr.Port
		}
	}
	return 0
}
