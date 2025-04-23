//go:build !windows
// +build !windows

package Resource

func SetAdminRun(path string) error {
	return nil
}
