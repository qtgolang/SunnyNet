//go:build windows
// +build windows

package WinDivert

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

var (
	errQueueLength = fmt.Errorf("Queue length is not correct, Max: %v, Min: %v", QueueLengthMax, QueueLengthMin)
	errQueueTime   = fmt.Errorf("Queue time is not correct, Max: %v, Min: %v", QueueTimeMax, QueueTimeMin)
	errQueueSize   = fmt.Errorf("Queue size is not correct, Max: %v, Min: %v", QueueSizeMax, QueueSizeMin)
	errQueueParam  = errors.New("VersionMajor and VersionMinor only can be used in function GetParam")
	errPriority    = fmt.Errorf("Priority is not Correct, Max: %v, Min: %v", PriorityHighest, PriorityLowest)
)

const (
	// The driver files WinDivert32.sys or WinDivert64.sys were not found
	ErrFileNotFound = Error(windows.ERROR_FILE_NOT_FOUND)

	// The calling application does not have Administrator privileges
	ErrAccessDenied = Error(windows.ERROR_ACCESS_DENIED)

	// This indicates an invalid packet filter string, layer, priority, or flags
	ErrInvalidParameter = Error(windows.ERROR_INVALID_PARAMETER)

	// The WinDivert32.sys or WinDivert64.sys driver does not have a valid digital signature (see the driver signing requirements above)
	ErrInvalidImageHash = Error(windows.ERROR_INVALID_IMAGE_HASH)

	// An incompatible version of the WinDivert driver is currently loaded
	ErrDriverFailedPriorUnload = Error(windows.ERROR_DRIVER_FAILED_PRIOR_UNLOAD)

	// The handle was opened with the WINDIVERT_FLAG_NO_INSTALL flag and the WinDivert driver is not already installed
	ErrServiceDoseNotExist = Error(windows.ERROR_SERVICE_DOES_NOT_EXIST)

	// This error occurs for various reasons, including: the WinDivert driver is blocked by security software; or you are using a virtualization environment that does not support drivers
	ErrDriverBlocked = Error(windows.ERROR_DRIVER_BLOCKED)

	// The captured packet is larger than the pPacket buffer
	ErrInsufficientBuffer = Error(windows.ERROR_INSUFFICIENT_BUFFER)

	// The handle has been shutdown using WinDivertShutdown() and the packet queue is empty
	ErrNoData = Error(windows.ERROR_NO_DATA)

	// The error code ERROR_IO_PENDING indicates that the overlapped operation has been successfully initiated and that completion will be indicated at a later time
	ErrIOPending = Error(windows.ERROR_IO_PENDING)

	// This error occurs when an impostor packet (with pAddr->Impostor set to 1) is injected and the ip.TTL or ipv6.HopLimit field goes to zero. This is a defense of "last resort" against infinite loops caused by impostor packets
	ErrHostUnreachable = Error(windows.ERROR_HOST_UNREACHABLE)

	// This error occurs when the Base Filtering Engine service has been disabled
	ErrNotRegistered = Error(windows.EPT_S_NOT_REGISTERED)

	// The I/O operation has been aborted because of either a thread exit or an application request
	ErrOperationAborted = Error(windows.ERROR_OPERATION_ABORTED)

	// The handle is invalid
	ErrInvalidHandle = Error(windows.ERROR_INVALID_HANDLE)
)

// Error is ...
type Error windows.Errno

// Error is ...
func (e Error) Error() string {
	switch windows.Errno(e) {
	case windows.ERROR_FILE_NOT_FOUND:
		return "The driver files WinDivert32.sys or WinDivert64.sys were not found"
	case windows.ERROR_ACCESS_DENIED:
		return "The calling application does not have Administrator privileges"
	case windows.ERROR_INVALID_PARAMETER:
		return "This indicates an invalid packet filter string, layer, priority, or flags"
	case windows.ERROR_INVALID_IMAGE_HASH:
		return "The WinDivert32.sys or WinDivert64.sys driver does not have a valid digital signature (see the driver signing requirements above)"
	case windows.ERROR_DRIVER_FAILED_PRIOR_UNLOAD:
		return "An incompatible version of the WinDivert driver is currently loaded"
	case windows.ERROR_SERVICE_DOES_NOT_EXIST:
		return "The handle was opened with the WINDIVERT_FLAG_NO_INSTALL flag and the WinDivert driver is not already installed"
	case windows.ERROR_DRIVER_BLOCKED:
		return "This error occurs for various reasons, including: the WinDivert driver is blocked by security software; or you are using a virtualization environment that does not support drivers"
	case windows.ERROR_INSUFFICIENT_BUFFER:
		return "The captured packet is larger than the pPacket buffer"
	case windows.ERROR_NO_DATA:
		return "The handle has been shutdown using WinDivertShutdown() and the packet queue is empty"
	case windows.ERROR_IO_PENDING:
		return "The error code ERROR_IO_PENDING indicates that the overlapped operation has been successfully initiated and that completion will be indicated at a later time"
	case windows.ERROR_HOST_UNREACHABLE:
		return "This error occurs when an impostor packet (with pAddr->Impostor set to 1) is injected and the ip.TTL or ipv6.HopLimit field goes to zero. This is a defense of \"last resort\" against infinite loops caused by impostor packets"
	case windows.EPT_S_NOT_REGISTERED:
		return "This error occurs when the Base Filtering Engine service has been disabled"
	case windows.ERROR_OPERATION_ABORTED:
		return "The I/O operation has been aborted because of either a thread exit or an application request"
	case windows.ERROR_INVALID_HANDLE:
		return "The handle is invalid"
	default:
		return windows.Errno(e).Error()
	}
}
