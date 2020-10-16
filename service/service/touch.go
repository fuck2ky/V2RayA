package service

import (
	"github.com/v2rayA/v2rayA/db/configure"
)

func DeleteWhich(ws []*configure.Which) (err error) {
	var data configure.Whiches
	//对要删除的touch去重
	data.Set(ws)
	data.Set(data.GetNonDuplicated())
	//对要删除的touch排序，将大的下标排在前面，从后往前删
	data.Sort()
	touches := data.Get()
	cs := configure.GetConnectedServer()
	subscriptionsIndexes := make([]int, 0, len(ws))
	serversIndexes := make([]int, 0, len(ws))
	bDeletedSubscription := false
	bDeletedServer := false
	for _, v := range touches {
		ind := v.ID - 1
		switch v.TYPE {
		case configure.SubscriptionType: //这里删的是某个订阅
			//检查现在连接的结点是否在该订阅中，是的话断开连接
			if cs != nil && cs.TYPE == configure.SubscriptionServerType {
				if ind == cs.Sub {
					err = Disconnect()
					if err != nil {
						return
					}
				} else if ind < cs.Sub {
					cs.Sub -= 1
				}
			}
			subscriptionsIndexes = append(subscriptionsIndexes, ind)
			bDeletedSubscription = true
		case configure.ServerType:
			//检查现在连接的结点是否是该服务器，是的话断开连接
			if cs != nil && cs.TYPE == configure.ServerType {
				if v.ID == cs.ID {
					err = Disconnect()
					if err != nil {
						return
					}
				} else if v.ID < cs.ID {
					cs.ID -= 1
				}
			}
			serversIndexes = append(serversIndexes, ind)
			bDeletedServer = true
		case configure.SubscriptionServerType: //订阅的结点的不能删的
			continue
		}
	}
	if bDeletedSubscription {
		err = configure.RemoveSubscriptions(subscriptionsIndexes)
		if err != nil {
			return
		}
	}
	if bDeletedServer {
		err = configure.RemoveServers(serversIndexes)
		if err != nil {
			return
		}
	}
	if configure.GetConnectedServer() != nil { //如果已经disconnect了，就不要再set回去了
		err = configure.SetConnect(cs) //由于删除了一些servers或subscriptions，当前which可能会有下标上的变化
	}
	return
}
