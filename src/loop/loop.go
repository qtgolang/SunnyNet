package loop

import (
	"errors" //错误定义
	"net"    //网络连接
	"sync"   //并发锁
)

// 错误定义
var (
	errNotTCP = errors.New("connection is not TCP") //非TCP连接错误
)

// 端口映射与过滤表
var (
	mu      sync.RWMutex              //读写锁保护共享map
	portMap = make(map[uint16]uint16) //key: 本地端口 value: 远端端口
	filter  = make(map[uint16]bool)   //需要过滤的端口集合
)

// extractPorts 从连接中提取本地端口和远端端口
func extractPorts(conn net.Conn) (local uint16, remote uint16, err error) {
	if conn == nil { //连接为空直接报错
		return 0, 0, errNotTCP
	}
	tcpLocal, ok1 := conn.LocalAddr().(*net.TCPAddr)   //本地TCP地址
	tcpRemote, ok2 := conn.RemoteAddr().(*net.TCPAddr) //对端TCP地址
	if !ok1 || !ok2 {                                  //非TCP连接
		return 0, 0, errNotTCP
	}
	return uint16(tcpLocal.Port), uint16(tcpRemote.Port), nil
}

// Add 记录一条新的连接端口映射(local -> remote)
func Add(conn net.Conn) {
	localPort, remotePort, err := extractPorts(conn) //提取端口
	if err != nil {
		return
	}
	mu.Lock()                       //加写锁
	portMap[localPort] = remotePort //写入映射
	mu.Unlock()                     //解锁
}

// Un 移除一条连接的端口映射
func Un(conn net.Conn) {
	localPort, _, err := extractPorts(conn) //提取本地端口
	if err != nil {
		return
	}
	mu.Lock()                  //加写锁
	delete(portMap, localPort) //删除映射
	mu.Unlock()                //解锁
}

// AddLoopFilter 加入过滤集合
func AddLoopFilter(t uint16) {
	mu.Lock()        //加写锁
	filter[t] = true //记录远端端口
	mu.Unlock()      //解锁
}

// UnLoopFilter 从过滤集合移除
func UnLoopFilter(t uint16) {
	mu.Lock()         //加写锁
	delete(filter, t) //删除远端端口
	mu.Unlock()       //解锁
}

// IsFiltered 判断端口是否在过滤集合中
func IsFiltered(port uint16) bool {
	mu.RLock()            //加读锁
	_, ok := filter[port] //查询端口
	mu.RUnlock()          //解锁
	return ok
}

// IsFilterConn 判断端口是否在过滤集合中
func IsFilterConn(conn net.Conn) bool {
	_, remotePort, err := extractPorts(conn) //提取本地端口
	if err != nil {
		return false
	}
	mu.RLock()                  //加读锁
	ok, _ := filter[remotePort] //查询端口
	mu.RUnlock()                //解锁
	return ok
}

// Check 检测当前连接是否形成“反向环路”
// 判定条件：
// 1) 已记录过 remote -> local 的映射
// 2) 当前出现 local -> remote（端口对调）
// 3) localPort == ServerPort（只限定从服务器端口发起的回连）
func Check(conn net.Conn, ServerPort uint16) bool {
	localPort, remotePort, err := extractPorts(conn) //提取端口
	if err != nil {
		return false
	}
	mu.RLock()                        //加读锁
	mapped, ok := portMap[remotePort] //查找是否存在remotePort作为本地端口的旧映射
	isFilter := filter[remotePort]
	mu.RUnlock() //解锁

	if !ok { //未命中映射
		return false
	}
	if isFilter || mapped != localPort {
		return false
	}

	if localPort != ServerPort { //不是限定的服务器端口
		return false
	}
	return true
}
