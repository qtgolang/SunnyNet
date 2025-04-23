package Api

import "C"
import (
	"bytes"
	"golang.org/x/image/webp"
	"image/jpeg"
	"image/png"
	"os"
)

// WebpToPng Webp图片转Png图片 根据文件名
func WebpToPng(webpName, save string) bool {
	f0, err := os.Open(webpName)
	if err != nil {
		return false
	}
	defer func() {
		_ = f0.Close()
	}()
	img0, err := webp.Decode(f0)
	if err != nil {
		return false
	}
	pngFile, err := os.Create(save)
	if err != nil {
		return false
	}
	defer func() {
		_ = pngFile.Close()
	}()
	err = (&png.Encoder{CompressionLevel: png.NoCompression}).Encode(pngFile, img0)
	if err != nil {
		return false
	}
	return true
}

// WebpToJpeg 图片转JEG图片 根据文件名 SaveQuality=质量(默认75)
func WebpToJpeg(webpName, save string, SaveQuality int) bool {
	f0, err := os.Open(webpName)
	if err != nil {
		return false
	}
	defer func() {
		_ = f0.Close()
	}()
	img0, err := webp.Decode(f0)
	if err != nil {
		return false
	}
	pngFile, err := os.Create(save)
	if err != nil {
		return false
	}
	defer func() {
		_ = pngFile.Close()
	}()
	_SaveQuality := SaveQuality
	if _SaveQuality < 1 {
		SaveQuality = 75
	}
	err = jpeg.Encode(pngFile, img0, &jpeg.Options{Quality: _SaveQuality})
	if err != nil {
		return false
	}
	return true
}

// WebpToPngBytes Webp图片转Png图片字节数组
func WebpToPngBytes(_webp []byte) []byte {
	var b bytes.Buffer
	b.Write(_webp)
	defer func() {
		b.Reset()
	}()
	img0, err := webp.Decode(&b)
	if err != nil {
		return nil
	}
	var bs bytes.Buffer
	defer func() {
		bs.Reset()
	}()
	err = (&png.Encoder{CompressionLevel: png.NoCompression}).Encode(&bs, img0)
	if bs.Len() < 1 || err != nil {
		return nil
	}
	return bs.Bytes()
}

// WebpToJpegBytes Webp图片转JEG图片字节数组 SaveQuality=质量(默认75)
func WebpToJpegBytes(_webp []byte, SaveQuality int) []byte {
	var b bytes.Buffer
	b.Write(_webp)
	defer func() {
		b.Reset()
	}()
	img0, err := webp.Decode(&b)
	if err != nil {
		return nil
	}
	var bs bytes.Buffer
	defer func() {
		bs.Reset()
	}()
	_SaveQuality := SaveQuality
	if _SaveQuality < 1 {
		SaveQuality = 75
	}
	err = jpeg.Encode(&bs, img0, &jpeg.Options{Quality: _SaveQuality})
	if bs.Len() < 1 || err != nil {
		return nil
	}
	return bs.Bytes()
}
