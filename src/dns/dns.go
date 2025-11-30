package dns

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
)

var dnsConfig = &tls.Config{
	ClientSessionCache: tls.NewLRUClientSessionCache(32),
	InsecureSkipVerify: true,
}
var dnsList = make(map[string]*rsIps)
var dnsLock sync.Mutex
var dnsTools = make(map[string]*tools)
var dnsServer = "localhost" //223.5.5.5:853  阿里云公共DNS解析服务器
const dnsServerLocal = "localhost"

func init() {
	go clean()
}

func newResolver(proxy string, outRouterIP *net.TCPAddr, Dial func(network, address string, outRouterIP *net.TCPAddr) (net.Conn, error)) *net.Resolver {
	var dialer net.Dialer
	_default_ := &net.Resolver{
		PreferGo: true,
		Dial: func(context context.Context, network_, address string) (net.Conn, error) {
			dnsLock.Lock()
			_dnsServer := dnsServer + ""
			dnsLock.Unlock()
			var conn net.Conn
			var err error

			if _dnsServer == "" {
				if proxy == "" {
					return dialer.DialContext(context, network_, address)
				}
				//使用代理进行查询，代理仅支持TCP
				return Dial("tcp", address, outRouterIP)
			}
			_tlsTCP := strings.HasSuffix(_dnsServer, ":853")

			if _tlsTCP {
				if proxy == "" {
					conn, err = dialer.DialContext(context, network_, _dnsServer)
				} else {
					//使用代理连接到自定义DNS服务器，代理仅支持TCP
					conn, err = Dial("tcp", _dnsServer, outRouterIP)
				}
				if err != nil {
					return nil, err
				}

				_ = conn.(*net.TCPConn).SetKeepAlive(true)
				_ = conn.(*net.TCPConn).SetKeepAlivePeriod(30 * time.Second)
				return tls.Client(conn, dnsConfig), nil
			}

			if proxy == "" {
				return dialer.DialContext(context, network_, _dnsServer)
			}
			//使用代理连接到自定义DNS服务器，代理仅支持TCP
			return Dial("tcp", _dnsServer, outRouterIP)
		},
	}
	return _default_
}
func clean() {
	for {
		time.Sleep(time.Minute)
		dnsLock.Lock()
		for key, value := range dnsTools {
			if time.Now().Sub(value.time) > time.Minute*10 {
				delete(dnsTools, key)
			}
		}
		for key, value := range dnsList {
			if time.Now().Sub(value.time) > time.Minute*10 {
				delete(dnsList, key)
			}
		}
		dnsLock.Unlock()
	}
}

type tools struct {
	rs   *net.Resolver
	time time.Time
}
type rsIps struct {
	ips   []net.IP
	first net.IP
	time  time.Time
}

func SetDnsServer(server string) {
	dnsLock.Lock()
	if server == "local" {
		dnsServer = "localhost"
	} else {
		dnsServer = server
	}
	dnsList = make(map[string]*rsIps)
	dnsTools = make(map[string]*tools)
	dnsLock.Unlock()
}
func IsRemoteDnsServer() bool {
	dnsLock.Lock()
	ok := strings.ToLower(dnsServer) == "remote"
	dnsLock.Unlock()
	return ok
}
func GetDnsServer() string {
	return dnsServer
}
func SetFirstIP(host string, proxyHost string, ip net.IP) {
	key := ""
	if proxyHost == "" {
		key = "_default_" + host
	} else {
		key = proxyHost + "|" + host
	}
	dnsLock.Lock()
	if ip == nil {
		delete(dnsList, key)
	} else {
		ips := dnsList[key]
		if ips != nil {
			ips.first = ip
			ips.time = time.Now()
		}
	}
	dnsLock.Unlock()
}
func GetFirstIP(host string, proxyHost string) net.IP {
	key := ""
	if proxyHost == "" {
		key = "_default_" + host
	} else {
		key = proxyHost + "|" + host
	}
	var ip net.IP
	dnsLock.Lock()
	ips := dnsList[key]
	if ips != nil {
		ip = ips.first
		ips.time = time.Now()
	}
	dnsLock.Unlock()
	return ip
}

