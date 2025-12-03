// Package public /*
/*

									 Package public
------------------------------------------------------------------------------------------------
                                   程序所用到的所有公共方法
------------------------------------------------------------------------------------------------
*/
package public

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"hash/fnv"
	"io/ioutil"
	"math"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/qtgolang/SunnyNet/src/ReadWriteObject"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/http"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var Timeout = "timeout"
var MaxBig = new(big.Int).Lsh(big.NewInt(1), 128)
var Theology = int64(0)               //中间件唯一ID
var MaxUploadLength = int64(10240000) //<-10M  3.95M->4096000 //POST数据最大数据长度,超过这个长度请求的Body将无法查看
var MaxUploadMsg = http.MaxUploadMsg
var ProvideForwardingServiceOnly = http.ProvideForwardingServiceOnly

// RemoveFile 删除文件
func RemoveFile(Filename string) error {
	return os.Remove(Filename)
}

// CheckFileIsExist 检查文件是否存在
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// WriteBytesToFile 写入数据到文件
func WriteBytesToFile(bytes []byte, Filename string) error {
	var f *os.File
	var err error
	//文件是否存在
	if CheckFileIsExist(Filename) {
		//存在 删除
		err = RemoveFile(Filename)
		if err != nil {
			return err
		}
	}
	//创建文件
	f, err = os.Create(Filename)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// 写入
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

// GetMethod 从指定数据中获取HTTP请求 Method
func GetMethod(s []byte) string {
	if len(s) < 11 {
		return NULL
	}
	method := strings.ToUpper(SubString("+"+CopyString(string(s[0:11])), "+", Space))
	if !IsHttpMethod(method) {
		return NULL
	}
	return method
}

// IsHttpMethod 通过指定字符串判断是否HTTP数据
func IsHttpMethod(methods string) bool {
	method := strings.ToUpper(methods)
	if method == HttpMethodGET {
		return true
	}
	if method == HttpMethodPUT {
		return true
	}
	if method == HttpMethodPOST {
		return true
	}
	if method == HttpMethodDELETE {
		return true
	}
	if method == HttpMethodHEAD {
		return true
	}
	if method == HttpMethodOPTIONS {
		return true
	}
	if method == HttpMethodCONNECT {
		return true
	}
	if method == HttpMethodTRACE {
		return true
	}
	if method == HttpMethodPATCH {
		return true
	}
	return false
}

// LocalBuildBody 本地文件响应数据转为Bytes
func LocalBuildBody(ContentType string, Body interface{}) []byte {
	var buffer bytes.Buffer
	var b []byte
	switch v := Body.(type) {
	case []byte:
		b = v
		break
	case string:
		b = []byte(v)
		break
	default:
		break
	}
	l := strconv.Itoa(len(b))
	buffer.WriteString("HTTP/1.1 200 OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: " + l + "\r\nContent-Type: " + ContentType + CRLF + CRLF)
	buffer.Write(b)
	return CopyBytes(buffer.Bytes())
}

// SplitHostPort 分割Host 和 Port 忽略IPV6地址
func SplitHostPort(ip string) (host, port string, err error) {
	arr := strings.Split(ip, ":")
	if len(arr) < 3 {
		return net.SplitHostPort(ip)
	}
	return net.SplitHostPort(ip)
}

// IsIPv6 是否IPV6
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip.To4() == nil
}

// IsIPv4 是否IPV4
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip.To4() != nil
}

// ReadWriterPeek 在读写对象中读取n个字节,而不推进读取器
func ReadWriterPeek(f *ReadWriteObject.ReadWriteObject, n int) string {
	r, _ := f.Peek(n)
	return strings.ToUpper(string(r))
}

// IsHTTPRequest 在读写对象中判断是否为HTTP请求
func IsHTTPRequest(i byte, f *ReadWriteObject.ReadWriteObject) bool {
	switch i {
	case 'C':
		r := ReadWriterPeek(f, 8)
		return r == "CONNECT "
	case 'O':
		r := ReadWriterPeek(f, 8)
		return r == "OPTIONS "
	case 'H':
		r := ReadWriterPeek(f, 5)
		return r == "HEAD "
	case 'D':
		r := ReadWriterPeek(f, 7)
		return r == "DELETE "
	case 'G':
		r := ReadWriterPeek(f, 4)
		return r == "GET "
	case 'P':
		r := ReadWriterPeek(f, 5)
		if r == "POST " {
			return true
		}
		r = ReadWriterPeek(f, 4)
		if r == "PUT " {
			return true
		}
		r = ReadWriterPeek(f, 6)
		return r == "PATCH "
	case 'T':
		r := ReadWriterPeek(f, 6)
		if r == "TRACE " {
			return true
		}

	}
	return false
}

// SubString 截取字符串中间部分
func SubString(str, left, Right string) string {
	s := strings.Index(str, left)
	if s < 0 {
		return NULL
	}
	s += len(left)
	e := strings.Index(str[s:], Right)
	if e+s <= s {
		return NULL
	}
	bs := make([]byte, e)
	copy(bs, str[s:s+e])
	return string(bs)
}

