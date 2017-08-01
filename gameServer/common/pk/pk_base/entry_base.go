package pk_base

import (
	"errors"
	. "mj/common/cost"
	"mj/common/msg"
	"mj/common/msg/nn_tb_msg"
	. "mj/gameServer/common"
	"mj/gameServer/common/pk"
	"mj/gameServer/common/room_base"
	"mj/gameServer/conf"
	"mj/gameServer/db/model"
	"mj/gameServer/db/model/base"
	"mj/gameServer/user"
	"time"

	"mj/gameServer/db/model/stats"

	"github.com/lovelly/leaf/log"
	"github.com/lovelly/leaf/timer"
)

//创建的配置文件
type NewPKCtlConfig struct {
	BaseMgr  room_base.BaseManager
	TimerMgr room_base.TimerManager
	UserMgr  room_base.UserManager
	DataMgr  pk.DataManager
	LogicMgr pk.LogicManager
}

//消息入口文件
type Entry_base struct {
	room_base.BaseManager
	UserMgr  room_base.UserManager
	TimerMgr room_base.TimerManager

	DataMgr  pk.DataManager
	LogicMgr pk.LogicManager

	Temp            *base.GameServiceOption //模板
	Status          int
	IsClose         bool
	DelayCloseTimer *timer.Timer

	BtCardSpecialData []int
}

func NewPKBase(info *model.CreateRoomInfo) *Entry_base {
	Temp, ok1 := base.GameServiceOptionCache.Get(info.KindId, info.ServiceId)
	log.Debug("new pk base %d %d", info.KindId, info.ServiceId)
	if !ok1 {
		log.Error("at NewPKBase not foud config .... ")
		return nil
	}

	pk := new(Entry_base)
	pk.Temp = Temp
	return pk
}

func (r *Entry_base) RegisterBaseFunc() {
	r.GetChanRPC().Register("Sitdown", r.Sitdown)
	r.GetChanRPC().Register("UserStandup", r.UserStandup)
	r.GetChanRPC().Register("GetUserChairInfo", r.GetUserChairInfo)
	r.GetChanRPC().Register("DissumeRoom", r.DissumeRoom)
	r.GetChanRPC().Register("UserReady", r.UserReady)
	r.GetChanRPC().Register("userRelogin", r.UserReLogin)
	r.GetChanRPC().Register("userOffline", r.UserOffline)
	r.GetChanRPC().Register("SetGameOption", r.SetGameOption)
	r.GetChanRPC().Register("ReqLeaveRoom", r.ReqLeaveRoom)
	r.GetChanRPC().Register("ReplyLeaveRoom", r.ReplyLeaveRoom)
	r.GetChanRPC().Register("AddPlayCnt", r.AddPlayCnt)
}

func (r *Entry_base) Init(cfg *NewPKCtlConfig) {
	r.UserMgr = cfg.UserMgr
	r.DataMgr = cfg.DataMgr
	r.BaseManager = cfg.BaseMgr
	r.LogicMgr = cfg.LogicMgr
	r.TimerMgr = cfg.TimerMgr
	r.RoomRun(r.DataMgr.GetRoomId())
	r.DataMgr.OnCreateRoom()
	logInfo := make(map[string]interface{})
	myLogInfo := make(map[string]interface{})
	AddLogDb := stats.RoomLogOp
	logInfo["room_id"] = r.DataMgr.GetRoomId()
	logInfo["kind_id"] = r.Temp.KindID
	logInfo["service_id"] = r.Temp.ServerID
	logData, err1 := AddLogDb.GetByMap(logInfo)
	if err1 != nil {
		log.Error("Select Data from recode Error:%v", err1.Error())
	}
	r.TimerMgr.StartCreatorTimer(func() {
		log.Debug("not start game close ")
		myLogInfo["timeout_nostart"] = 1
		now := time.Now()
		myLogInfo["end_time"] = &now
		log.Debug("pk超时未开启ddebug======================================================%d", r.DataMgr.GetRoomId())
		myLogInfo["start_endError"] = 1
		err := AddLogDb.UpdateWithMap(logData.RecodeId, myLogInfo)
		if err != nil {
			log.Error("pk超时未开启更新失败：%s", err.Error())
		}
		r.OnEventGameConclude(0, nil, GER_DISMISS)
	})

}

