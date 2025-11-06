//go:build windows
// +build windows

package Proxifier

/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -lws2_32
#include "Proxifier.hpp"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/ProcessCheck"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"
)

var HandleClientConn func(net.Conn)
var myPid = os.Getpid()

func Write(hPipe C.HANDLE, bs []byte) {
	l := len(bs)
	if l < 1 {
		return
	}
	b := C.CString(string(bs))
	C.ProxifierWriteFile(hPipe, b, C.DWORD(l))
	C.free(unsafe.Pointer(b))
}

var mu sync.Mutex

//export Call
func Call(hPipe C.HANDLE, raw uintptr) {
	__pid := int(binary.LittleEndian.Uint16(CStringToBytes(raw+0x4EC, 2)))
	path := wcharPtrToString(raw + 8)

	mu.Lock()
	Handle := HandleClientConn
	if __pid == myPid {
		mu.Unlock()
		return
	}
	if Handle == nil {
		mu.Unlock()
		return
	}
	mu.Unlock()
	fileName := filepath.Base(path)
	if ProcessCheck.CheckPidByName(int32(__pid), fileName) {
		return
	}
	family := int16(binary.LittleEndian.Uint16(CStringToBytes(raw+0x419, 2)))
	if family == 0 {
		WriteData := make([]byte, 1020)
		WriteData[0] = 0xfc
		WriteData[1] = 0x3
		WriteData[4] = 0x1
		WriteData[0x3f8] = 0x1
		Write(hPipe, WriteData)
		//fmt.Println(hex.Dump(CStringToBytes(raw, 0x534)))
		return
	}
	if family != 2 && family != 23 {
		return
	}
	domain := wcharPtrToString(raw + 528)
	port := int(binary.BigEndian.Uint16(CStringToBytes(raw+0x419+2, 2)))
	if domain == "" {
		bs := make([]byte, 0)
		if family == 23 {
			bs = CStringToBytes(raw+0x419+8, 16)
		} else {
			bs = CStringToBytes(raw+0x419+4, 4)
		}
		ip := net.IP(bs)
		if ip.To4() != nil {
			domain = ip.String()
		} else {
			domain = ip.String()
		}
	}
	if port < 1 || port > 65535 {
		return
	}
	if Info.IsFilterRequests(fileName, domain) {
		return
	}
	var listener net.Listener
	var err error
	WriteData := make([]byte, 1020)
	//固定标志
	WriteData[0] = 0xfc
	WriteData[1] = 0x03
	WriteData[9] = 0x00
	ISV6 := family == 23
	if ISV6 {
		WriteData[8] = 0x17
		listener, err = net.Listen("tcp", "[::1]:")
	} else {
		WriteData[8] = 0x02
		listener, err = net.Listen("tcp", "127.0.0.1:")
	}
	if err != nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		connChan := make(chan net.Conn)
		go func() {
			if er := recover(); er != nil {
			}
			conn, _ := listener.Accept()
			_ = listener.Close()
			connChan <- conn
		}()

		select {
		case conn := <-connChan:
			if conn != nil {
				ip := net.ParseIP(domain)
				if ip == nil {
					ip = net.ParseIP("[" + domain + "]")
				}
				_ISV6 := false
				p4 := ip.To4()
				p6 := ip.To16()
				if p4 == nil && p6 != nil {
					_ISV6 = true
				}
				var obj = &proxyProcessInfo{listener: listener, RemoteAddress: domain, RemotePort: uint16(port), V6: _ISV6, Pid: fmt.Sprintf("%d", __pid)}
				connLocalAddr := conn.RemoteAddr().(*net.TCPAddr)
				connPort := uint16(connLocalAddr.Port)
				ProcessCheck.AddDevObj(connPort, obj)
				_ = conn.SetDeadline(time.Time{})
				Handle(conn)
				_ = conn.Close()
				ProcessCheck.DelDevObj(connPort)
			}
			_ = listener.Close()
			return
		case <-ctx.Done():
			_ = listener.Close()
			return
		}
	}()
	binary.BigEndian.PutUint16(WriteData[10:], uint16(listener.Addr().(*net.TCPAddr).Port))
	if ISV6 {
		//[::1]
		WriteData[0x1f] = 0x01

		WriteData[0x3f0] = 0x17
	} else {
		//127.0.0.1
		WriteData[12] = 0x7f
		WriteData[13] = 0x00
		WriteData[14] = 0x00
		WriteData[15] = 0x01
		WriteData[0x3f0] = 0x02
	}
	//不知道什么玩意
	WriteData[1012] = 0x06
	WriteData[1016] = 0x02
	Write(hPipe, WriteData)

}

type proxyProcessInfo struct {
	Id            uint64
	Pid           string
	RemoteAddress string
	RemotePort    uint16
	V6            bool
	listener      net.Listener
}

func (p *proxyProcessInfo) GetRemoteAddress() string {
	return p.RemoteAddress
}
func (p *proxyProcessInfo) GetRemotePort() uint16 {
	return p.RemotePort
}
func (p *proxyProcessInfo) GetPid() string {
	return p.Pid
}
func (p *proxyProcessInfo) IsV6() bool {
	return p.V6
}
func (p *proxyProcessInfo) ID() uint64 {
	return p.Id
}
func (p *proxyProcessInfo) Close() error {
	mu.Lock()
	if p.listener != nil {
		_ = p.listener.Close()
	}
	p.listener = nil
	mu.Unlock()
	return nil
}
func wcharPtrToString(ptr uintptr) string {
	var length int
	// 计算宽字符的长度
	for {
		wchar := *(*C.wchar_t)(unsafe.Pointer(ptr + uintptr(length)*unsafe.Sizeof(C.wchar_t(0))))
		if wchar == 0 {
			break
		}
		length++
	}

	// 创建一个 Go 字符串切片
	runes := make([]rune, length)

	for i := 0; i < length; i++ {
		runes[i] = rune(*(*C.wchar_t)(unsafe.Pointer(ptr + uintptr(i)*unsafe.Sizeof(C.wchar_t(0)))))
	}

	return string(runes)
}

func CStringToBytes(r uintptr, dataLen int) []byte {
	data := make([]byte, 0)
	if r == 0 || dataLen == 0 {
		return data
	}
	for i := 0; i < dataLen; i++ {
		data = append(data, *(*byte)(unsafe.Pointer(r + uintptr(i))))
	}
	return data
}
func IsInit() bool {
	return int(C.ProxifierIsInit()) == 1 || HandleClientConn != nil
}
func SetHandle(Handle func(conn net.Conn)) bool {
	res := 0
	mu.Lock()
	if Handle == nil {
		res = int(C.StopProxifier())
	} else {
		res = int(C.StartProxifier())
	}
	HandleClientConn = Handle
	mu.Unlock()
	return res == 1
}

func init() {
	go func() {
		C.ProxifierInit(C.int(myPid))
	}()
}