// StructureBody HTTP响应体转为字节数组
func StructureBody(heads *http.Response) []byte {
	var buffer bytes.Buffer
	status := heads.StatusCode
	if status == 0 {
		status = 200
	}
	buffer.Write([]byte("HTTP/1.1 " + strconv.Itoa(status) + " " + http.StatusText(status) + CRLF))
	if heads != nil {
		if heads.Header != nil {
			for name, values := range heads.Header {
				for _, value := range values {
					buffer.Write([]byte(name + ": " + value + CRLF))
				}
			}
		}
	}

	buffer.Write([]byte(CRLF))
	if heads != nil {
		if heads.Body != nil {
			bodyBytes, _ := ioutil.ReadAll(heads.Body)
			heads.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			buffer.Write(bodyBytes)
		}
	}
	return CopyBytes(buffer.Bytes())
}

// CStringToBytes C字符串转字节数组
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

// BytesToCString C字符串转字节数组
func BytesToCString(r uintptr) string {
	data := make([]byte, 0)
	if r == 0 {
		return ""
	}
	i := 0
	for {
		p := *(*byte)(unsafe.Pointer(r + uintptr(i)))
		if p == 0 {
			break
		}
		data = append(data, p)
		i++
	}
	return string(data)
}

// GetCertificateName 提取证书名称
func GetCertificateName(certBytes []byte) string {
	// 尝试解析 PEM 格式的证书
	block, _ := pem.Decode(certBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return ""
	}
	// 解析 DER 格式的证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ""
	}
	// 返回证书的主题名称
	return cert.Subject.CommonName
}

// WriteErr 将错误信息写入指针 请确保指针内的空间足够
func WriteErr(err error, Ptr uintptr) {
	if err == nil || Ptr == 0 {
		return
	}
	bin := []byte(err.Error())
	for i := 0; i < len(bin); i++ {
		*(*byte)(unsafe.Pointer(Ptr + uintptr(i))) = bin[i]
	}
	*(*byte)(unsafe.Pointer(Ptr + uintptr(len(bin)))) = 0
}

// IntToBytes int转字节数组
func IntToBytes(n int) []byte {
	data := int64(n)
	byteBuf := bytes.NewBuffer([]byte{})
	_ = binary.Write(byteBuf, binary.BigEndian, data)
	return byteBuf.Bytes()
}

// Int64ToBytes int64转字节数组
func Int64ToBytes(data int64) []byte {
	byteBuf := bytes.NewBuffer([]byte{})
	_ = binary.Write(byteBuf, binary.BigEndian, &data)
	ss := byteBuf.Bytes()
	return ss
}

func Float64ToBytes(float float64) []byte {
	bits := math.Float64bits(float)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, bits)
	return b
}

// BytesCombine 连接多个字节数组
func BytesCombine(pBytes ...[]byte) []byte {
	Len := len(pBytes)
	s := make([][]byte, Len)
	for index := 0; index < Len; index++ {
		s[index] = pBytes[index]
	}
	sep := []byte(NULL)
	return bytes.Join(s, sep)
}

// ContentTypeIsText 是否是文档类型
func ContentTypeIsText(ContentType string) bool {
	contentType := strings.ToLower(ContentType)
	if strings.Index(contentType, "application/json") != -1 {
		return true
	}
	if strings.Index(contentType, "application/javascript") != -1 {
		return true
	}
	if strings.Index(contentType, "application/x-javascript") != -1 {
		return true
	}
	if strings.Index(contentType, "text/") != -1 {
		return true
	}
	return false

}

// IsForward 是否是大文件类型
func IsForward(ContentType string) bool {
	contentType := strings.ToLower(ContentType)
	if strings.Index(contentType, "video/") != -1 {
		return true
	}
	if strings.Index(contentType, "audio/") != -1 {
		return true
	}
	return false
}

var HostArray sync.Mutex

// IsCerRequest 是否下载证书的请求
func IsCerRequest(request *http.Request, port int) bool {
	portStr := strconv.Itoa(port)
	if request == nil {
		return false
	}
	HostArray.Lock()
	defer HostArray.Unlock()
	if request.URL != nil {
		_host := request.URL.Host
		if _host == "127.0.0.1:"+portStr {
			return true
		}
		if _host == "localhost:"+portStr {
			return true
		}
	}
	return false
}