// deepCopyIPs 进行深拷贝
func deepCopyIPs(src []net.IP) []net.IP {
	dst := make([]net.IP, len(src)) // 创建与源数组相同大小的目标切片
	for i, ip := range src {
		if ip != nil { // 避免空值
			dst[i] = make(net.IP, len(ip)) // 为每个 IP 重新分配内存
			copy(dst[i], ip)               // 复制数据
		}
	}
	return dst
}
func LookupIP(host string, proxy string, outRouterIP *net.TCPAddr, Dial func(network, address string, outRouterIP *net.TCPAddr) (net.Conn, error)) ([]net.IP, error) {
	dnsLock.Lock()
	localDns, _ := GetLocalEntry(host)
	if localDns != nil {
		dnsLock.Unlock()
		if len(localDns) == 0 {
			return nil, NoLocalDnsEntry
		}
		rand.Seed(time.Now().UnixNano())
		randomIndex := rand.Intn(len(localDns))
		SetFirstIP(host, proxy, localDns[randomIndex])
		return localDns, nil
	}
	if dnsServer == dnsServerLocal {
		dnsLock.Unlock()
		return localLookupIP(host, proxy, outRouterIP)
	}
	dnsLock.Unlock()
	ips, err := lookupIP(host, proxy, outRouterIP, Dial, "ip4")
	if len(ips) > 0 {
		return deepCopyIPs(ips), err
	}
	ips, err = lookupIP(host, proxy, outRouterIP, Dial, "ip")
	if len(ips) > 0 {
		return deepCopyIPs(ips), err
	}
	if proxy == "" {
		return deepCopyIPs(ips), err
	}
	//如果远程没有解析成功,则使用本地DNS解析一次
	return localLookupIP(host, proxy, outRouterIP)
}
func lookupIP(host string, proxy string, outRouterIP *net.TCPAddr, Dial func(network, address string, outRouterIP *net.TCPAddr) (net.Conn, error), Net string) ([]net.IP, error) {
	if proxy == "" {
		return localLookupIP(host, proxy, outRouterIP)
	}
	key := proxy + "|" + host
	dnsLock.Lock()
	ips := dnsList[key]
	if ips != nil {
		ips.time = time.Now()
		dnsLock.Unlock()
		return ips.ips, nil
	}
	resolver := dnsTools[proxy]
	if resolver == nil {
		t := &tools{rs: newResolver(proxy, outRouterIP, Dial)}
		dnsTools[proxy] = t
	}
	resolver = dnsTools[proxy]
	resolver.time = time.Now()
	dnsLock.Unlock()
	_ips_, _err := resolver.rs.LookupIP(context.Background(), Net, host)
	_ips := deepCopyIPs(_ips_)
	if len(_ips) > 0 {
		t := &rsIps{ips: _ips, time: time.Now()}
		dnsLock.Lock()
		dnsList[key] = t
		dnsLock.Unlock()
	}
	return _ips, _err
}
func localLookupIP(host, proxyHost string, outRouterIP *net.TCPAddr) ([]net.IP, error) {
	key := ""
	if proxyHost == "" {
		key = "_default_" + host
	} else {
		key = proxyHost + "|" + host
	}
	dnsLock.Lock()
	ips := dnsList[key]
	if ips != nil {
		ips.time = time.Now()
		dnsLock.Unlock()
		return ips.ips, nil
	}
	dnsLock.Unlock()
	var _ips []net.IP
	var _err error

	DefaultResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var ps SunnyProxy.Proxy
			return ps.DialWithTimeout(network, address, 3*time.Second, outRouterIP)
		},
	}
	_ips, _err = DefaultResolver.LookupIP(context.Background(), "ip", host)

	_ips_ := deepCopyIPs(_ips)
	if len(_ips_) > 0 {
		t := &rsIps{ips: _ips_, time: time.Now()}
		dnsLock.Lock()
		dnsList[key] = t
		dnsLock.Unlock()
	}
	return _ips_, _err
}

// hostEntry 表示一个 hosts 文件中的条目
type hostEntry struct {
	IP        string   // IP 地址
	Hostnames []string // 域名列表
	RawLine   string   // 原始行内容（如果是注释或空行）
}

// ReadAndParseHosts 读取并解析 hosts 文件
func readAndParseHosts() ([]hostEntry, error) {
	filePath := getHostsFilePath()
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	var entries []hostEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 忽略空行或注释
		if line == "" || strings.HasPrefix(line, "#") {
			entries = append(entries, hostEntry{RawLine: line})
			continue
		}

		// 拆分 IP 与域名
		fields := strings.Fields(line)
		if len(fields) < 2 {
			entries = append(entries, hostEntry{RawLine: line})
			continue
		}

		ip := fields[0]
		hostnames := fields[1:]

		entry := hostEntry{
			IP:        ip,
			Hostnames: hostnames,
			RawLine:   line,
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件出错: %v", err)
	}
	return entries, nil
}

// getHostsFilePath 返回操作系统对应的 hosts 文件路径
func getHostsFilePath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return `/etc/hosts`
}

var localDnsHostEntry, _ = readAndParseHosts()
var NoLocalDnsEntry = errors.New("No Local Dns Entry ")

func GetLocalEntry(host string) ([]net.IP, error) {
	// 打印读取到的所有条目
	var ips []net.IP
	for _, entry := range localDnsHostEntry {
		for _, v := range entry.Hostnames {
			if v == host && host != "" {
				ips = append(ips, net.ParseIP(entry.IP))
			}
		}
	}
	if len(ips) > 0 {
		return ips, nil
	}
	return nil, NoLocalDnsEntry

}
