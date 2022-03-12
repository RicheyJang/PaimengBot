package sign_util

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_util/constant"
	"io"
	"io/ioutil"

	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func GetMD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func GetDs() string {
	currentTime := time.Now().Unix()
	stringRom := GetRandString(6, currentTime)
	stringAdd := fmt.Sprintf("salt=%s&t=%d&r=%s", constant.Salt, currentTime, stringRom)
	stringMd5 := GetMD5(stringAdd)
	return fmt.Sprintf("%d,%s,%s", currentTime, stringRom, stringMd5)
}

func GetRandString(len int, seed int64) string {
	bytes := make([]byte, len)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < len; i++ {
		b := r.Intn(36)
		if b > 9 {
			b += 39
		}
		b += 48
		bytes[i] = byte(b)
	}
	return string(bytes)
}

//ReadFile 读取整个文件
func ReadFile(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("unable to read file." + err.Error())
	}
	return b, nil
}

//ReadFileAllLine 按行读取文件
func ReadFileAllLine(path string, handle func(string)) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("unable to read file." + err.Error())
	}
	defer f.Close()

	bufReader := bufio.NewReader(f)

	line, is, err := bufReader.ReadLine()
	for ; !is && err == nil; line, is, err = bufReader.ReadLine() {
		s := string(line)
		handle(s)
	}
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func StructToJSON(v interface{}) ([]byte, error) {
	jsonByte, err := json.Marshal(v)
	if err != nil {
		return nil, errors.New("unable convert struct to json." + err.Error())
	}
	return jsonByte, nil
}

//CheckFileIsExist 判断文件是否存在，存在返回true，不存在返回false
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

//GetCurrentDir 获取当前运行目录
func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

//GetCurrentDir2 获取当前运行目录
func GetCurrentDir2() string {
	exePath, _ := os.Executable()

	dir, _ := filepath.EvalSymlinks(filepath.Dir(exePath))

	tmpDir := os.Getenv("TEMP")
	if tmpDir == "" {
		tmpDir = os.Getenv("TMP")
	}
	link, _ := filepath.EvalSymlinks(tmpDir)

	if strings.Contains(dir, link) {
		var abPath string
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			abPath = path.Dir(filename)
		}
		return abPath
	}
	return dir
}

func GetDayTimestamp(t time.Time) int64 {
	timeStr := t.Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t2.AddDate(0, 0, 1).Unix()
}

func GetCurrentDayTimestamp() int64 {
	return GetDayTimestamp(time.Now())
}
