package VirtualFile

import (
	"bytes"
	"golang.org/x/sys/windows"
	"io"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	FILE_MAP_ALL_ACCESS = 0xF001F
	FILE_MAP_READ       = 0x0004 // 只读访问权限
)

type VirtualFile struct {
	f    windows.Handle
	view uintptr
	size int
}

func (v *VirtualFile) Write(p []byte) (n int, err error) {
	if v == nil || v.view == 0 {
		return 0, io.EOF // 如果虚拟文件无效，返回 EOF
	}
	var buf bytes.Buffer
	buf.Write(p)
	buf.WriteByte(0)
	bs := buf.Bytes()
	// 计算要写入的字节数
	bytesToWrite := buf.Len()
	if bytesToWrite+8 > v.size {
		bytesToWrite = v.size - 8
	}
	// 写入数据
	dst := (*[1 << 30]byte)(unsafe.Pointer(v.view))[8 : 8+bytesToWrite] // 创建指向视图的字节切片
	copy(dst, bs[:bytesToWrite])
	return bytesToWrite, nil // 返回写入的字节数
}

func (v *VirtualFile) Close() {
	if v == nil {
		return
	}
	if v.f != 0 {
		_ = windows.CloseHandle(v.f) // 关闭文件映射句柄
	}
	if v.view != 0 {
		_ = windows.UnmapViewOfFile(v.view) // 解除映射视图
	}
}

func Create(path string, size int) (io.Writer, error) {
	var Virtual VirtualFile
	Virtual.size = size
	mappingName := `Global\` + path
	fileMapping, err := windows.CreateFileMapping(windows.InvalidHandle, nil, windows.PAGE_READWRITE, 0, uint32(size), windows.StringToUTF16Ptr(mappingName))
	if err != nil {
		return nil, err
	}
	Virtual.f = fileMapping
	view, err := windows.MapViewOfFile(fileMapping, FILE_MAP_ALL_ACCESS, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	Virtual.view = view
	hexStr := strconv.FormatInt(int64(size), 16)
	p := 8 - len(hexStr)
	if p > 0 {
		for i := 0; i < p; i++ {
			hexStr = "0" + hexStr
		}
	}
	Virtual.Write([]byte(hexStr))
	return &Virtual, nil
}

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	openFileMapping = kernel32.NewProc("OpenFileMappingW")
)

/*
Read 读取虚拟文件
*/
func Read(path string, isChar bool) ([]byte, error) {
	// 打开命名的内存映射文件
	mappingName := syscall.StringToUTF16Ptr(`Global\` + path)
	hMapFile, _, err := openFileMapping.Call(FILE_MAP_READ, 0, uintptr(unsafe.Pointer(mappingName)))
	if hMapFile == 0 {
		return nil, err
	}
	defer syscall.CloseHandle(syscall.Handle(hMapFile))

	// 映射视图
	view, err := syscall.MapViewOfFile(syscall.Handle(hMapFile), FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.UnmapViewOfFile(view)
	if isChar {
		var buff bytes.Buffer
		src := (*[1 << 30]byte)(unsafe.Pointer(view))[:] // 创建指向视图的字节切片
		i := 8
		for {
			b := src[i]
			if b != 0 {
				buff.WriteByte(b)
				i++
			} else {
				break
			}
		}
		return buff.Bytes(), nil
	}
	src := (*[1 << 30]byte)(unsafe.Pointer(view))[:]
	pLen := make([]byte, 8)
	copy(pLen, src)
	NUM, _ := strconv.ParseInt(string(pLen), 16, 32)
	if NUM > 0 {
		NUM -= 8
		p := make([]byte, NUM)
		copy(p, src[8:])
		return p, nil
	}
	return nil, nil
}