func (r *Entry_base) GetRoomId() int {
	return r.DataMgr.GetRoomId()
}

//坐下
func (r *Entry_base) Sitdown(args []interface{}) {
	chairID := args[0].(int)
	u := args[1].(*user.User)

	retcode := 0
	defer func() {
		u.WriteMsg(&msg.G2C_UserSitDownRst{Code: retcode})
	}()
	if r.Status == RoomStatusStarting && r.Temp.DynamicJoin == 1 {
		retcode = GameIsStart
		return
	}

	retcode = r.UserMgr.Sit(u, chairID)

}
func (r *Entry_base) AddPlayCnt(args []interface{}) (interface{}, error) {
	if r.IsClose {
		return nil, errors.New("room is close ")
	}
	addCnt := args[0].(int)
	r.TimerMgr.AddMaxPlayCnt(addCnt)
	if r.DelayCloseTimer != nil {
		r.DelayCloseTimer.Stop()
		r.DelayCloseTimer = nil
	}
	return nil, nil
}

//起立
func (r *Entry_base) UserStandup(args []interface{}) {
	//recvMsg := args[0].(*msg.C2G_UserStandup{})
	u := args[1].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode))
		}
	}()

	if r.Status == RoomStatusStarting {
		retcode = ErrGameIsStart
		return
	}

	r.UserMgr.Standup(u)
}

//获取对方信息
func (room *Entry_base) GetUserChairInfo(args []interface{}) {
	recvMsg := args[0].(*msg.C2G_REQUserChairInfo)
	u := args[1].(*user.User)
	info := room.UserMgr.GetUserInfoByChairId(recvMsg.ChairID).(*msg.G2C_UserEnter)
	if info == nil {
		log.Error("at GetUserChairInfo no foud tagUser %v, userId:%d", args[0], u.Id)
		return
	}
	u.WriteMsg(info)
}

//玩家离开房间
func (room *Entry_base) ReqLeaveRoom(args []interface{}) {
	player := args[0].(*user.User)
	leaveFunc := func() {
		if room.UserMgr.LeaveRoom(player, room.Status) {
			player.WriteMsg(&msg.G2C_LeaveRoomRsp{})
		} else {
			player.WriteMsg(&msg.G2C_LeaveRoomRsp{Code: ErrLoveRoomFaild})
		}
		room.UserMgr.DeleteReply(player.Id)
	}
	if room.Status == RoomStatusReady {
		leaveFunc()
	} else {
		room.UserMgr.SendMsgAllNoSelf(player.Id, &msg.G2C_LeaveRoomBradcast{UserID: player.Id})
		room.TimerMgr.StartReplytIimer(player.Id, func() {
			room.OnEventGameConclude(player.ChairId, player, USER_LEAVE)
		})
	}
}

//其他玩家响应玩家离开房间的请求
func (room *Entry_base) ReplyLeaveRoom(args []interface{}) {
	player := args[0].(*user.User)
	Agree := args[1].(bool)
	ReplyUid := args[2].(int64)
	ret := room.UserMgr.ReplyLeave(player, Agree, ReplyUid, room.Status)
	if ret == 1 {
		room.OnEventGameConclude(player.ChairId, player, USER_LEAVE)
	} else if ret == 0 {
		room.TimerMgr.StopReplytIimer(ReplyUid)
	}
}

//解散房间
func (room *Entry_base) DissumeRoom(args []interface{}) {
	u := args[0].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode, "解散房间失败."))
		}
	}()

	if !room.DataMgr.CanOperatorRoom(u.Id) {
		retcode = NotOwner
		return
	}

	room.UserMgr.ForEachUser(func(u *user.User) {
		room.UserMgr.LeaveRoom(u, room.Status)
	})

	room.OnEventGameConclude(0, nil, GER_DISMISS)
	room.Destroy(room.DataMgr.GetRoomId())
	logInfo := make(map[string]interface{})
	myLogInfo := make(map[string]interface{})
	AddLogDb := stats.RoomLogOp
	logInfo["room_id"] = room.DataMgr.GetRoomId()
	logInfo["kind_id"] = room.Temp.KindID
	logInfo["service_id"] = room.Temp.ServerID
	logData, err1 := AddLogDb.GetByMap(logInfo)
	if err1 != nil {
		log.Error("Select Data from recode Error:%v", err1.Error())
	}
	now := time.Now()
	myLogInfo["end_time"] = &now
	if retcode != 0 && u != nil {
		myLogInfo["start_endError"] = 1
	}
	err := AddLogDb.UpdateWithMap(logData.RecodeId, myLogInfo)
	if err != nil {
		log.Error("pk结束时间和结束状态记录更新失败：%s", err.Error())
	}
}

