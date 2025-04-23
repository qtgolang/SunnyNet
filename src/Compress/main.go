package Compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"io"
	"io/ioutil"
)

var _null_bytes = make([]byte, 0)

// DeflateCompress Deflate压缩 (可能等同于zlib压缩)
func DeflateCompress(data []byte) []byte {
	var o bytes.Buffer
	f, _ := flate.NewWriter(&o, flate.BestCompression)
	if a, b := f.Write(data); a == 0 || b != nil {
		return _null_bytes
	}
	if f.Flush() != nil {
		return _null_bytes
	}
	return o.Bytes()
}

// DeflateUnCompress Deflate解压缩 (可能等同于zlib解压缩)
func DeflateUnCompress(data []byte) []byte {
	zr := flate.NewReader(ioutil.NopCloser(bytes.NewBuffer(data)))
	bx, _ := io.ReadAll(zr)
	_ = zr.Close()
	return bx
}

// ZlibUnCompress zlib解压缩
func ZlibUnCompress(data []byte) []byte {
	b := bytes.NewReader(data)
	var out bytes.Buffer
	r, e := zlib.NewReader(b)
	if e != nil {
		return _null_bytes
	}
	_, _ = io.Copy(&out, r)
	_ = r.Close()
	return out.Bytes()
}

// ZlibCompress zlib压缩
func ZlibCompress(data []byte) []byte {
	var buf bytes.Buffer
	compressor, err := zlib.NewWriterLevel(&buf, zlib.DefaultCompression)
	if err != nil {
		return _null_bytes
	}
	_, _ = compressor.Write(data)
	_ = compressor.Close()
	return buf.Bytes()
}

// GzipCompress Gzip压缩
func GzipCompress(data []byte) []byte {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	_, _ = writer.Write(data)
	_ = writer.Close()
	return buffer.Bytes()
}

// BrUnCompress br解压缩
func BrUnCompress(data []byte) []byte {
	r := ioutil.NopCloser(bytes.NewBuffer(data))
	b, _ := io.ReadAll(brotli.NewReader(r))
	_ = r.Close()
	return b
}

// BrCompress br压缩
func BrCompress(data []byte) []byte {
	var compressed bytes.Buffer
	writer := brotli.NewWriter(&compressed)
	_, _ = writer.Write(data)
	_ = writer.Close()
	return compressed.Bytes()
}

// GzipUnCompress Gzip解压缩
func GzipUnCompress(data []byte) []byte {
	r := ioutil.NopCloser(bytes.NewBuffer(data))
	gr, err := gzip.NewReader(r)
	if err != nil {
		_ = r.Close()
		return _null_bytes
	}
	b, _ := io.ReadAll(gr)
	_ = r.Close()
	return b
}

func ZSTDCompress(input []byte) []byte {
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(4))
	if err != nil {
		return _null_bytes
	}
	return encoder.EncodeAll(input, make([]byte, 0, len(input)))
}

func ZSTDDecompress(input []byte) []byte {
	var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	a, _ := decoder.DecodeAll(input, nil)
	return a
}
