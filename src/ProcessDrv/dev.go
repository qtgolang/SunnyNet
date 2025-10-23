package ProcessDrv

type Dev interface {
	//Install 安装依赖驱动
	Install() bool
	//IsRun 是否已运行
	IsRun() bool
	//SetHandle 设置回调
	SetHandle() bool
	//Run 开始运行
	Run() bool
	//Close 关闭
	Close() bool
	//Name 驱动名称
	Name() string
	//UnInstall 卸载驱动
	UnInstall() bool
}
