package Api

import "C"
import (
	"github.com/qtgolang/SunnyNet/src/Compress"
	"github.com/qtgolang/SunnyNet/src/public"
)

// DeflateCompress Deflate压缩 (可能等同于zlib压缩)
func DeflateCompress(bin []byte) []byte {
	if len(bin) < 1 {
		return nil
	}
	bx := Compress.DeflateCompress(bin)
	if len(bx) < 1 {
		return nil
	}
	return bx
}

// DeflateUnCompress Deflate解压缩 (可能等同于zlib解压缩)
func DeflateUnCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	bx := Compress.DeflateUnCompress(bin)
	if len(bx) < 1 {
		return 0
	}
	bx = public.BytesCombine(public.IntToBytes(len(bx)), bx)
	return public.PointerPtr(string(bx))
}

// ZlibUnCompress zlib解压缩
func ZlibUnCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	bx := Compress.ZlibUnCompress(bin)
	if len(bx) < 1 {
		return 0
	}
	bx = public.BytesCombine(public.IntToBytes(len(bx)), bx)
	return public.PointerPtr(string(bx))
}

// ZlibCompress zlib压缩
func ZlibCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	out := Compress.ZlibCompress(bin)
	if len(out) < 1 {
		return 0
	}
	out = public.BytesCombine(public.IntToBytes(len(out)), out)
	return public.PointerPtr(string(out))
}

// GzipCompress Gzip压缩
func GzipCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	out := Compress.GzipCompress(bin)
	if len(out) < 1 {
		return 0
	}
	out = public.BytesCombine(public.IntToBytes(len(out)), out)
	return public.PointerPtr(string(out))
}

// BrUnCompress br解压缩
func BrUnCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	b := Compress.BrUnCompress(bin)
	if len(b) < 1 {
		return 0
	}
	b = public.BytesCombine(public.IntToBytes(len(b)), b)
	return public.PointerPtr(string(b))
}

// BrCompress br压缩
func BrCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	compressedData := Compress.BrCompress(bin)
	if len(compressedData) < 1 {
		return 0
	}
	compressedData = public.BytesCombine(public.IntToBytes(len(compressedData)), compressedData)
	return public.PointerPtr(string(compressedData))
}

// GzipUnCompress Gzip解压缩
func GzipUnCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}

	b := Compress.GzipUnCompress(bin)
	if len(b) < 1 {
		return 0
	}
	b = public.BytesCombine(public.IntToBytes(len(b)), b)
	return public.PointerPtr(b)
}

// ZSTDCompress ZSTD压缩
func ZSTDCompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	compressedData := Compress.ZSTDCompress(bin)
	if len(compressedData) < 1 {
		return 0
	}
	compressedData = public.BytesCombine(public.IntToBytes(len(compressedData)), compressedData)
	return public.PointerPtr(string(compressedData))
}

// ZSTDDecompress ZSTD 解压缩
func ZSTDDecompress(data uintptr, dataLen int) uintptr {
	bin := public.CStringToBytes(data, dataLen)
	if len(bin) < 1 {
		return 0
	}
	b := Compress.ZSTDDecompress(bin)
	if len(b) < 1 {
		return 0
	}
	b = public.BytesCombine(public.IntToBytes(len(b)), b)
	return public.PointerPtr(b)
}
