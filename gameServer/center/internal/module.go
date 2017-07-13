package internal

import (
	"mj/common/cost"
	"mj/gameServer/base"
	"mj/gameServer/conf"

	"github.com/lovelly/leaf/chanrpc"
	"github.com/lovelly/leaf/module"
	"github.com/lovelly/leaf/nsq/cluster"
)

var (
	skeleton   = base.NewSkeleton()
	ChanRPC    = skeleton.ChanRPCServer
	Users      = make(map[int]*chanrpc.Server) //本服玩家
	OtherUsers = make(map[int]string)          //其他服登录的玩家  key is uid， values is NodeId
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
	cfg := &cluster.Cluster_config{
		LogLv:              "Debug",
		Channel:            conf.ServerNsqCahnnel(),
		Csmtopics:          []string{cost.GamePrefix, conf.ServerName()}, //需要订阅的主题
		CsmNsqdAddrs:       conf.Server.NsqdAddrs,
		CsmNsqLookupdAddrs: conf.Server.NsqLookupdAddrs,
		PdrNsqdAddr:        conf.Server.PdrNsqdAddr, //生产者需要连接的nsqd地址
		SelfName:           conf.ServerName(),
	}

	cluster.Start(cfg)
}

func (m *Module) OnDestroy() {
	cluster.Stop()
}
