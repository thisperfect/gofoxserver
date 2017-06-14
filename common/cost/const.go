package cost

import (
	"fmt"
	"mj/common/msg"
)

//login error code 0  ~ 100
const (
	NotFoudAccout        = 1 //没找到账号
	ParamError           = 2 //参数错误
	AlreadyExistsAccount = 3 //账号已经存在
	InsertAccountError   = 4 //服务器内部错误
	LoadUserInfoError    = 5 //玩家数据加载失败
	CreateUserError      = 6 // 创建玩家失败
	ErrUserDoubleLogin   = 7 //重复登录
	ErrPasswd            = 8 //密码错误
)

//房间错误码 100 ~ 200
const (
	RoomFull           = 101 //房间满了，不能再创建
	NotFoudGameType    = 102 //玩家不存在
	CreateParamError   = 103 //参数错误
	NoFoudTemplate     = 104 //配置没找到
	ConfigError        = 105 //配置错误
	NotEnoughFee       = 106 //代币不足
	RandRoomIdError    = 107 //生成房间id失败
	MaxSoucrce         = 108 // 低分太高
	ChairHasUser       = 109 //位置有玩家， 不能坐下
	GameIsStart        = 110 //游戏已经开始， 不能加入
	ErrNotOwner        = 111 // 不是房主 没权限操作
	ErrNoSitdowm       = 112 //请先坐下在操作
	ErrGameIsStart     = 113 //游戏已开始，不能离开房间
	ErrCreateRoomFaild = 114 //创建聊天室失败
	NotOwner           = 115 //不是房主
)

//红中麻将错误码
const (
	NotValidCard     = 201 //无效的牌
	ErrUserNotInRoom = 202 //玩家不在房间
	ErrNotFoudCard   = 203 //没找到牌
	ErrGameNotStart  = 204 //游戏没开始
	ErrNotSelfOut    = 205 //不是自己出牌
)

///////// 无效的数字
const (
	//无效数值
	INVALID_BYTE   = 0xFF       //无效数值
	INVALID_WORD   = 0xFFFF     //无效数值
	INVALID_DWORD  = 0xFFFFFFFF //无效数值
	INVALID_CHAIR  = 0xFFFF     //无效椅子
	INVALID_TABLE  = 0xFFFF     //无效桌子
	INVALID_SERVER = 0xFFFF     //无效房间
	INVALID_KIND   = 0xFFFF     //无效游戏
)

///////////////游戏模式.
const (
	GAME_GENRE_GOLD     = 0x0001 //金币类型
	GAME_GENRE_SCORE    = 0x0002 //点值类型
	GAME_GENRE_MATCH    = 0x0004 //比赛类型
	GAME_GENRE_EDUCATE  = 0x0008 //训练类型
	GAME_GENRE_PERSONAL = 0x0010 //约战类型
)

/// 通用状态
const (
	//用户状态
	US_NULL    = 0x00 //没有状态
	US_FREE    = 0x01 //站立状态
	US_SIT     = 0x02 //坐下状态
	US_READY   = 0x03 //同意状态
	US_LOOKON  = 0x04 //旁观状态
	US_PLAYING = 0x05 //游戏状态
	US_OFFLINE = 0x06 //断线状态
)

const (
	//房间状态
	RoomStatusReady    = 0
	RoomStatusStarting = 1
	RoomStatusEnd      = 2
)

const (
	//结束原因
	GER_NORMAL        = 0x00 //常规结束
	GER_DISMISS       = 0x01 //游戏解散
	GER_USER_LEAVE    = 0x02 //用户离开
	GER_NETWORK_ERROR = 0x03 //网络错误
)

//积分修改类型
///
const (
	HZMJ_CHANGE_SOURCE = 1
)

//////////////////////////////////////////////
//标识前缀
const (
	HallPrefix = "HallSvr" //房间服
	GamePrefix = "GameSvr"
)

func LOBYTE(w int) int {
	return w & 0xFF
}
func HIBYTE(w int) int {
	return w & 0xFF00
}

/////
func RenderErrorMessage(code int, Desc ...string) *msg.ShowErrCode {
	var des string
	if len(Desc) < 1 {
		des = fmt.Sprintf("请求错误, 错误码: %d", code)
	} else {
		des = fmt.Sprintf(Desc[0]+"请求错误, 错误码: %d", code)
	}
	return &msg.ShowErrCode{
		ErrorCode:      code,
		DescribeString: des,
	}
}

func GetGameSvrName(sververId int) string {
	return fmt.Sprintf(GamePrefix+"_%d", sververId)
}
func GetHallSvrName(sververId int) string {
	return fmt.Sprintf(HallPrefix+"_%d", sververId)
}
