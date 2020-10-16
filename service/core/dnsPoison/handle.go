package dnsPoison

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"
	"github.com/v2rayA/v2rayA/common/netTools"
	"golang.org/x/net/dns/dnsmessage"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	v2router "v2ray.com/core/app/router"
	"v2ray.com/core/common/strmatcher"
)

type handle struct {
	dnsPoison *DnsPoison
	done      chan interface{}
	running   bool
	*pcapgo.EthernetHandle
	inspectedBlackDomains *domainBlacklist
	portCache             *portCache
}

func newHandle(dnsPoison *DnsPoison, ethernetHandle *pcapgo.EthernetHandle) *handle {
	return &handle{
		dnsPoison:             dnsPoison,
		done:                  make(chan interface{}),
		EthernetHandle:        ethernetHandle,
		inspectedBlackDomains: newDomainBlacklist(),
		portCache:             newPortCache(),
	}
}

type handleResult int

const (
	Pass handleResult = iota
	Spoofing
	AddBlacklist
	ProposeBlacklist
	RemoveBlacklist
	AgainstBlacklist
)

type domainHandleResult struct {
	domain string
	result handleResult
}

func (interfaceHandle *handle) handleSendMessage(m *dnsmessage.Message, sAddr, sPort, dAddr, dPort *gopacket.Endpoint, whitelistDomains *strmatcher.MatcherGroup) (ip [4]byte) {
	// dns请求一般只有一个question
	q := m.Questions[0]
	if (q.Type != dnsmessage.TypeA && q.Type != dnsmessage.TypeAAAA) ||
		q.Class != dnsmessage.ClassINET {
		return
	}
	dm := q.Name.String()
	dmNoTQDN := strings.TrimSuffix(dm, ".")
	if dm == "" {
		return
	} else if index := whitelistDomains.Match(dmNoTQDN); index > 0 {
		// whitelistDomains
		return
	} else if index := strings.Index(dmNoTQDN, "."); index <= 0 {
		// 跳过随机的顶级域名
		return
	}
	if q.Type == dnsmessage.TypeA {
		//不在已探测黑名单中的放行
		if !interfaceHandle.inspectedBlackDomains.Exists(q.Name.String()) {
			return
		}
	}
	//TODO: 不支持IPv6，AAAA投毒返回空，加速解析
	return interfaceHandle.poison(m, dAddr, dPort, sAddr, sPort)
}

func (interfaceHandle *handle) handleReceiveMessage(m *dnsmessage.Message) (results []*domainHandleResult, msg string) {
	//探测该域名是否被污染为本地回环
	spoofed := false
	emptyRecord := true
	var msgs []string
	q := m.Questions[0]
	dm := q.Name.String()
	dms := []string{dm}
	//CNAME has multiple answers including an AResource
	for _, a := range m.Answers {
		switch a := a.Body.(type) {
		case *dnsmessage.CNAMEResource:
			cname := a.CNAME.String()
			msgs = append(msgs, "CNAME:"+strings.TrimSuffix(cname, "."))
			dms = append(dms, cname)
		case *dnsmessage.AResource:
			msgs = append(msgs, "A:"+fmt.Sprintf("%d.%d.%d.%d", a.A[0], a.A[1], a.A[2], a.A[3]))
			if netTools.IsJokernet4(&a.A) {
				spoofed = true
			}
			emptyRecord = false
		}
	}
	if emptyRecord {
		//空记录不能影响白名单探测
		return nil, msg
	}
	defer func() {
		if results != nil {
			msg = "[" + strings.Join(msgs, ", ") + "]"
		}
	}()
	for _, d := range dms {
		exists := interfaceHandle.inspectedBlackDomains.Exists(d)
		if spoofed {
			if !exists {
				results = append(results, &domainHandleResult{domain: d, result: ProposeBlacklist})
			}
			if interfaceHandle.inspectedBlackDomains.Propose(d) {
				results = append(results, &domainHandleResult{domain: d, result: AddBlacklist})
			}
		}
	}
	return results, msg
}

