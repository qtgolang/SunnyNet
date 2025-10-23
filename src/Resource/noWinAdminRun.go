//go:build !windows
// +build !windows

package Resource

func SetAdminRun(path string) error {
	return nil
}

var WinDivert64 []byte
 
var WinDivert32 []byte
