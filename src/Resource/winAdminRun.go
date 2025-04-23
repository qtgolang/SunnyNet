//go:build windows
// +build windows

package Resource

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"strings"
)

func setCompatibilitySettings(programPath string, compatibilityVersion int, colorMode *int, run640x480 bool, disableFullscreen bool, runAsAdmin bool, useOneDrive bool, dpiOverride *int, programDPI *int) error {
	var compatibility string
	var colorSetting string
	var resolutionSetting string
	var fullscreenSetting string
	var adminSetting string
	var dpiSetting string
	var dpiProgramSetting string
	var oneDriveSetting string

	// 选择兼容版本
	switch compatibilityVersion {
	case 1:
		compatibility = "WIN95 "
	case 2:
		compatibility = "WIN98 "
	case 3:
		compatibility = "NT4SP5 "
	case 4:
		compatibility = "WIN2000 "
	case 5:
		compatibility = "WINXPSP2 "
	case 6:
		compatibility = "WINXPSP3 "
	case 7:
		compatibility = "VISTARTM "
	case 8:
		compatibility = "VISTASP1 "
	case 9:
		compatibility = "VISTASP2 "
	case 10:
		compatibility = "WIN7RTM "
	case 11:
		compatibility = "WIN8RTM "
	default:
		compatibility = ""
	}

	// 简化的颜色模式
	if colorMode != nil {
		switch *colorMode {
		case 1:
			colorSetting = "256COLOR "
		case 2:
			colorSetting = "16BITCOLOR "
		default:
			colorSetting = ""
		}
	}

	// 640x480分辨率设置
	if run640x480 {
		resolutionSetting = "640X480 "
	}

	// 禁用全屏优化
	if disableFullscreen {
		fullscreenSetting = "DISABLEDXMAXIMIZEDWINDOWEDMODE "
	}

	// 管理员运行
	if runAsAdmin {
		adminSetting = "RUNASADMIN "
	}

	// 代替DPI缩放
	if dpiOverride != nil {
		switch *dpiOverride {
		case 1:
			dpiSetting = "HIGHDPIAWARE "
		case 2:
			dpiSetting = "DPIUNAWARE "
		case 3:
			dpiSetting = "GDIDPISCALING DPIUNAWARE "
		default:
			dpiSetting = ""
		}
	}

	// 程序DPI设置
	if programDPI != nil {
		switch *programDPI {
		case 1:
			dpiProgramSetting = "PERPROCESSSYSTEMDPIFORCEOFF "
		case 2:
			dpiProgramSetting = "PERPROCESSSYSTEMDPIFORCEON "
		default:
			dpiProgramSetting = ""
		}
	}

	// 使用 OneDrive 文件
	if useOneDrive {
		oneDriveSetting = "PLACEHOLDERFILES "
	}

	// 构建完整的注册表值
	compatibilitySettings := compatibility + colorSetting + resolutionSetting + fullscreenSetting + adminSetting + dpiSetting + dpiProgramSetting + oneDriveSetting
	compatibilitySettings = strings.TrimSpace(compatibilitySettings)

	// 写入注册表
	return writeRegistryValue(programPath, compatibilitySettings)
}

func writeRegistryValue(programPath, value string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %v", err)
	}
	defer key.Close()

	err = key.SetStringValue(programPath, value)
	if err != nil {
		return fmt.Errorf("failed to write registry value: %v", err)
	}

	return nil
}
func SetAdminRun(path string) error {
	return setCompatibilitySettings(path, 10, nil, false, false, true, false, nil, nil)

}