func packetFilter(portCache *portCache, pPacket *gopacket.Packet, whitelistDnsServers *v2router.GeoIPMatcher) (m *dnsmessage.Message, pSAddr, pSPort, pDAddr, pDPort *gopacket.Endpoint) {
	packet := *pPacket
	trans := packet.TransportLayer()
	//跳过非传输层的包
	if trans == nil {
		return
	}
	transflow := trans.TransportFlow()
	sPort, dPort := transflow.Endpoints()
	//跳过非常规DNS端口53端口的包
	strDport := dPort.String()
	strSport := sPort.String()
	if strDport != "53" && strSport != "53" {
		return
	}
	sAddr, dAddr := packet.NetworkLayer().NetworkFlow().Endpoints()
	// TODO: 暂不支持IPv6
	sIp := net.ParseIP(sAddr.String()).To4()
	if len(sIp) != net.IPv4len {
		return
	}
	// Domain-Name-Server whitelistIps
	if ok := whitelistDnsServers.Match(sIp); ok {
		return
	}
	dIp := net.ParseIP(dAddr.String()).To4()
	if len(dIp) != net.IPv4len {
		return
	}
	// Domain-Name-Server whitelistIps
	if ok := whitelistDnsServers.Match(dIp); ok {
		return
	}
	//尝试解析为dnsmessage格式
	var dmessage dnsmessage.Message
	err := dmessage.Unpack(trans.LayerPayload())
	if err != nil {
		return
	}
	//跳过非A且非AAAA，或不包含"."的域名
	if len(dmessage.Questions) > 0 {
		name := dmessage.Questions[0].Name.String()
		if (dmessage.Questions[0].Type != dnsmessage.TypeA && dmessage.Questions[0].Type != dnsmessage.TypeAAAA) ||
			!strings.ContainsRune(strings.TrimSuffix(name, "."), '.') {
			return
		}
	}
	//跳过已处理过dns响应的端口的包
	portCache.Lock()
	if strDport != "53" && portCache.Exists(localPort(strDport)) {
		portCache.Unlock()
		return
	}
	//一个本地port记录5秒
	portCache.Set(localPort(strDport), 5*time.Second)
	portCache.Unlock()
	return &dmessage, &sAddr, &sPort, &dAddr, &dPort
}

func (interfaceHandle *handle) handlePacket(packet gopacket.Packet, ifname string, whitelistDnsServers *v2router.GeoIPMatcher, whitelistDomains *strmatcher.MatcherGroup) {
	m, sAddr, sPort, dAddr, dPort := packetFilter(interfaceHandle.portCache, &packet, whitelistDnsServers)
	if m == nil {
		return
	}
	// dns请求一般只有一个question
	dm := m.Questions[0].Name.String()
	if !m.Response {
		ip := interfaceHandle.handleSendMessage(m, sAddr, sPort, dAddr, dPort, whitelistDomains)
		// TODO: 不显示AAAA的投毒，因为暂时不支持IPv6
		if ip[3] != 0 && m.Questions[0].Type == dnsmessage.TypeA {
			log.Println("dnsPoison["+ifname+"]:", sAddr.String()+":"+sPort.String(), "->", dAddr.String()+":"+dPort.String(), dm, "poisoned as", fmt.Sprintf("%v.%v.%v.%v", ip[0], ip[1], ip[2], ip[3]))
		}
	} else {
		results, msg := interfaceHandle.handleReceiveMessage(m)
		if results != nil {
			log.Println("dnsPoison["+ifname+"]:", dAddr.String()+":"+dPort.String(), "<-", sAddr.String()+":"+sPort.String(), "("+dm, "=>", msg+")")
			for _, r := range results {
				// print log
				switch r.result {
				case ProposeBlacklist:
					log.Println("dnsPoison["+ifname+"]: [propose]", r.domain, "proof:", dm+msg)
				case AgainstBlacklist:
					log.Println("dnsPoison["+ifname+"]: [against]", r.domain, "proof:", dm+msg)
				case AddBlacklist:
					log.Println("dnsPoison["+ifname+"]: {add blocklist}", r.domain)
				case RemoveBlacklist:
					log.Println("dnsPoison["+ifname+"]: {remove blocklist}", r.domain)
				}
			}
		}
	}
}

func (interfaceHandle *handle) poison(m *dnsmessage.Message, lAddr, lPort, rAddr, rPort *gopacket.Endpoint) (ip [4]byte) {
	q := m.Questions[0]
	m.RCode = dnsmessage.RCodeSuccess
	switch q.Type {
	case dnsmessage.TypeAAAA:
		//返回空回答
	case dnsmessage.TypeA:
		//对A查询返回一个公网地址以使得后续tcp连接经过网关嗅探，以dns污染解决dns污染
		ip = interfaceHandle.dnsPoison.reservedIpPool.Lookup(q.Name.String())
		m.Answers = []dnsmessage.Resource{{
			Header: dnsmessage.ResourceHeader{
				Name:  q.Name,
				Type:  q.Type,
				Class: q.Class,
				TTL:   0,
			},
			Body: &dnsmessage.AResource{A: ip},
		}}
	}
	m.Response = true
	m.RecursionAvailable = true
	// write back
	go func(m *dnsmessage.Message) {
		packed, _ := m.Pack()
		lport, _ := strconv.Atoi(lPort.String())
		conn, err := newDialer(lAddr.String(), uint32(lport), 30*time.Second).Dial("udp", rAddr.String()+":"+rPort.String())
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write(packed)
	}(m)
	return
}
