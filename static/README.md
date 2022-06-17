## 开发说明

此目录(static)专用于存放静态资源文件，例如背景图片、装饰图片等等。

此目录在编译时会被embed到可执行文件中，可以通过以下方式访问其中存放的文件：

#### 读取文件示例

假如说你在static目录下放置了一个something.txt文件，想要读取并输出至log
```go
import (
	"github.com/RicheyJang/PaimengBot/manager"
	log "github.com/sirupsen/logrus"
)

func readAndLog() {
    v, err := manager.ReadStaticFile("README.md") // 读取该文件的全部内容
    if err != nil {
        log.Errorf("read README.md error: %v", err)
        return
    }
    log.Infof("README.md content: %s", string(v))
}
```

#### 背景图片示例

假如说你在static/img目录下放置了一个background.png图片，想把它作为ImageCtx的背景图片：
```go
import (
    "image"

    "github.com/RicheyJang/PaimengBot/manager"
    "github.com/RicheyJang/PaimengBot/utils/images"
    log "github.com/sirupsen/logrus"
    zero "github.com/wdvxdr1123/ZeroBot"
)

func someHandler(ctx *zero.Ctx) {
    // 读取图片
    file, err := manager.GetStaticFile("img/background.png") // 获取该文件的fs.File
    if err != nil {
        log.Errorf("get background.png error: %v", err)
        return
    }
    im, _, err := image.Decode(file) // 解码为image.Image
    if err != nil {
        log.Errorf("Decode background.png error: %v", err)
        return
    }
    // 初始化ImageCtx
    img := images.NewImageCtxWithBG(im.Bounds().Dx(), im.Bounds().Dy(), im, 1) 

    // ...对img的其它操作，参见utils/images包
    
    // 发送图片
    msg, err := img.GenMessageAuto()
    if err != nil {
        log.Errorf("GenMessageAuto error: %v", err)
        return
    }
    ctx.Send(msg)
}
```
