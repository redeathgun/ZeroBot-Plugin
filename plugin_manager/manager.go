// Package manager 群管
package manager

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	timer "github.com/FloatTech/ZeroBot-Plugin-Timer"
)

const (
	datapath = "data/manager/"
	confile  = datapath + "config.pb"
	hint     = "====群管====\n" +
		"- 禁言@QQ 1分钟\n" +
		"- 解除禁言 @QQ\n" +
		"- 我要自闭 1分钟\n" +
		"- 开启全员禁言[群号(可选)]\n" +
		"- 解除全员禁言[群号(可选)]\n" +
		"- 升为管理@QQ\n" +
		"- 取消管理@QQ\n" +
		"- 修改名片@QQ XXX\n" +
		"- 修改头衔@QQ XXX\n" +
		"- 申请头衔 XXX\n" +
		"- 踢出群聊@QQ\n" +
		"- 退出群聊 1234\n" +
		"- 修改群名XXX" +
		"- 群聊转发 1234 XXX\n" +
		"- 私聊转发 0000 XXX\n" +
		"- 在MM月dd日的hh点mm分时(用http://url)提醒大家XXX\n" +
		"- 在MM月[每周|周几]的hh点mm分时(用http://url)提醒大家XXX\n" +
		"- 取消在MM月dd日的hh点mm分的提醒\n" +
		"- 取消在MM月[每周|周几]的hh点mm分的提醒\n" +
		"- 列出所有提醒\n" +
		"- 翻牌\n" +
		"- 设置欢迎语XXX\n" +
		"- [开启|关闭]入群验证"
)

var (
	config Config
)

