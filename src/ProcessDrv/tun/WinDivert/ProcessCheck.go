//go:build windows
// +build windows

package WinDivert

func (d *Divert) pidFromCheck(pid int32, name string) (ok bool) {
	if _myPid == pid {
		return true
	}
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if d.checkProcess == nil {
		return false
	}
	return d.checkProcess(pid, name)
}
