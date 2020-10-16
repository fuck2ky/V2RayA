package entity

import (
	"log"
	"os/exec"
	"sync"
	"time"
	"v2ray.com/core/app/router"
	"github.com/v2rayA/v2rayA/common/netTools"
	"github.com/v2rayA/v2rayA/core/dnsPoison"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/global"
	"github.com/v2rayA/v2rayA/db/configure"
)

var (
	poison            = dnsPoison.New()
	done              chan interface{}
	mutex             sync.Mutex
	limit             = make(chan interface{}, 1)
	whiteDnsServerIps []*router.CIDR
	whiteDomains      []*router.Domain
	wg                sync.WaitGroup
)

type ExtraInfo struct {
	DohIps       []string
	DohDomains   []string
	ServerIps    []string
	ServerDomain string
}

func ShouldDnsPoisonOpen() bool {
	if setting := configure.GetSettingNotNil();
		!(setting.Transparent != configure.TransparentClose &&
			setting.AntiPollution != configure.AntipollutionClosed &&
			(!global.SupportTproxy || setting.EnhancedMode)) {
		//redirect+poison增强方案
		return false
	}
	return true
}

func CheckAndSetupDnsPoisonWithExtraInfo(info *ExtraInfo) {
	if !ShouldDnsPoisonOpen() {
		return
	}
	whitedms := make([]*router.Domain, 0, len(info.DohDomains))
	for _, h := range info.DohDomains {
		whitedms = append(whitedms, &router.Domain{
			Type:  router.Domain_Full,
			Value: h,
		})
	}
	if len(info.ServerDomain) > 0 {
		whitedms = append(whitedms, &router.Domain{
			Type:  router.Domain_Full,
			Value: info.ServerDomain,
		})
	}
	whitedms = append(whitedms, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "v2raya.mzz.pub",
	}, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "v.mzz.pub",
	}, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "github.com",
	}, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "1password.com",
	}, &router.Domain{
		Type:  router.Domain_Regex,
		Value: `^dns\.`,
	}, &router.Domain{
		Type:  router.Domain_Regex,
		Value: `^doh\.`,
	})
	_ = StartDNSPoison(nil,
		whitedms)
}

func StartDNSPoison(externWhiteDnsServers []*router.CIDR, externWhiteDomains []*router.Domain) (err error) {
	defer func() {
		if err != nil {
			err = newError("StartDNSPoison").Base(err)
		}
	}()
	mutex.Lock()
	if done != nil {
		select {
		case <-done:
			//done has closed
		default:
			mutex.Unlock()
			return newError("DNSPoison is running")
		}
	}
	done = make(chan interface{})
	whiteDnsServerIps = externWhiteDnsServers
	whiteDomains = externWhiteDomains
	mutex.Unlock()
	go func(poison *dnsPoison.DnsPoison) {
		//并发限制1
		select {
		case limit <- nil:
		default:
			return
		}
		defer func() { <-limit }()
	out:
		for {
			//随时准备应对default interface变化
			f := func() {
				ifnames, err := netTools.GetDefaultInterface()
				if err != nil {
					return
				}
				mIfnames := make(map[string]interface{})
				mHandles := make(map[string]interface{})
				needToAdd := false
				for _, ifname := range ifnames {
					mIfnames[ifname] = nil
					if !poison.Exists(ifname) {
						needToAdd = true
					}
				}
				hs := poison.ListHandles()
				for _, h := range hs {
					mHandles[h] = nil
					if _, ok := mIfnames[h]; !ok {
						_ = poison.DeleteHandles(h)
					}
				}
				if !needToAdd {
					return
				}
				//准备白名单
				log.Println("DnsPoison: preparing whitelist")
				wlDms, err := asset.GetWhitelistCn(nil, whiteDomains)
				//var wlDms = new(strmatcher.MatcherGroup)
				if err != nil {
					log.Println("StartDNSPoisonConroutine:", err)
					return
				}
				ipMatcher := new(router.GeoIPMatcher)
				_ = ipMatcher.Init(whiteDnsServerIps)
				for _, ifname := range ifnames {
					if _, ok := mHandles[ifname]; !ok {
						err = poison.Prepare(ifname)
						if err != nil {
							log.Println("StartDNSPoisonConroutine["+ifname+"]:", err)
							return
						}
						go func(ifname string) {
							wg.Add(1)
							defer wg.Done()
							err = poison.Run(ifname, ipMatcher, wlDms)
							if err != nil {
								log.Println("StartDNSPoisonConroutine["+ifname+"]:", err)
							}
						}(ifname)
					}
				}
			}
			f()
			select {
			case <-done:
				poison.Clear()
				break out
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}(poison)
	return nil
}

func StopDNSPoison() {
	mutex.Lock()
	defer mutex.Unlock()
	if done != nil {
		select {
		case <-done:
			// channel 'done' has been closed
		default:
			close(done)
		}
	}
	wg.Wait()
	clearDNSCache()
}

func clearDNSCache() {
	switch global.ServiceControlMode {
	case global.ServiceMode:
		_, _ = exec.Command("sh -c", "service nscd restart").Output()
	case global.SystemctlMode:
		_, _ = exec.Command("sh -c", "systemctl restart nscd").Output()
	}
}
