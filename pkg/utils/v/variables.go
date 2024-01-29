package v

// Mb转化成Kb的单位大小
const Mb2KbUnitSize = 1024

// API版本号的定义
const (
	APIVersionV1 = "v1"
	APIVersionV2 = "v2"
	APIVersionV3 = "v3"
)

// 站点本地带宽的上下架
const (
	SiteLocalBandwidthEnabledStatus  = "enabled"  // 本地带宽上架状态
	SiteLocalBandwidthDisabledStatus = "disabled" // 本地带宽下架状态
)

// 本地带宽的计量规则的方向
const (
	MeterRuleEgressDirection  = "egress"  // 出口方向
	MeterRuleIngressDirection = "ingress" // 入口方向
)

// 是否打印curl相关的日志， "true": 表示允许打印， "false": 表示禁止打印
const (
	LogCurlEnable  = "true"
	LogCurlDisable = "false"
)

// 分布式网络类型
const (
	NetworkBackend5G      = "5g"
	NetworkBackendGeneral = "general"
)

// 站点类别
const (
	SiteCategoryCenter = "center"
	SiteCategoryEdge   = "edge"
)

// 分布式资源类型
const (
	// 通用资源类型
	CategoryGeneral = "general"
	// 5G资源类型
	Category5G = "5g"
	// 所有资源类型
	CategoryAll = "all"
)

// DefaultBatchSize 默认批量查询记录数
const DefaultBatchSize = 500

// 站点自动化配置相关的账号与用户
const (
	SiteAutoConfigAccount = "site_auto_config"
	SiteAutoConfigUser    = "site_auto_config"
)

// 站点认证类型
const (
	AuthTypeOpenstack = "openstack"
	AuthTypeYC        = "yunchang"
	AuthTypeMec       = "mec"
)

// ISP lists
const (
	IspChinaUnicom  = "ChinaUnicom"
	IspChinaTelecom = "ChinaTelecom"
	IspChinaMobile  = "ChinaMobile"
)

// DCS用户类型
const (
	YunjingUser = 0 // 云警管理员
	AdminUser   = 1 // 内部 admin user
	NormalUser  = 2 // 普通用户
)

// 请求数据来源
const (
	SourceWeaver = "weaver"
)
