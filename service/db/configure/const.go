package configure

type (
	AutoUpdateMode          string
	ProxyMode               string
	PacMode                 string
	PacRuleType             string
	PacMatchType            string
	RoutingDefaultProxyMode string
	TouchType               string
	DefaultYesNo            string
	TransparentMode         string
	Antipollution           string
)

const (
	TransparentClose     = TransparentMode("close")
	TransparentProxy     = TransparentMode("proxy")
	TransparentWhitelist = TransparentMode("whitelist")
	TransparentGfwlist   = TransparentMode("gfwlist")
	TransparentPac       = TransparentMode("pac")

	Default = DefaultYesNo("default")
	Yes     = DefaultYesNo("yes")
	No      = DefaultYesNo("no")

	NotAutoUpdate = AutoUpdateMode("none")
	AutoUpdate    = AutoUpdateMode("auto_update")

	ProxyModeDirect = ProxyMode("direct")
	ProxyModePac    = ProxyMode("pac")
	ProxyModeProxy  = ProxyMode("proxy")

	WhitelistMode = PacMode("whitelist")
	GfwlistMode   = PacMode("gfwlist")
	CustomMode    = PacMode("custom")
	RoutingAMode  = PacMode("routingA")

	DirectRule = PacRuleType("direct")
	ProxyRule  = PacRuleType("proxy")
	BlockRule  = PacRuleType("block")

	DomainMatchRule = PacMatchType("domain")
	IpMatchRule     = PacMatchType("ip")

	DefaultDirectMode = RoutingDefaultProxyMode("direct")
	DefaultProxyMode  = RoutingDefaultProxyMode("proxy")
	DefaultBlockMode  = RoutingDefaultProxyMode("block")

	SubscriptionType       = TouchType("subscription")
	ServerType             = TouchType("server")
	SubscriptionServerType = TouchType("subscriptionServer")

	DnsForward          = Antipollution("dnsforward")
	DoH                 = Antipollution("doh")
	AntipollutionNone   = Antipollution("none") //历史原因，none代表“仅防止dns劫持”，不代表关闭
	AntipollutionClosed = Antipollution("closed") //直接iptables略过udp
)
