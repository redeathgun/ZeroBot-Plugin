package lolicon

import (
	"github.com/wdvxdr1123/ZeroBot/message"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {

	zero.OnFullMatch("随机美图", zero.OnlyToMe).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Image("https://iw233.cn/api/Random.php"))
		})
}