// LegitimateRequest 解析HTTP 请求体，                 2024-01-20修改后的代码
func LegitimateRequest(s []byte) (bool, bool, int, int, bool) {
	a := strings.ToLower(CopyString(string(s)))
	arrays := strings.Split(a, Space)
	isHttpRequest := false
	Method := GetMethod(s)
	if IsHttpMethod(Method) && Method != HttpMethodCONNECT {
		if SubString(string(s), Space, Space) != "" {
			isHttpRequest = true
		}
	}
	if len(arrays) > 1 {
		//Body中是否有长度
		islet := strings.Index(a, "content-length: ") != -1
		if islet {
			ContentLength, _ := strconv.Atoi(SubString(a+CRLF, "content-length: ", CRLF))
			if ContentLength == 0 {
				if strings.Contains(a, CRLF+CRLF) {
					// 有长度  但长度为0 直接验证成功,并且有CRLF+CRLF
					return islet, true, 0, ContentLength, isHttpRequest
				}
				// 有长度  但长度为0 但是没有 有CRLF+CRLF 直接验证失败
				return islet, false, 0, ContentLength, isHttpRequest
			}
			arr := bytes.Split(s, []byte(CRLF+CRLF))
			if len(arr) < 2 {
				// 读取验证失败
				return islet, false, 0, ContentLength, isHttpRequest
			}
			var b bytes.Buffer
			for i := 0; i < len(arr); i++ {
				if i != 0 {
					b.Write(CopyBytes(arr[i]))
					b.Write([]byte{13, 10, 13, 10})
				}
			}
			if b.Len() == ContentLength || b.Len()-4 == ContentLength {
				b.Reset()
				// 有长度  读取验证成功
				return islet, true, 0, ContentLength, isHttpRequest
			}
			v := b.Len() - 4
			b.Reset()
			return islet, false, v, ContentLength, isHttpRequest
		} else if strings.Index(a, "transfer-encoding: chunked") != -1 {
			islet = true
			arr := bytes.Split(s, []byte(CRLF+CRLF))
			if len(arr) < 2 {
				// 读取验证失败
				return islet, false, 0, 0, isHttpRequest
			}
			BodyLen := 0
			var bs bytes.Buffer
			for i := 0; i < len(arr); i++ {
				if i != 0 {
					if len(arr[i]) == 0 {
						continue
					}
					bs.Write(CopyBytes(arr[i]))
					bs.Write([]byte{13, 10, 13, 10})
					BodyLen += len(arr[i]) + 4
				}
			}
			reader := bufio.NewReader(bytes.NewReader(bs.Bytes()))
			bLen := int64(0)
			for {
				T, _, e := reader.ReadLine()
				ContentLength2, _ := strconv.ParseInt(string(T), 16, 64)
				if ContentLength2 == 0 && e != nil {
					break
				}
				bLen += ContentLength2 + int64(len(T)+2)
				if ContentLength2 == 0 {
					bLen += 2
					break
				}
				p, _ := reader.Discard(int(ContentLength2))
				if p != int(ContentLength2) {
					v := bLen - ContentLength2 + int64(p)
					return islet, false, int(v), 0, isHttpRequest
				}
				bLen += 2
				p, _ = reader.Discard(2)
				if p != 2 {
					v := bLen - 2 + int64(p)
					return islet, false, int(v), 0, isHttpRequest
				}
			}
			if int64(BodyLen) == bLen {
				return islet, true, BodyLen, BodyLen, isHttpRequest
			}
			return islet, false, int(bLen), int(bLen), isHttpRequest
		}
		if (Method == HttpMethodGET || Method == HttpMethodOPTIONS || Method == HttpMethodHEAD) && len(s) > 4 && CopyString(string(s[len(s)-4:])) == CRLF+CRLF {
			return false, true, 0, 0, isHttpRequest
		}
		//没有长度  读取验证失败
		return false, false, 0, 0, isHttpRequest
	}
	return false, false, 0, 0, isHttpRequest

}

func Utf8ToGbk(input string) (string, error) {
	// 创建 GBK 编码的转换器
	encoder := simplifiedchinese.GBK.NewEncoder()
	// 转换
	result, _, err := transform.String(encoder, input)
	if err != nil {
		return "", err
	}
	return result, nil
}

func GbkToUtf8(input string) (string, error) {
	// 创建 GBK 解码器
	decoder := simplifiedchinese.GBK.NewDecoder()
	// 转换
	result, err := decoder.String(input)
	if err != nil {
		return "", err
	}
	return result, nil
}

// CopyBytes 拷贝 字节数组避免内存泄漏
var CopyBytes = http.CopyBytes

// CopyString 拷贝字符串 避免内存泄漏
func CopyString(src string) string {
	dst := make([]byte, len(src))
	copy(dst, src)
	return string(dst)
}

func HTTPBanRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}
func HTTPAllowRedirect(*http.Request, []*http.Request) error {
	return nil
}
func IsLocalIP(ipStr string) (bool, net.IP) {
	if ipStr == "" {
		return false, nil
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, ip
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return false, ip
	}

	for _, iface := range interfaces {
		addrs, err1 := iface.Addrs()
		if err1 != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.Contains(ip) {
					return true, ip
				}
			}
		}
	}

	return false, nil
}
func SumHashCode(s string) uint32 {
	var hash int32 = 0
	for _, ch := range s {
		hash = 31*hash + ch
	}
	return uint32(hash)
}

var RouterIPInspect = SunnyProxy.RouterIPInspect

func FNV32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
