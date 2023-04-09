package registry

type Registration struct {
	// 服务内容
	ServiceName ServiceName
	ServiceURL  string
	// 依赖内容
	RequiredServices []ServiceName
	// 选取一个URL，正常应该是多个
	ServiceUpdateURL string
	// 心跳服务
	HeartBeatURL string
}

type ServiceName string

// 枚举类型
const (
	LogService     = ServiceName("LogService")
	GradingService = ServiceName("GradingService")
	PortalService  = ServiceName("Portald")
)

// 定义发送用的变化信息
type pathEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []pathEntry
	Removed []pathEntry
}
