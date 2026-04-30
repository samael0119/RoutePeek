package overview

func overviewText(lang string, key string) string {
	if copy, ok := overviewCopy[lang][key]; ok {
		return copy
	}
	if copy, ok := overviewCopy["zh"][key]; ok {
		return copy
	}
	return key
}

var overviewCopy = map[string]map[string]string{
	"zh": {
		"health_ok_label":          "正常",
		"health_ok_summary":        "当前没有发现会明显影响上网的风险。",
		"health_offline_label":     "可能断网",
		"health_offline_summary":   "发现可能导致网页打不开或应用无法联网的问题，建议先处理首要行动。",
		"health_attention_label":   "需注意",
		"health_attention_summary": "网络可能仍可使用，但存在影响稳定性、速度或访问范围的风险。",
		"health_info_summary":      "网络基本正常，有少量配置项可按需检查。",
		"node_device":              "本机",
		"node_gateway":             "网关",
		"node_internet":            "互联网",
		"node_proxy":               "代理",
		"node_vm":                  "虚拟机网络",
		"link_gateway":             "默认出口",
		"link_internet":            "外网访问",
		"link_dns":                 "域名解析",
		"link_proxy":               "代理转发",
		"link_vpn":                 "隧道",
		"link_vm":                  "本地虚拟网段",
		"desc_proxy_on":            "系统代理已开启",
		"desc_waiting_snapshot":    "等待网络快照",
		"desc_no_active_iface":     "未发现活动 IPv4 网卡",
		"desc_no_gateway":          "未发现默认网关",
		"desc_no_dns":              "未配置 DNS",
		"desc_no_public_ip":        "未获取到公网 IP",
		"action_default_impact":    "这个发现可能影响网络连通性或稳定性。",
		"action_default_step1":     "先记录当前网络状态。",
		"action_default_step2":     "按提示检查相关配置。",
		"action_default_step3":     "每次只改一个设置，方便回退。",
		"action_default_verify":    "修改后刷新 RoutePeek，并重新测试之前异常的网站或应用。",
	},
	"en": {
		"health_ok_label":          "Healthy",
		"health_ok_summary":        "No risks that clearly affect Internet access were found.",
		"health_offline_label":     "Likely offline",
		"health_offline_summary":   "RoutePeek found an issue that may prevent websites or apps from connecting. Start with the first action.",
		"health_attention_label":   "Needs attention",
		"health_attention_summary": "The network may still work, but there are risks that can affect stability, speed, or reachability.",
		"health_info_summary":      "The network looks generally healthy, with a few settings worth checking if needed.",
		"node_device":              "Device",
		"node_gateway":             "Gateway",
		"node_internet":            "Internet",
		"node_proxy":               "Proxy",
		"node_vm":                  "VM Network",
		"link_gateway":             "Default exit",
		"link_internet":            "Internet access",
		"link_dns":                 "Name lookup",
		"link_proxy":               "Proxy forwarding",
		"link_vpn":                 "Tunnel",
		"link_vm":                  "Local VM subnet",
		"desc_proxy_on":            "System proxy is enabled",
		"desc_waiting_snapshot":    "Waiting for network snapshot",
		"desc_no_active_iface":     "No active IPv4 interface found",
		"desc_no_gateway":          "No default gateway found",
		"desc_no_dns":              "No DNS configured",
		"desc_no_public_ip":        "No public IP detected",
		"action_default_impact":    "This finding may affect network connectivity or stability.",
		"action_default_step1":     "Record the current network state first.",
		"action_default_step2":     "Check the related settings from the finding.",
		"action_default_step3":     "Change one setting at a time so it is easy to roll back.",
		"action_default_verify":    "Refresh RoutePeek and retest the website or app that had problems.",
	},
}

