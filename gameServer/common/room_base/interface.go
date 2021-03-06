package room_base

import (
	"mj/common/msg"
	"mj/gameServer/user"
	"time"

	"github.com/lovelly/leaf/chanrpc"
	"github.com/lovelly/leaf/module"
	"github.com/lovelly/leaf/timer"
)

type Module interface {
	GetChanRPC() *chanrpc.Server
	GetClientCount() int
	GetTableCount() int
	OnDestroy()
	OnInit()
	Run(chan bool)
	CreateRoom(args ...interface{}) bool
}

type TimerManager interface {
	StartCreatorTimer(cb func())
	StopCreatorTimer()
	StartKickoutTimer(uid int64, cb func())
	StopOfflineTimer(uid int64)
	StartReplytIimer(uid int64, cb func())
	StopReplytIimer(uid int64)

	GetTimeLimit() int
	GetPlayCount() int
	AddPlayCount()
	ResetPlayCount()
	GetMaxPlayCnt() int
	GetRoomPlayCnt() int
	AddMaxPlayCnt(int)
	GetCreatrTime() int64
	GetTimeOutCard() int
	GetTimeOperateCard() int
}

type UserManager interface {
	Sit(*user.User, int, int) int
	Standup(*user.User) bool
	ForEachUser(fn func(*user.User))
	GetLeaveInfo(int64) *msg.LeaveReq
	LeaveRoom(*user.User, int) bool
	SetUsetStatus(*user.User, int)
	ReLogin(*user.User, int)
	IsAllReady() bool
	RoomDissume(cbReason int)
	SendUserInfoToSelf(*user.User)
	SendMsgAll(data interface{})
	SendMsgAllNoSelf(selfid int64, data interface{})
	WriteTableScore(source []*msg.TagScoreInfo, usercnt, Type int)
	SendDataToHallUser(chiairID int, data interface{})
	SendMsgToHallServerAll(data interface{})
	ReplyLeave(*user.User, bool, int64, int) int
	DelLeavePly(uid int64)
	AddLeavePly(uid int64)
	GetBeginPlayer() int
	ResetBeginPlayer()
	CheckRoomReturnMoney(roomStatus, CreatorNodeId, roomId int, creatorId int64)

	GetCurPlayerCnt() int
	GetPayType() int
	IsPublic() bool
	GetMaxPlayerCnt() int
	GetUserInfoByChairId(int) interface{}
	GetUserByChairId(int) *user.User
	GetUserByUid(userId int64) (*user.User, int)
	SetUsetTrustee(chairId int, isTruste bool)
	IsTrustee(chairId int) bool
	GetTrustees() []bool
}

type BaseManager interface {
	Destroy(int)
	RoomRun(int)
	GetSkeleton() *module.Skeleton
	AfterFunc(d time.Duration, cb func()) *timer.Timer
	GetChanRPC() *chanrpc.Server
}

type BData interface {
	GetRoomId() int
	InitRoomOne()
	ResetGameAfterRenewal()
	BeforeStartGame(UserCnt int) //开始前的处理
	StartGameing()               //游戏开始种的处理
	AfterStartGame()             //开始游戏的善后处理
	GetUserScore(chairid int) int
	SendPersonalTableTip(*user.User) //发送没开始前的场景信息
	SendStatusReady(u *user.User)    //发送准备
	SendStatusPlay(u *user.User)     //发送开始后的处理
	GetCreator() int64
	GetCreatorNodeId() int
	ResetRoomCreator(uid int64, nodeid int)
	NormalEnd(Reason int)    //正常结束
	DismissEnd(Reason int)   //解散结束
	TrusteeEnd(cbReason int) //房间托管结束
}
