package jokes

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"unsafe"
)

// 笑话列表（来自文件）
var jokesList []string

func init() {
	_, _ = utils.MakeDir(consts.JokesDir)
	LoadJokesFromDir(consts.JokesDir)
}
func LoadJokesFromDir(dir string) {
	jokesList = []string{}
	log.Infof("<jokes> 开始重载笑话列表")
	if !utils.DirExists(dir) {
		return
	}
	path := dir + "/jokes.txt"
	fi, err := os.Stat(path)
	if err != nil && !os.IsExist(err) {
		log.Infof("<jokes> 笑话文件不存在,已新建")
		_, err := os.Create(path)
		if err != nil {
			log.Errorf("<jokes> 笑话文件创建失败,err=" + err.Error())

			return
		}
		return
	}
	if fi.Size() >= 1024*1024*1024 { // > 1GB
		log.Error(fmt.Errorf("<jokes> 文件过大(>1GB)"))
		return
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(fmt.Errorf(err.Error()))
		return
	}
	// 解析
	if len(content) == 0 { // 文件内容为空
		return
	}
	contentStr := *(*string)(unsafe.Pointer(&content))
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		jokesList = append(jokesList, line)
	}
	log.Infof(fmt.Sprintf("<jokes> 共加载了%d条笑话", len(lines)))
}
