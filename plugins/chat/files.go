package chat

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/RicheyJang/PaimengBot/basic/nickname"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// AddDialogueCollection 以 DialoguesCollection 的形式添加问答集，仅在本次运行中生效
func AddDialogueCollection(groupID int64, dc *DialoguesCollection) {
	group2Dialogues.Merge(groupID, dc) // 保存至内存中
}

// GetDialogueByFilesRandom 随机获取一条答句（来自文件）消息
func GetDialogueByFilesRandom(ctx *zero.Ctx, groupID int64, question string) message.Message {
	answers := group2Dialogues.Load(groupID, question)
	if len(answers) > 0 { // 随机选择一个答案
		randIndex := rand.Intn(len(answers))
		return preprocessFileAnswer(ctx, answers[randIndex])
	}
	return nil
}

// 将文件中的答句解析为消息
func preprocessFileAnswer(ctx *zero.Ctx, answer string) message.Message {
	if strings.Contains(answer, "{bot}") {
		answer = strings.ReplaceAll(answer, "{bot}", utils.GetBotNickname())
	}
	if strings.Contains(answer, "{nickname}") {
		answer = strings.ReplaceAll(answer, "{nickname}", nickname.GetNickname(ctx.Event.UserID, "你"))
	}
	if strings.Contains(answer, "{id}") {
		answer = strings.ReplaceAll(answer, "{id}", strconv.FormatInt(ctx.Event.UserID, 10))
	}
	answer = strings.ReplaceAll(answer, "\\n", "\n") // 支持换行符
	return message.ParseMessageFromString(answer)    // 支持CQ码
}

// 在正则匹配成功后，返回答案集前调用：将正则问句的答句内{reg[i]}替换为正则表达式中的第i个组
func preprocessRegAnswer(regD regexpDialogue, question string) (answers []string) {
	matches := regD.reg.FindStringSubmatch(question)
	if len(matches) <= 1 {
		return regD.answers
	}
	answers = make([]string, len(regD.answers))
	copy(answers, regD.answers)
	for i, match := range matches {
		if i == 0 {
			continue
		}
		key := "{reg[" + strconv.Itoa(i) + "]}"
		for k, org := range answers {
			answers[k] = strings.ReplaceAll(org, key, match)
		}
	}
	return
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
		dc, err := ParseDialoguesFile(path)
		if err != nil {
			log.Warnf("解析问答集文件[%v]失败：%v", path, err)
			return nil
		}
		for _, id := range ids {
			AddDialogueCollection(id, dc)
		}
		log.Infof("成功通过%v文件为群%v载入%d条问答", path, ids, dc.Length())
		return nil
	})
}

// 问答集映射结构
type dialoguesMap struct {
	mutex sync.RWMutex
	dcMap map[int64]*DialoguesCollection // 群号 -> 问答集
}

// DialoguesCollection 一个问答集
type DialoguesCollection struct {
	full map[string][]string // 全匹配：问题 -> 答句列表
	regs []regexpDialogue    // 正则匹配：问题正则及其答句列表
}

// 问句为正则格式的单个问答
type regexpDialogue struct {
	reg     *regexp.Regexp
	answers []string
}

// 群对话映射（来自文件）
var group2Dialogues dialoguesMap

func newDialoguesCollection(mp map[string][]string, regs []regexpDialogue) *DialoguesCollection {
	return &DialoguesCollection{
		full: mp,
		regs: regs,
	}
}

// Length 获取问答集的问句个数
func (dc DialoguesCollection) Length() int {
	return len(dc.full) + len(dc.regs)
}

// Load 获取答句列表
func (dc DialoguesCollection) Load(question string) []string {
	// 优先全匹配
	if dc.full != nil {
		if ans, ok := dc.full[question]; ok {
			return ans
		}
	}
	// 随后遍历正则
	for _, regD := range dc.regs {
		if regD.reg != nil && regD.reg.MatchString(question) {
			return preprocessRegAnswer(regD, question)
		}
	}
	return nil
}

