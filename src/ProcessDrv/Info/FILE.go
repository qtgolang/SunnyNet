package Info

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// MoveFileToTempDir 将指定文件移动到 Windows 临时目录
// srcFile: 源文件路径，destFileName：目标文件名
// 返回值：目标文件的路径，以及可能出现的错误
func MoveFileToTempDir(srcFile, destFileName string) string {
	tempDir := os.TempDir()
	// 拼接目标文件路径
	destPath := filepath.Join(tempDir, destFileName)
	// 移动文件
	err := os.Rename(srcFile, destPath)
	_ = os.Remove(destPath)
	if err != nil {
		return ""
	}
	return destPath
}

// 生成指定长度的随机字母串
func RandomLetters(length int) string {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())
	// 生成指定长度的随机字母
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	result := make([]rune, length)

	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}
