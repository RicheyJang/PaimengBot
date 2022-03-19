package note

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name:  "定时提醒",
	Usage: ``,
}
var proxy *manager.PluginProxy
var errNoMatched = fmt.Errorf("no regex matched")
var errAlreadyPassed = fmt.Errorf("already passed")

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
}

func noteHandler(ctx *zero.Ctx) {}

func cancelNoteHandler(ctx *zero.Ctx) {}

func listNoteHandler(ctx *zero.Ctx) {}