func guidanceTemplates(lang string) map[string]guidanceTemplate {
	if lang == "en" {
		return map[string]guidanceTemplate{
			"NO_DEFAULT_GW": {
				category:   "connectivity",
				impact:     "The computer has no default path to the Internet, so most websites and apps may be unable to connect.",
				steps:      []string{"Confirm Ethernet, Wi-Fi, or hotspot is connected.", "Reconnect to the network, or restart the router and try again.", "On a managed network, ask the administrator to confirm the gateway settings."},
				verify:     "Open a common website, or refresh RoutePeek and check whether a default gateway appears.",
				targetNode: "gateway",
				confidence: "high",
			},
			"INVALID_GW": {
				category:   "connectivity",
				impact:     "The default gateway address is invalid, so traffic may be sent to a dead network exit.",
				steps:      []string{"Disconnect and reconnect to the current network.", "Check whether a gateway was entered manually.", "Prefer automatic IP and gateway assignment."},
				verify:     "Refresh RoutePeek and confirm the default gateway is no longer 0.0.0.0.",
				targetNode: "gateway",
				confidence: "high",
			},
			"NO_ACTIVE_IFACE": {
				category:   "connectivity",
				impact:     "No usable network adapter is active, so the computer may not have a path to any network.",
				steps:      []string{"Check whether Wi-Fi, Ethernet, or hotspot is connected.", "Confirm the network adapter is enabled in system settings.", "If this is a virtual machine, confirm the VM network adapter is attached."},
				verify:     "Refresh RoutePeek and confirm an active IPv4 interface appears.",
				targetNode: "device",
				confidence: "high",
			},
			"NO_ROUTES": {
				category:   "route",
				impact:     "RoutePeek could not see route table data, so it cannot fully explain which network exit the system will use.",
				steps:      []string{"Check whether networking is still starting up.", "If websites work, treat this as a collection limitation before changing settings.", "On managed machines, confirm route-reading commands are allowed."},
				verify:     "Refresh RoutePeek and confirm route entries appear in Details.",
				targetNode: "gateway",
				confidence: "medium",
			},
			"NO_DNS": {
				category:   "dns",
				impact:     "Without DNS, website names cannot be resolved to server addresses. Pages may fail even when direct IP access works.",
				steps:      []string{"Open system network settings and find the active network.", "Set DNS to automatic, or use trusted DNS such as 1.1.1.1 or 8.8.8.8.", "Save, then disconnect and reconnect to the network."},
				verify:     "Try opening example.com, or refresh RoutePeek and confirm DNS servers are listed.",
				targetNode: "dns",
				confidence: "high",
			},
			"SUSPICIOUS_DNS": {
				category:   "dns",
				impact:     "The DNS server is not a common public or private address. It may slow down lookup, redirect traffic, or indicate hijacking.",
				steps:      []string{"Confirm whether this DNS comes from your router, workplace, school, or VPN.", "If the source is unclear, switch to automatic DNS or a trusted public DNS.", "Clear browser cache and retry the affected website."},
				verify:     "Refresh RoutePeek, confirm the DNS address is expected, and retry the website that behaved oddly.",
				targetNode: "dns",
				confidence: "medium",
			},
			"VPN_GLOBAL_ROUTE": {
				category:   "vpn",
				impact:     "The VPN is taking over all traffic, which can change access to websites, company resources, or local devices.",
				steps:      []string{"If you only need company resources, check whether split tunneling is available.", "Temporarily disconnect the VPN and compare whether the app or website recovers.", "For long-term use, ask the VPN administrator to confirm the routing policy."},
				verify:     "Toggle VPN status, refresh RoutePeek, and compare the network path and access result.",
				targetNode: "vpn",
				confidence: "medium",
			},
			"VM_NO_ROUTE": {
				category:   "vm",
				impact:     "A VM network exists but has no matching route, so the host and VM may not reach each other.",
				steps:      []string{"Open the VM software and confirm the virtual network is enabled.", "Check whether the VM network mode matches your expectation.", "Restart the VM network adapter and retry."},
				verify:     "Refresh RoutePeek and confirm a route appears near the VM subnet.",
				targetNode: "vm",
				confidence: "medium",
			},
			"PROXY_INCOMPLETE": proxyGuidanceEN(),
			"PROXY_HTTPS_ONLY": proxyGuidanceEN(),
			"METRIC_CONFLICT": {
				category:   "route",
				impact:     "Multiple routes with the same priority can make the system switch between network exits and cause unstable connections.",
				steps:      []string{"Keep only the network connection you currently need.", "Temporarily disable unused Ethernet, Wi-Fi, or virtual adapters.", "In managed networks, ask the administrator to adjust route priority."},
				verify:     "Refresh RoutePeek and confirm the same destination no longer has multiple equal-priority exits.",
				targetNode: "gateway",
				confidence: "medium",
			},
			"MULTI_NIC_ACTIVE": {
				category:   "route",
				impact:     "When several physical network adapters are active, the system may choose the wrong exit. Some sites can be slow or unreachable.",
				steps:      []string{"Decide whether Wi-Fi or Ethernet should be the main connection.", "Temporarily disable unused network connections.", "If multiple networks are required, confirm the default route priority is intended."},
				verify:     "Refresh RoutePeek and confirm only needed adapters remain active, or the affected access has recovered.",
				targetNode: "gateway",
				confidence: "low",
			},
			"NO_PUBLIC_IP": {
				category:   "public_ip",
				impact:     "RoutePeek could not fetch the public IP. The Internet may be down, or the public-IP lookup service may simply be blocked.",
				steps:      []string{"First open a common website to see whether Internet access actually works.", "If websites work, this warning can usually be ignored.", "If websites fail, check gateway, DNS, proxy, or VPN next."},
				verify:     "Open a webpage, then refresh RoutePeek and check whether the public IP appears.",
				targetNode: "internet",
				confidence: "medium",
			},
		}
	}
	return map[string]guidanceTemplate{
		"NO_DEFAULT_GW": {
			category:   "connectivity",
			impact:     "电脑没有找到通往互联网的默认出口，网页和大多数应用可能都无法联网。",
			steps:      []string{"确认网线、Wi-Fi 或热点已经连接。", "重新连接当前网络，或重启路由器后再试。", "如果在公司或学校网络中，联系网络管理员确认网关配置。"},
			verify:     "重新打开一个常用网站，或再次刷新 RoutePeek 查看默认网关是否出现。",
			targetNode: "gateway",
			confidence: "high",
		},
		"INVALID_GW": {
			category:   "connectivity",
			impact:     "默认网关地址异常，电脑可能把流量发到了无效出口。",
			steps:      []string{"断开后重新连接当前网络。", "检查是否手动填写过网关地址。", "优先使用自动获取 IP 和网关。"},
			verify:     "刷新 RoutePeek，确认默认网关不再是 0.0.0.0。",
			targetNode: "gateway",
			confidence: "high",
		},
		"NO_ACTIVE_IFACE": {
			category:   "connectivity",
			impact:     "没有可用网卡处于活动状态，电脑可能还没有连接到任何网络。",
			steps:      []string{"确认 Wi-Fi、有线网络或热点已经连接。", "在系统网络设置里确认网卡没有被禁用。", "如果运行在虚拟机中，确认虚拟网卡已经连接。"},
			verify:     "刷新 RoutePeek，确认出现活动 IPv4 网卡。",
			targetNode: "device",
			confidence: "high",
		},
		"NO_ROUTES": {
			category:   "route",
			impact:     "RoutePeek 没有看到路由表，暂时无法完整判断系统会选择哪个网络出口。",
			steps:      []string{"确认系统网络服务已经启动完成。", "如果网页实际可用，先把它视为采集受限，不要急着改系统设置。", "如果是受管设备，确认读取路由表的命令没有被限制。"},
			verify:     "刷新 RoutePeek，并在详情页确认路由条目已经出现。",
			targetNode: "gateway",
			confidence: "medium",
		},
		"NO_DNS": {
			category:   "dns",
			impact:     "DNS 缺失会导致输入网址时无法找到对应服务器，表现为网页打不开但直连 IP 可能可用。",
			steps:      []string{"打开系统网络设置，找到当前正在使用的网络。", "将 DNS 改为自动获取，或填写可信 DNS，例如 1.1.1.1 / 8.8.8.8。", "保存后断开并重新连接网络。"},
			verify:     "尝试打开 example.com，或再次刷新 RoutePeek 确认 DNS 服务器已出现。",
			targetNode: "dns",
			confidence: "high",
		},
		"SUSPICIOUS_DNS": {
			category:   "dns",
			impact:     "当前 DNS 不在常见公共或内网地址范围内，可能导致解析变慢、跳转异常或访问被劫持。",
			steps:      []string{"确认这个 DNS 是否来自公司、学校、路由器或 VPN。", "如果来源不明确，切换到自动获取 DNS 或可信公共 DNS。", "清理浏览器缓存后重新访问异常网站。"},
			verify:     "刷新 RoutePeek，确认 DNS 地址符合预期，并检查之前异常的网站是否恢复。",
			targetNode: "dns",
			confidence: "medium",
		},
		"VPN_GLOBAL_ROUTE": {
			category:   "vpn",
			impact:     "VPN 正在接管全部网络流量，可能让国内外网站、公司内网或本地设备访问路径发生变化。",
			steps:      []string{"如果只是访问公司资源，检查 VPN 是否支持分流模式。", "临时断开 VPN 后对比网页或应用是否恢复。", "需要长期使用时，让 VPN 管理员确认路由策略。"},
			verify:     "切换 VPN 状态后刷新 RoutePeek，观察网络路径和访问结果是否变化。",
			targetNode: "vpn",
			confidence: "medium",
		},
		"VM_NO_ROUTE": {
			category:   "vm",
			impact:     "虚拟机网络存在但没有匹配路由，宿主机和虚拟机之间可能互相访问不到。",
			steps:      []string{"打开虚拟机软件，确认对应虚拟网络已启用。", "检查虚拟机网络模式是否与预期一致。", "重启虚拟机网络适配器后再试。"},
			verify:     "刷新 RoutePeek，确认虚拟机网段附近出现对应路由。",
			targetNode: "vm",
			confidence: "medium",
		},
		"PROXY_INCOMPLETE": proxyGuidanceZH(),
		"PROXY_HTTPS_ONLY": proxyGuidanceZH(),
		"METRIC_CONFLICT": {
			category:   "route",
			impact:     "多条同优先级路由可能让系统在不同网络出口之间摇摆，导致连接不稳定。",
			steps:      []string{"保留当前最需要的网络连接。", "临时关闭不使用的有线、无线或虚拟网卡。", "如果是办公环境，联系管理员调整路由优先级。"},
			verify:     "刷新 RoutePeek，确认同一目标不再出现多个相同优先级出口。",
			targetNode: "gateway",
			confidence: "medium",
		},
		"MULTI_NIC_ACTIVE": {
			category:   "route",
			impact:     "多个真实网卡同时联网时，系统可能选错出口，表现为某些网站慢或打不开。",
			steps:      []string{"确认当前主要使用 Wi-Fi 还是有线网络。", "临时关闭不用的网络连接。", "如果需要同时连接多个网络，确认默认路由优先级符合预期。"},
			verify:     "刷新 RoutePeek，确认只保留必要的活动网卡，或异常访问已经恢复。",
			targetNode: "gateway",
			confidence: "low",
		},
		"NO_PUBLIC_IP": {
			category:   "public_ip",
			impact:     "RoutePeek 没有获取到公网 IP，可能是外网不通，也可能只是公网查询服务被拦截。",
			steps:      []string{"先尝试打开一个常用网站确认是否真的无法上网。", "如果网站可用，可以暂时忽略这个提示。", "如果网站不可用，再检查网关、DNS、代理或 VPN。"},
			verify:     "打开网页后刷新 RoutePeek，确认公网 IP 是否出现。",
			targetNode: "internet",
			confidence: "medium",
		},
	}
}

func proxyGuidanceEN() guidanceTemplate {
	return guidanceTemplate{
		category:   "proxy",
		impact:     "Only part of the proxy configuration is set, so some apps may connect while others fail.",
		steps:      []string{"Open system or browser proxy settings.", "Confirm whether both HTTP and HTTPS should use the proxy.", "If a proxy is not needed, turn it off and retest."},
		verify:     "Reopen the affected app or website, then refresh RoutePeek and check proxy status.",
		targetNode: "proxy",
		confidence: "medium",
	}
}

func proxyGuidanceZH() guidanceTemplate {
	return guidanceTemplate{
		category:   "proxy",
		impact:     "代理只配置了一部分协议，可能出现部分应用能联网、部分应用不能联网。",
		steps:      []string{"打开系统或浏览器代理设置。", "确认 HTTP 和 HTTPS 是否都需要代理。", "如果不需要代理，关闭代理后重新测试。"},
		verify:     "重新打开之前异常的应用或网页，并刷新 RoutePeek 查看代理状态。",
		targetNode: "proxy",
		confidence: "medium",
	}
}
