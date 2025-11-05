//go:build toolsxxxxxxxxx
// +build toolsxxxxxxxxx
// 这个文件带了 //go:build toolsxxxxxxxxx，所以 正常构建不会把这个 import 编进去
// 只有在你显式指定 -tags=toolsxxxxxxxxx 时，这个 import 才生效，divert 包才会被认为是依赖。
// 用 toolsxxxxxxxxx 构建标签“骗”一下 go mod vendor 才会拉取到divert目录

package WinDivert

import _"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/WinDivert/divert"