package chat

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// GetDialogueByFilesRandom 随机获取一条答句（来自文件）消息
func GetDialogueByFilesRandom(groupID int64, question string) message.Message {
	answers := group2Dialogues.Load(groupID, question)
	if len(answers) > 0 { // 随机选择一个答案
		randIndex := rand.Intn(len(answers))
		return preprocessFileAnswer(answers[randIndex])
	}
	return nil
}

// 将文件中的答句解析为消息
func preprocessFileAnswer(answer string) message.Message {
	answer = strings.ReplaceAll(answer, "\\n", "\n") // 支持换行符
	return message.ParseMessageFromString(answer)    // 支持CQ码
}

func init() {
	_, _ = utils.MakeDir(consts.DIYDialogueDir)
	LoadDialoguesFromDir(consts.DIYDialogueDir)

	// 监听文件夹变化
	watchDialogueFileChange()
}

// LoadDialoguesFromDir 从文件夹中读取问答集
func LoadDialoguesFromDir(dir string) {
	group2Dialogues.Clear()
	log.Infof("开始重载文件问答集，原文件问答集已清空")
	if !utils.DirExists(dir) {
		return
	}
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		// 群号列表
		ids := []int64{0} // 默认全局生效
		dotName := strings.SplitN(d.Name(), ".", 2)
		if len(dotName) > 0 {
			tmpIDs := make([]int64, 0)
			idsStr := utils.MergeStringSlices(strings.Split(dotName[0], ","))
			for _, str := range idsStr {
				id, ierr := strconv.ParseInt(str, 10, 64)
				if ierr != nil {
					continue
				}
				tmpIDs = append(tmpIDs, id)
			}
			if len(tmpIDs) > 0 { // 在文件名中指定了群号
				ids = tmpIDs
			}
		}
		// 解析
		mp, err := ParseDialoguesFile(path)
		if err != nil {
			log.Warnf("解析问答集文件[%v]失败：%v", path, err)
			return nil
		}
		for _, id := range ids {
			group2Dialogues.Merge(id, mp)
		}
		log.Infof("成功通过%v文件为群%v载入%d条问答", path, ids, len(mp))
		return nil
	})
}

type dialoguesMap struct {
	mutex sync.RWMutex
	mp    map[int64]map[string][]string // 群号 -> 问题 -> 答句列表
}

// 群对话映射（来自文件）
var group2Dialogues dialoguesMap

// Load 读出一个答案集
func (dm *dialoguesMap) Load(id int64, question string) []string {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	if dm.mp == nil { // 映射为空
		return nil
	}
	d, ok := dm.mp[id] // 获取指定群对话映射
	if !ok || d == nil {
		return nil
	}
	ans, ok := d[question] // 获取答句
	if !ok {
		return nil
	}
	return ans
}

// Clear 清空
func (dm *dialoguesMap) Clear() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	for k := range dm.mp {
		delete(dm.mp, k)
	}
}

// Merge 将新的问答集合并入
func (dm *dialoguesMap) Merge(id int64, mp map[string][]string) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	if dm.mp == nil {
		dm.mp = make(map[int64]map[string][]string)
	}
	if d, ok := dm.mp[id]; !ok || d == nil { // 直接记录
		dm.mp[id] = mp
		return
	}
	// 与原问答映射合并
	dm.mp[id] = mergeMaps(dm.mp[id], mp)
}

// merge two map
func mergeMaps(mps ...map[string][]string) map[string][]string {
	mp := make(map[string][]string)
	for _, m := range mps {
		for k, v := range m {
			if _, ok := mp[k]; !ok {
				mp[k] = v
			} else {
				mp[k] = utils.MergeStringSlices(mp[k], v)
			}
		}
	}
	return mp
}

// ParseDialoguesFile 解析问答集文件
func ParseDialoguesFile(filename string) (map[string][]string, error) {
	fi, err := os.Stat(filename)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	if fi.Size() >= 1024*1024*1024 { // > 1GB
		return nil, fmt.Errorf("文件过大(>1GB)")
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// 解析
	if len(content) == 0 { // 文件内容为空
		return make(map[string][]string), nil
	}
	if content[0] == '{' {
		return parseDialoguesJSONFile(content)
	} else if content[0] == 'Q' || content[0] == 'q' {
		return parseDialoguesTXTFile(content)
	}
	return nil, fmt.Errorf("文件格式错误或暂不支持")
}

func parseDialoguesTXTFile(content []byte) (map[string][]string, error) {
	contentStr := *(*string)(unsafe.Pointer(&content)) // 强制转换
	lines := strings.Split(contentStr, "\n")
	result := make(map[string][]string)
	var currentQ string
	// 遍历每一行
	for index, line := range lines {
		if preLen := getPrefixLen(line, "Q:", "q:", "Q：", "q："); preLen > 0 {
			currentQ = strings.TrimSpace(line[preLen:])
			if len(currentQ) == 0 {
				return nil, fmt.Errorf("第%v行: 问句长度为0", index)
			}
		} else if preLen := getPrefixLen(line, "A:", "a:", "A：", "a："); preLen > 0 {
			if len(currentQ) == 0 {
				return nil, fmt.Errorf("第%v行: 没有所对应的问句", index)
			}
			ans := strings.TrimSpace(line[preLen:])
			if len(ans) == 0 {
				return nil, fmt.Errorf("第%v行: 答句长度为0", index)
			}
			result[currentQ] = append(result[currentQ], ans)
		}
	}
	return result, nil
}

// get prefixes length
func getPrefixLen(str string, prefixes ...string) int {
	for _, prefix := range prefixes {
		if strings.HasPrefix(str, prefix) {
			return len(prefix)
		}
	}
	return 0
}

func parseDialoguesJSONFile(content []byte) (map[string][]string, error) {
	result := make(map[string][]string)
	err := json.Unmarshal(content, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func watchDialogueFileChange() {
	initWG := sync.WaitGroup{}
	initWG.Add(1)
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Error(err)
			return
		}
		defer watcher.Close()

		eventsWG := sync.WaitGroup{}
		eventsWG.Add(1)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Error("Load Dialogue File Goroutine Panic: ", err)
				}
			}()
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok { // 'Events' channel is closed
						eventsWG.Done()
						return
					}

					const opMask = fsnotify.Write | fsnotify.Create | fsnotify.Remove | fsnotify.Rename
					if event.Op&opMask != 0 {
						LoadDialoguesFromDir(consts.DIYDialogueDir)
					}

				case err, ok := <-watcher.Errors:
					if ok { // 'Errors' channel is not closed
						log.Warnf("chat dialogue files watcher error: %v\n", err)
					}
					continue
				}
			}
		}()
		addErr := watcher.Add(consts.DIYDialogueDir)
		if addErr != nil {
			log.Warnf("Watch dir=%v err: %v", consts.DIYDialogueDir, addErr)
		}
		initWG.Done()   // done initializing the watch in this go routine, so the parent routine can move on...
		eventsWG.Wait() // now, wait for event loop to end in this go-routine...
	}()
	initWG.Wait() // make sure that the go routine above fully ended before returning
}