func init() { // 插件主体
	loadConfig()
	// 菜单
	zero.OnFullMatch("群管系统", zero.AdminPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(hint)
		})
	// 升为管理
	zero.OnRegex(`^升为管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				true,
			)
			nickname := ctx.GetGroupMemberInfo( // 被升为管理的人的昵称
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text(nickname + " 升为了管理~"))
		})
	// 取消管理
	zero.OnRegex(`^取消管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被取消管理的人的昵称
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text("残念~ " + nickname + " 暂时失去了管理员的资格"))
		})
	// 踢出群聊
	zero.OnRegex(`^踢出群聊.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupKick(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被踢出群聊的人的昵称
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text("残念~ " + nickname + " 被放逐"))
		})
	// 退出群聊
	zero.OnRegex(`^退出群聊.*?(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupLeave(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要退出的群的群号
				true,
			)
		})
	// 开启全体禁言
	zero.OnRegex(`^开启全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				true,
			)
			ctx.SendChain(message.Text("全员自闭开始~"))
		})

	zero.OnRegex(`^开启全员禁言.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要解除的群的群号
				true,
			)
			ctx.SendChain(message.Text("全员自闭开始~"))
		})

	// 解除全员禁言
	zero.OnRegex(`^解除全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				false,
			)
			ctx.SendChain(message.Text("全员自闭结束~"))
		})

	zero.OnRegex(`^解除全员禁言.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要解除的群的群号
				false,
			)
			ctx.SendChain(message.Text("全员自闭结束~"))
		})

	// 禁言
	zero.OnRegex(`^禁言.*?(\d+).*?\s(\d+)(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			duration := strToInt(ctx.State["regex_matched"].([]string)[2])
			switch ctx.State["regex_matched"].([]string)[3] {
			case "分钟":
				//
			case "小时":
				duration *= 60
			case "天":
				duration *= 60 * 24
			default:
				//
			}
			if duration >= 43200 {
				duration = 43199 // qq禁言最大时长为一个月
			}
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要禁言的人的qq
				duration*60, // 要禁言的时间（分钟）
			)
			ctx.SendChain(message.Text("小黑屋收留成功~"))
		})
	// 解除禁言
	zero.OnRegex(`^解除禁言.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要解除禁言的人的qq
				0,
			)
			ctx.SendChain(message.Text("小黑屋释放成功~"))
		})
	// 自闭禁言
	zero.OnRegex(`^我要自闭.*?(\d+)(.*)`, zero.OnlyGroup, zero.OnlyToMe).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			duration := strToInt(ctx.State["regex_matched"].([]string)[1])
			switch ctx.State["regex_matched"].([]string)[2] {
			case "分钟":
				//
			case "小时":
				duration *= 60
			case "天":
				duration *= 60 * 24
			default:
				//
			}
			if duration >= 43200 {
				duration = 43199 // qq禁言最大时长为一个月
			}
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				ctx.Event.UserID,
				duration*60, // 要自闭的时间（分钟）
			)
			ctx.SendChain(randText("那我就不手下留情了~", "那咱就不客气了~"))
		})
	// 修改名片
	zero.OnRegex(`^修改名片.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupCard(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被修改群名片的人
				ctx.State["regex_matched"].([]string)[2],           // 修改成的群名片
			)
			ctx.SendChain(message.Text("嗯！已经修改了"))
		})
	// 修改头衔
	zero.OnRegex(`^修改头衔.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[2],           // 修改成的群头衔
			)
			ctx.SendChain(message.Text("嗯！已经修改了"))
		})
	// 申请头衔
	zero.OnRegex(`^申请头衔(.*)`, zero.OnlyGroup).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				ctx.Event.UserID,                         // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[1], // 修改成的群头衔
			)
			ctx.SendChain(message.Text("嗯！不错的头衔呢~"))
		})
	/*zero.OnRegex(`^改名(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
	Handle(func(ctx *zero.Ctx) {
		ctx.SetGroupSpecialTitle(
			ctx.Event.GroupID,
			ctx.Event.SelfID,                         // 被修改群头衔的人
			ctx.State["regex_matched"].([]string)[1], // 修改成的群头衔
		)
		ctx.SendChain(message.Text("嗯！不错的名字呢~"))
	})*/

	/*zero.OnRegex(`^修改群名.*?(\d+).*?\s(.*)`, zero.AdminPermission).SetBlock(true).SetPriority(40).
	Handle(func(ctx *zero.Ctx) {
		ctx.SetGroupName(
			strToInt(ctx.State["regex_matched"].([]string)[1]), // 被修改群头衔的人
			ctx.State["regex_matched"].([]string)[1],           // 修改成的群名
		)
		ctx.SendChain(message.Text("嗯！已经修改了"))
	})*/
	// 群聊转发
	zero.OnRegex(`^群聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			// 对CQ码进行反转义
			content := ctx.State["regex_matched"].([]string)[2]
			content = strings.ReplaceAll(content, "&#91;", "[")
			content = strings.ReplaceAll(content, "&#93;", "]")
			ctx.SendGroupMessage(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 需要发送的群
				content, // 需要发送的信息
			)
			ctx.SendChain(message.Text("📧 --> " + ctx.State["regex_matched"].([]string)[1]))
		})
	// 私聊转发
	zero.OnRegex(`^私聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			// 对CQ码进行反转义
			content := ctx.State["regex_matched"].([]string)[2]
			content = strings.ReplaceAll(content, "&#91;", "[")
			content = strings.ReplaceAll(content, "&#93;", "]")
			ctx.SendPrivateMessage(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 需要发送的人的qq
				content, // 需要发送的信息
			)
			ctx.SendChain(message.Text("📧 --> " + ctx.State["regex_matched"].([]string)[1]))
		})
	// 定时提醒
	zero.OnRegex(`^在(.{1,2})月(.{1,3}日|每?周.?)的(.{1,3})点(.{1,3})分时(用.+)?提醒大家(.*)`, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID > 0 {
				dateStrs := ctx.State["regex_matched"].([]string)
				ts := timer.GetFilledTimeStamp(dateStrs, false)
				ts.Grpid = uint64(ctx.Event.GroupID)
				if ts.Enable {
					go timer.RegisterTimer(ts, true)
					ctx.Send("记住了~")
				} else {
					ctx.Send("参数非法!")
				}
			}
		})
	// 取消定时
	zero.OnRegex(`^取消在(.{1,2})月(.{1,3}日|每?周.?)的(.{1,3})点(.{1,3})分的提醒`, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID > 0 {
				dateStrs := ctx.State["regex_matched"].([]string)
				ts := timer.GetFilledTimeStamp(dateStrs, true)
				ts.Grpid = uint64(ctx.Event.GroupID)
				ti := timer.GetTimerInfo(ts)
				t, ok := (*timer.Timers)[ti]
				if ok {
					t.Enable = false
					delete(*timer.Timers, ti) // 避免重复取消
					_ = timer.SaveTimers()
					ctx.Send("取消成功~")
				} else {
					ctx.Send("没有这个定时器哦~")
				}
			}
		})
	// 列出本群所有定时
	zero.OnFullMatch("列出所有提醒", zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID > 0 {
				ctx.Send(fmt.Sprint(timer.ListTimers(uint64(ctx.Event.GroupID))))
			}
		})
	// 随机点名
	zero.OnFullMatchGroup([]string{"翻牌"}, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID > 0 {
				list := ctx.GetGroupMemberList(ctx.Event.GroupID)
				rand.Seed(time.Now().UnixNano())
				randIndex := fmt.Sprint(rand.Intn(int(list.Get("#").Int())))
				randCard := list.Get(randIndex + ".card").String()
				if randCard == "" {
					randCard = list.Get(randIndex + ".nickname").String()
				}
				ctx.Send("@" + randCard + "，就是你啦!")
			}
		})
	// 入群欢迎
	zero.OnNotice().SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_increase" {
				word, ok := config.Welcome[uint64(ctx.Event.GroupID)]
				if ok {
					ctx.Send(word)
				} else {
					ctx.SendChain(randText("欢迎~", "欢迎入群", "welcome"))
				}
				enable, ok1 := config.Checkin[uint64(ctx.Event.GroupID)]
				if ok1 && enable {
					uid := ctx.Event.UserID
					a := rand.Intn(100)
					b := rand.Intn(100)
					r := a + b
					ctx.SendChain(message.At(uid), message.Text(fmt.Sprintf("考你一道题：%d+%d=?\n如果60秒之内答不上来，%s就要把你踢出去了哦~", a, b, zero.BotConfig.NickName[0])))
					// 匹配发送者进行验证
					rule := func(ctx *zero.Ctx) bool {
						for _, elem := range ctx.Event.Message {
							if elem.Type == "text" {
								text := strings.ReplaceAll(elem.Data["text"], " ", "")
								ans, err := strconv.Atoi(text)
								if err == nil {
									if ans != r {
										ctx.SendChain(randText("答案不对哦，再想想吧~", "再想想吧"))
										return false
									}
									return true
								}
							}
						}
						return false
					}
					next := zero.NewFutureEvent("message", 999, false, zero.CheckUser(ctx.Event.UserID), rule)
					recv, cancel := next.Repeat()
					select {
					case <-time.After(time.Minute):
						ctx.Send("拜拜啦~")
						ctx.SetGroupKick(ctx.Event.GroupID, uid, false)
						cancel()
					case <-recv:
						cancel()
						ctx.SendChain(randText("答对啦~", "回答正确！", "恭喜答对~"))
					}
				}
			}
		})
	// 退群提醒
	zero.OnNotice().SetBlock(false).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_decrease" {
				ctx.SendChain(randText("有人跑路了~", "有人走掉了呢~"))
			}
		})
	// 设置欢迎语
	zero.OnRegex(`^设置欢迎语([\s\S]*)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			config.Welcome[uint64(ctx.Event.GroupID)] = ctx.State["regex_matched"].([]string)[1]
			if saveConfig() == nil {
				ctx.Send("记住啦!")
			} else {
				ctx.Send("出错啦!")
			}
		})
	// 入群验证开关
	zero.OnRegex(`^(.*)入群验证$`, zero.OnlyGroup, zero.OwnerPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["regex_matched"].([]string)[1]
			switch option {
			case "开启":
				config.Checkin[uint64(ctx.Event.GroupID)] = true
			case "关闭":
				config.Checkin[uint64(ctx.Event.GroupID)] = false
			default:
				return
			}
			if saveConfig() == nil {
				ctx.Send("已" + option)
			} else {
				ctx.Send("出错啦!")
			}
		})
	// 运行 CQ 码
	zero.OnRegex(`^run(.*)$`, zero.SuperUserPermission).SetBlock(true).SetPriority(0).
		Handle(func(ctx *zero.Ctx) {
			var cmd = ctx.State["regex_matched"].([]string)[1]
			cmd = strings.ReplaceAll(cmd, "&#91;", "[")
			cmd = strings.ReplaceAll(cmd, "&#93;", "]")
			ctx.Send(cmd)
		})
}

func strToInt(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}

func randText(text ...string) message.MessageSegment {
	length := len(text)
	return message.Text(text[rand.Intn(length)])
}

// loadConfig 加载设置，没有则手动初始化
func loadConfig() {
	mkdirerr := os.MkdirAll(datapath, 0755)
	if mkdirerr == nil {
		if _, err := os.Stat(confile); err == nil || os.IsExist(err) {
			f, err := os.Open(confile)
			if err == nil {
				data, err1 := io.ReadAll(f)
				if err1 == nil {
					if len(data) > 0 {
						if config.Unmarshal(data) == nil {
							return
						}
					}
				}
			}
		}
		config.Checkin = make(map[uint64]bool)
		config.Welcome = make(map[uint64]string)
	} else {
		panic(mkdirerr)
	}
}

// saveConfig 保存设置，无此文件则新建
func saveConfig() error {
	data, err := config.Marshal()
	if err != nil {
		return err
	} else if _, err := os.Stat(datapath); err == nil || os.IsExist(err) {
		f, err1 := os.OpenFile(confile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err1 != nil {
			return err1
		}
		defer f.Close()
		_, err2 := f.Write(data)
		return err2
	}
	return nil
}