//玩家准备
func (room *Entry_base) UserReady(args []interface{}) {
	u := args[1].(*user.User)
	if u.Status == US_READY {
		log.Debug("user status is ready at UserReady")
		return
	}

	log.Debug("at UserReady")
	room.UserMgr.SetUsetStatus(u, US_READY)

	if room.UserMgr.IsAllReady() {
		log.Debug("all user are ready start game")
		room.Status = RoomStatusStarting
		room.TimerMgr.StopCreatorTimer()
		//派发初始扑克
		room.TimerMgr.AddPlayCount()
		room.DataMgr.BeforeStartGame(room.UserMgr.GetCurPlayerCnt())
		room.DataMgr.StartGameing()
		room.DataMgr.AfterStartGame()
	}
}

//玩家重登
func (room *Entry_base) UserReLogin(args []interface{}) error {
	u := args[0].(*user.User)
	roomUser := room.getRoomUser(u.Id)
	if roomUser == nil {
		return errors.New(" UserReLogin not old user ")
	}
	log.Debug("at ReLogin have old user ")
	u.ChairId = roomUser.ChairId
	u.RoomId = roomUser.RoomId
	u.Status = roomUser.Status
	u.ChatRoomId = roomUser.ChatRoomId
	room.UserMgr.ReLogin(u, room.Status)
	return nil
}

//玩家离线
func (room *Entry_base) UserOffline(args []interface{}) {
	u := args[0].(*user.User)
	log.Debug("at UserOffline .... uid:%d", u.Id)
	if u.Status == US_READY {
		log.Debug("user status is ready at UserReady")
		return
	}

	room.UserMgr.SetUsetStatus(u, US_OFFLINE)
	room.TimerMgr.StartKickoutTimer(u.Id, func() {
		room.OffLineTimeOut(u)
	})
}

//离线超时踢出
func (room *Entry_base) OffLineTimeOut(u *user.User) {
	room.UserMgr.LeaveRoom(u, room.Status)
	if room.Status != RoomStatusReady {
		room.OnEventGameConclude(0, nil, GER_DISMISS)
	} else {
		if room.UserMgr.GetCurPlayerCnt() == 0 { //没人了直接销毁
			room.Destroy(room.DataMgr.GetRoomId())
		}
	}
}

//获取房间基础信息
func (room *Entry_base) GetBirefInfo() *msg.RoomInfo {
	BirefInf := &msg.RoomInfo{}
	BirefInf.ServerID = room.Temp.ServerID
	BirefInf.KindID = room.Temp.KindID
	BirefInf.NodeID = conf.Server.NodeId
	BirefInf.SvrHost = conf.Server.WSAddr
	BirefInf.PayType = room.UserMgr.GetPayType()
	BirefInf.RoomID = room.DataMgr.GetRoomId()
	BirefInf.CurCnt = room.UserMgr.GetCurPlayerCnt()
	BirefInf.MaxPlayerCnt = room.UserMgr.GetMaxPlayerCnt() //最多多人数
	BirefInf.PayCnt = room.TimerMgr.GetMaxPlayCnt()        //可玩局数
	BirefInf.CurPayCnt = room.TimerMgr.GetPlayCount()      //已玩局数
	BirefInf.CreateTime = room.TimerMgr.GetCreatrTime()    //创建时间
	//BirefInf.CreateUserId = room.DataMgr.GetCreater()
	BirefInf.IsPublic = room.UserMgr.IsPublic()
	BirefInf.Players = make(map[int64]*msg.PlayerBrief)
	BirefInf.MachPlayer = make(map[int64]struct{})
	return BirefInf
}

