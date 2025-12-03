package loop

import (
	"errors"
	"net"
	"sync"
)

// 读写锁用于保护端口映射表
var (
	mu      sync.RWMutex
	portMap = map[uint16]uint16{} // key: 本地端口，value: 远端端口
)

// extractPorts 从连接中提取本地端口和远端端口
func extractPorts(conn net.Conn) (uint16, uint16, error) {
	tcpLocal, ok1 := conn.LocalAddr().(*net.TCPAddr)   // 本地 TCP 地址
	tcpRemote, ok2 := conn.RemoteAddr().(*net.TCPAddr) // 对端 TCP 地址
	if !ok1 || !ok2 {
		return 0, 0, errors.New("connection is not TCP")
	}
	return uint16(tcpLocal.Port), uint16(tcpRemote.Port), nil
}

// Add 记录一条新的连接端口映射（local -> remote）
func Add(conn net.Conn) {
	localPort, remotePort, err := extractPorts(conn)
	if err != nil {
		return
	}

	mu.Lock()
	portMap[localPort] = remotePort
	mu.Unlock()
}

// Un 移除一条连接的端口映射
func Un(conn net.Conn) {
	localPort, _, err := extractPorts(conn)
	if err != nil {
		return
	}

	mu.Lock()
	delete(portMap, localPort)
	mu.Unlock()
}

// Check 检测当前连接是否形成“反向环路”
// 规则：
//
//	已记录过一条 remote -> local
//	当前出现 local -> remote
//	并且其中一端是 ServerPort（用于限定只检测从服务器端口发起的回连）
//	则判定为环路
func Check(conn net.Conn, ServerPort uint16) (rs bool) {
	localPort, remotePort, err := extractPorts(conn)
	if err != nil {
		return false
	}

	mu.RLock()
	defer mu.RUnlock()

	// 查找是否存在 remotePort 作为本地端口的旧映射
	if mapped, ok := portMap[remotePort]; ok {
		// 要求 mapped == localPort 表示出现端口对调
		// 再额外要求 localPort == ServerPort 避免端口误判
		if mapped == localPort && localPort == ServerPort {
			return true
		}
	}

	return false
}
