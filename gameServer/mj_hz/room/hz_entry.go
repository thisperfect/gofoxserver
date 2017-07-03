package room

import (
	"mj/common/msg/mj_hz_msg"
	"mj/gameServer/common/mj/mj_base"
	"mj/gameServer/db/model"
)

func NewHZEntry(info *model.CreateRoomInfo) *hz_entry {
	e := new(hz_entry)
	e.Mj_base = mj_base.NewMJBase(info)
	return e
}

type hz_entry struct {
	*mj_base.Mj_base
}

func (e *hz_entry) ZhaMa(args []interface{}) {
	recvMsg := args[0].(*mj_hz_msg.C2G_HZMJ_ZhaMa)
	e.DataMgr.OnZhuaHua()
}
