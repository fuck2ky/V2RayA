package touch

import (
	"fmt"
	"github.com/v2rayA/v2rayA/db/configure"
	"net/url"
	"strings"
	"time"
)

/*
Touch是树型结构的前后端通信形式，其结构设计和前端统一。
*/
type SubscriptionStatus string
type Touch struct {
	Servers         []TouchServer    `json:"servers"`
	Subscriptions   []Subscription   `json:"subscriptions"`
	ConnectedServer *configure.Which `json:"connectedServer"` //冗余一个信息，方便查找
}
type TouchServer struct {
	ID          int                 `json:"id"`
	TYPE        configure.TouchType `json:"_type"`
	Name        string              `json:"name"`
	Address     string              `json:"address"`
	Net         string              `json:"net"`
	Connected   bool                `json:"connected"`
	PingLatency string              `json:"pingLatency"`
}
type Subscription struct {
	Remarks string              `json:"remarks,omitempty"`
	ID      int                 `json:"id"`
	TYPE    configure.TouchType `json:"_type"`
	Host    string              `json:"host"`
	Status  SubscriptionStatus  `json:"status"`
	Info    string              `json:"info"`
	Servers []TouchServer       `json:"servers"`
}

func NewUpdateStatus() SubscriptionStatus {
	return SubscriptionStatus(time.Now().Local().Format("2006-1-2 15:04:05"))
}
func NewUpdateFailStatus(reason string) SubscriptionStatus {
	return SubscriptionStatus(time.Now().Local().Format("2006-1-2 15:04:05") + "尝试更新失败：" + reason)
}

/* 将[]TouchServerRaw映射到[]TouchServer */
func serverRawsToServers(rss []configure.ServerRaw) (ts []TouchServer) {
	w := configure.GetConnectedServer()
	var tsr *configure.ServerRaw
	var err error
	if w != nil {
		tsr, err = w.LocateServer()
	}
	ts = make([]TouchServer, len(rss))
	for i, v := range rss {
		if v.VmessInfo.Protocol == "" {
			v.VmessInfo.Protocol = "vmess"
		}
		protocol := strings.ToUpper(v.VmessInfo.Protocol)
		var protoToShow string
		switch v.VmessInfo.Protocol {
		case "", "vmess", "vless":
			if v.VmessInfo.TLS == "tls" {
				protoToShow = fmt.Sprintf("%v(%v+tls)", protocol, v.VmessInfo.Net)
				break
			}
			fallthrough
		case "ss", "shadowsocks":
			protoToShow = fmt.Sprintf("%v(%v)", protocol, v.VmessInfo.Net)
		case "ssr", "shadowsocksr":
			protoToShow = fmt.Sprintf("%v(%v)", protocol, v.VmessInfo.Type)
		default:
			protoToShow = protocol
		}
		var address string
		if v.VmessInfo.Port == "" {
			address = v.VmessInfo.Add
		} else {
			address = v.VmessInfo.Add + ":" + v.VmessInfo.Port
		}
		ts[i] = TouchServer{
			ID:        i + 1,
			Name:      v.VmessInfo.Ps,
			Address:   address,
			Net:       protoToShow,
			Connected: w != nil && err == nil && &tsr.VmessInfo == &v.VmessInfo,
		}
	}
	return
}

/* 根据Configure创建一个Touch */
func GenerateTouch() (t Touch) {
	t.Servers = serverRawsToServers(configure.GetServers())
	subscriptions := configure.GetSubscriptions()
	t.Subscriptions = make([]Subscription, len(subscriptions))
	for i, v := range subscriptions {
		u, err := url.Parse(v.Address)
		if err != nil {
			continue
		}
		t.Subscriptions[i] = Subscription{
			Remarks: v.Remarks,
			ID:      i + 1,
			Host:    u.Host,
			Status:  SubscriptionStatus(v.Status),
			Servers: serverRawsToServers(v.Servers),
			Info:    v.Info,
		}
	}
	t.ConnectedServer = configure.GetConnectedServer()
	//补充TYPE
	for i := range t.Subscriptions {
		t.Subscriptions[i].TYPE = configure.SubscriptionType
		for j := range t.Subscriptions[i].Servers {
			t.Subscriptions[i].Servers[j].TYPE = configure.SubscriptionServerType
		}
	}
	for i := range t.Servers {
		t.Servers[i].TYPE = configure.ServerType
	}
	return
}
