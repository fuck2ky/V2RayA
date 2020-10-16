package service

import (
	"bytes"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/nodeData"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/db/configure"
	"log"
	"net/http"
	"strings"
	"time"
	"v2ray.com/core/common/errors"
)

//func ResolveSubscription(source string) (infos []*nodeData.NodeData, err error) {
//	return ResolveSubscriptionWithClient(source, http.DefaultClient)
//}

func ResolveSubscriptionWithClient(source string, client *http.Client) (infos []*nodeData.NodeData, status string, err error) {
	// get请求source
	c := *client
	c.Timeout = 30 * time.Second
	res, err := httpClient.HttpGetUsingSpecificClient(&c, source)
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	defer res.Body.Close()
	// base64解码, raw是多行vmess/ss/ssr/trojan
	raw, err := common.Base64StdDecode(buf.String())
	if err != nil {
		raw, _ = common.Base64URLDecode(buf.String())
	}
	// 切分raw
	rows := strings.Split(strings.TrimSpace(raw), "\n")
	// 解析
	infos = make([]*nodeData.NodeData, 0)
	for _, row := range rows {
		if strings.HasPrefix(row, "STATUS=") {
			status = strings.TrimPrefix(row, "STATUS=")
			continue
		}
		var data *nodeData.NodeData
		data, err = ResolveURL(row)
		if err != nil {
			if errors.Cause(err) != ErrorEmptyAddress {
				log.Println(row, err)
			}
			err = nil
			continue
		}
		infos = append(infos, data)
	}
	return
}

func UpdateSubscription(index int, disconnectIfNecessary bool) (err error) {
	subscriptions := configure.GetSubscriptions()
	addr := subscriptions[index].Address
	c, err := httpClient.GetHttpClientAutomatically()
	if err != nil {
		reason := "failed to get proxy"
		return newError(reason)
	}
	infos, status, err := ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		reason := "failed to resolve subscription address: " + err.Error()
		log.Println(infos, err)
		return newError(reason)
	}
	tsrs := make([]configure.ServerRaw, len(infos))
	var connectedServer *configure.ServerRaw
	cs := configure.GetConnectedServer()
	toFindConnectedServer := false
	var found bool
	if cs != nil {
		connectedServer, _ = cs.LocateServer()
		if connectedServer != nil && cs.TYPE == configure.SubscriptionServerType && cs.Sub == index {
			toFindConnectedServer = true
			found = false
		}
	}
	//将列表更换为新的，并且找到一个跟现在连接的server值相等的，设为Connected，如果没有，则断开连接
	for i, info := range infos {
		tsr := configure.ServerRaw{
			VmessInfo: info.VmessInfo,
		}
		if toFindConnectedServer && connectedServer.VmessInfo == tsr.VmessInfo {
			err = configure.SetConnect(&configure.Which{
				TYPE:    configure.SubscriptionServerType,
				ID:      i + 1,
				Sub:     index,
				Latency: "",
			})
			if err != nil {
				return
			}
			toFindConnectedServer = false
			found = true
		}
		tsrs[i] = tsr
	}
	if toFindConnectedServer && !found {
		if disconnectIfNecessary {
			err = Disconnect()
			if err != nil {
				reason := "failed to disconnect previous server"
				return newError(reason)
			}
		} else if connectedServer != nil {
			//将之前连接的节点append进去
			tsrs = append(tsrs, *connectedServer)
			cs.ID = len(tsrs)
			err = configure.SetConnect(cs)
			if err != nil {
				return
			}
		}
	}
	subscriptions[index].Servers = tsrs
	subscriptions[index].Status = string(touch.NewUpdateStatus())
	subscriptions[index].Info = status
	return configure.SetSubscription(index, &subscriptions[index])
}

func ModifySubscriptionRemark(subscription touch.Subscription) (err error) {
	raw := configure.GetSubscription(subscription.ID - 1)
	if raw == nil {
		return newError("failed to find the corresponding subscription")
	}
	raw.Remarks = subscription.Remarks
	return configure.SetSubscription(subscription.ID-1, raw)
}