//游戏配置
func (room *Entry_base) SetGameOption(args []interface{}) {
	u := args[1].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode))
		}
	}()

	if u.ChairId == INVALID_CHAIR {
		retcode = ErrNoSitdowm
		return
	}

	AllowLookon := 0
	if u.Status == US_LOOKON {
		AllowLookon = 1
	}
	u.WriteMsg(&msg.G2C_GameStatus{
		GameStatus:  room.Status,
		AllowLookon: AllowLookon,
	})

	room.DataMgr.SendPersonalTableTip(u)

	if room.Status == RoomStatusReady { // 没开始
		room.DataMgr.SendStatusReady(u)
	} else { //开始了
		//把所有玩家信息推送给自己
		room.UserMgr.SendUserInfoToSelf(u)
		room.DataMgr.SendStatusPlay(u)
	}
}

//游戏结束
func (room *Entry_base) OnEventGameConclude(ChairId int, user *user.User, cbReason int) {
	switch cbReason {
	case GER_NORMAL: //常规结束
		room.DataMgr.NormalEnd(cbReason)
		room.AfterEnd(false)
		return
	case GER_DISMISS: //游戏解散
		room.DataMgr.DismissEnd(cbReason)
		room.AfterEnd(true)
	case USER_LEAVE: //用户请求解散
		room.DataMgr.NormalEnd(cbReason)
		room.AfterEnd(true)
	}
	log.Error("at OnEventGameConclude error  ")
	return
}

// 如果这里不能满足 afertEnd 请重构这个到个个组件里面
func (room *Entry_base) AfterEnd(Forced bool) {
	if Forced || room.TimerMgr.GetPlayCount() >= room.TimerMgr.GetMaxPlayCnt() {
		if room.DelayCloseTimer != nil {
			room.DelayCloseTimer.Stop()
		}
		room.DelayCloseTimer = room.AfterFunc(time.Duration(GetGlobalVarInt(DelayDestroyRoom))*time.Second, func() {
			room.IsClose = true
			log.Debug("Forced :%v, PlayTurnCount:%v, temp PlayTurnCount:%d", Forced, room.TimerMgr.GetPlayCount(), room.TimerMgr.GetMaxPlayCnt())
			room.UserMgr.SendMsgToHallServerAll(&msg.RoomEndInfo{
				RoomId: room.DataMgr.GetRoomId(),
				Status: room.Status,
			})
			room.Destroy(room.DataMgr.GetRoomId())
			room.UserMgr.RoomDissume()
		})
		return
	}

	room.UserMgr.ForEachUser(func(u *user.User) {
		room.UserMgr.SetUsetStatus(u, US_SIT)
	})
}

//计算税收 暂时未实现
func (room *Entry_base) CalculateRevenue(ChairId, lScore int) int {
	//效验参数

	UserCnt := room.UserMgr.GetMaxPlayerCnt()
	if ChairId >= UserCnt {
		return 0
	}

	return 0
}

//叫分(倍数)
func (room *Entry_base) CallScore(args []interface{}) {
	recvMsg := args[0].(*nn_tb_msg.C2G_TBNN_CallScore)
	u := args[1].(*user.User)

	room.DataMgr.CallScore(u, recvMsg.CallScore)
	return
}

//加注
func (r *Entry_base) AddScore(args []interface{}) {
	recvMsg := args[0].(*nn_tb_msg.C2G_TBNN_AddScore)
	u := args[1].(*user.User)

	r.DataMgr.AddScore(u, recvMsg.Score)
	return
}

// 亮牌
func (r *Entry_base) OpenCard(args []interface{}) {
	recvMsg := args[0].(*nn_tb_msg.C2G_TBNN_OpenCard)
	u := args[1].(*user.User)

	r.DataMgr.OpenCard(u, recvMsg.CardType, recvMsg.CardData)
	return
}

// 十三水摊牌
func (r *Entry_base) ShowSSsCard(args []interface{}) {
	//recvMsg := args[0].(*pk_sss_msg.C2G_SSS_Open_Card)
	//u := args[1].(*user.User)

	//r.DataMgr.ShowSSSCard(u, recvMsg.Dragon, recvMsg.SpecialType, recvMsg.SpecialData, recvMsg.FrontCard, recvMsg.MidCard, recvMsg.BackCard)
	return
}

func (r *Entry_base) getRoomUser(uid int64) *user.User {
	u, _ := r.UserMgr.GetUserByUid(uid)
	return u
}
