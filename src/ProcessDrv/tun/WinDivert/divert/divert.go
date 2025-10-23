//go:build windows

package divert

import "C"

var i int

func Init() int {
	i++
	return i
}