// AutoSeparateReg 自动从全匹配map中分离出正则
func (dc *DialoguesCollection) AutoSeparateReg() error {
	for q, ans := range dc.full {
		if len(q) >= 3 && strings.HasPrefix(q, "/") && strings.HasSuffix(q, "/") { // 符合正则问句约定
			expr := q[1 : len(q)-1]
			reg, err := regexp.Compile(expr)
			if err != nil {
				return fmt.Errorf("正则%v不符合规范：%v", q, err)
			}
			// 将该问答从全匹配map切换至正则切片
			dc.regs = append(dc.regs, regexpDialogue{
				reg:     reg,
				answers: ans,
			})
			delete(dc.full, q)
		}
	}
	return nil
}

// Merge 与另一问答集合并
func (dc *DialoguesCollection) Merge(another *DialoguesCollection) {
	if dc == nil || another == nil {
		return
	}
	dc.full = mergeMaps(dc.full, another.full)
	// 合并正则切片
	for _, anotherReg := range another.regs {
		needAppend := true
		for i, thisReg := range dc.regs {
			if thisReg.reg.String() == anotherReg.reg.String() { // 已存在相同正则，合并答案列表
				dc.regs[i].answers = utils.MergeStringSlices(thisReg.answers, anotherReg.answers)
				needAppend = false
				break
			}
		}
		if needAppend { // 不存在相同正则，直接添加
			dc.regs = append(dc.regs, anotherReg)
		}
	}
}

// Load 读出一个答案集
func (dm *dialoguesMap) Load(id int64, question string) []string {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	if dm.dcMap == nil { // 映射为空
		return nil
	}
	d, ok := dm.dcMap[id] // 获取指定群对话映射
	if !ok || d == nil {
		return nil
	}
	return d.Load(question) // 获取答句
}

// Clear 清空
func (dm *dialoguesMap) Clear() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	for k := range dm.dcMap {
		delete(dm.dcMap, k)
	}
}

// Merge 将新的问答集合并入
func (dm *dialoguesMap) Merge(id int64, dc *DialoguesCollection) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	if dm.dcMap == nil {
		dm.dcMap = make(map[int64]*DialoguesCollection)
	}
	if d, ok := dm.dcMap[id]; !ok || d == nil { // 直接记录
		dm.dcMap[id] = dc
		return
	}
	// 与原问答集合并
	dm.dcMap[id].Merge(dc)
}

// 合并多个map[string][]string类型的map
func mergeMaps(mps ...map[string][]string) map[string][]string {
	mp := make(map[string][]string)
	for _, m := range mps {
		if m == nil {
			continue
		}
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
func ParseDialoguesFile(filename string) (dc *DialoguesCollection, err error) {
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
		return newDialoguesCollection(make(map[string][]string), nil), nil
	}
	defer func() {
		if dc != nil && err == nil { // 分离出正则问句
			err = dc.AutoSeparateReg()
		}
	}()
	if content[0] == '{' {
		return parseDialoguesJSONFile(content)
	} else if content[0] == 'Q' || content[0] == 'q' {
		return parseDialoguesTXTFile(content)
	}
	return nil, fmt.Errorf("文件格式错误或暂不支持")
}

// 解析TXT格式问答集
func parseDialoguesTXTFile(content []byte) (*DialoguesCollection, error) {
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
	return newDialoguesCollection(result, nil), nil
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

// 解析JSON格式问答集
func parseDialoguesJSONFile(content []byte) (*DialoguesCollection, error) {
	result := make(map[string][]string)
	err := json.Unmarshal(content, &result)
	if err != nil {
		return nil, err
	}
	// ！JSON格式并不支持正则问句
	return newDialoguesCollection(result, nil), nil
}

// 监听预置问答集文件目录变化，实时更新问答集
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
					if event.Op&opMask != 0 &&
						(strings.HasSuffix(event.Name, ".txt") || strings.HasSuffix(event.Name, ".json")) {
						// 仅限TXT文件和JSON文件
						LoadDialoguesFromDir(consts.DIYDialogueDir) // 更新问答集
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
