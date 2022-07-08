package inspection

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// 获取最新版本号
func getLatestVersion() (string, error) {
	cli := client.NewHttpClient(&client.HttpOptions{TryTime: 3, Timeout: 15 * time.Second})
	rsp, err := cli.GetGJson("https://api.github.com/repos/RicheyJang/PaimengBot/releases/latest")
	if err != nil {
		return unknownVersion, err
	}
	return rsp.Get("tag_name").String(), nil
}

// 下载并将最新的可执行文件放在oldFilepath
func downloadAndReplace(version string, destFilepath string) error {
	// 获取更新压缩包URL
	cli := client.NewHttpClient(&client.HttpOptions{TryTime: 3, Timeout: 15 * time.Second})
	rsp, err := cli.GetGJson("https://api.github.com/repos/RicheyJang/PaimengBot/releases/tags/" + version)
	if err != nil {
		return err
	}
	var downloadAsset gjson.Result
	for _, asset := range rsp.Get("assets").Array() {
		name := strings.ToLower(asset.Get("name").String())
		if checkOS(name) && checkArch(name) {
			downloadAsset = asset
			break
		}
	}
	if !downloadAsset.Exists() {
		return fmt.Errorf("no suitable asset found")
	}
	downloadURL := downloadAsset.Get("browser_download_url").String()
	// 验证URL
	if len(downloadURL) == 0 {
		return fmt.Errorf("downloadURL is empty")
	}
	index := strings.LastIndex(downloadURL, "/")
	if index == -1 {
		return fmt.Errorf("downloadURL is invalid")
	}
	if res, err := cli.Head(downloadURL); err != nil || res.StatusCode != http.StatusOK { // 测试链接可用性
		return fmt.Errorf("HEAD download URL(%v) failed, code=%v, err: %v", downloadURL, res.StatusCode, err)
	}
	filename := downloadURL[index+1:]
	// 下载压缩包
	downloadDir := filepath.Join(consts.TempRootDir, "update")
	if _, err := utils.MakeDir(downloadDir); err != nil {
		return err
	}
	defer utils.RemovePath(downloadDir) // 下载更新完成后删除临时文件夹
	downloadPath := filepath.Join(downloadDir, filename)
	timeoutStr := proxy.GetConfigString("timeout") // 获取超时时间
	timeout, _ := time.ParseDuration(timeoutStr)
	if timeout < 5*time.Second {
		timeout = 5 * time.Second
	}
	cli = client.NewHttpClient(&client.HttpOptions{Timeout: timeout, TryTime: 2})
	if err := cli.DownloadToFile(downloadPath, downloadURL); err != nil {
		return err
	}
	// 解压出可执行文件
	var newBinaryPath string
	if strings.HasSuffix(filename, ".tar.gz") {
		newBinaryPath, err = tarGzFile(downloadPath, downloadDir)
	} else {
		newBinaryPath, err = unzipFile(downloadPath, downloadDir)
	}
	if err != nil {
		return fmt.Errorf("decompress err: %v", err)
	}
	log.Infof("新版本可执行文件已解压至%v，即将复制到%v", newBinaryPath, destFilepath)
	// 将可执行文件复制到指定路径
	oldFile, err := os.OpenFile(destFilepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer oldFile.Close()
	newFile, err := os.Open(newBinaryPath)
	if err != nil {
		return err
	}
	defer newFile.Close()
	if _, err = io.Copy(oldFile, newFile); err != nil {
		return err
	}
	return nil
}

// 检查名称name与当前操作系统是否对应
func checkOS(name string) bool {
	if strings.Contains(name, "linux") && runtime.GOOS == "linux" {
		return true
	}
	if strings.Contains(name, "windows") && runtime.GOOS == "windows" {
		return true
	}
	if strings.Contains(name, "macos") && runtime.GOOS == "darwin" {
		return true
	}
	return false
}

// 检查名称name与当前CPU架构是否对应
func checkArch(name string) bool {
	if strings.Contains(name, "x86_64") && runtime.GOARCH == "amd64" {
		return true
	}
	if strings.Contains(name, "arm64") && runtime.GOARCH == "arm64" {
		return true
	}
	if strings.Contains(name, "x86_32") && runtime.GOARCH == "386" {
		return true
	}
	return false
}

// 解压缩zip中的第一个可执行文件到dest目录
func unzipFile(zipPath, dest string) (string, error) {
	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer rc.Close()
	// 遍历压缩包
	for _, f := range rc.File {
		if isExecutable(f.FileInfo()) { // 只需解压出可执行文件
			fr, err := f.Open()
			if err != nil {
				return "", err
			}
			dest = filepath.Join(dest, filepath.Base(f.Name))
			return dest, decompressFileTo(fr, dest)
		}
	}
	return "", fmt.Errorf("executable file not found")
}

// 解压缩tar.gz中的第一个可执行文件到dest目录
func tarGzFile(zipPath, dest string) (string, error) {
	file, err := os.Open(zipPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// gzip read
	gr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gr.Close()
	// tar read
	tr := tar.NewReader(gr)
	// 读取文件
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if isExecutable(h.FileInfo()) {
			dest = filepath.Join(dest, filepath.Base(h.Name))
			return dest, decompressFileTo(tr, dest)
		}
	}
	return "", fmt.Errorf("executable file not found")
}

// 从fr读取生成destPath文件
func decompressFileTo(fr io.Reader, destPath string) error {
	fw, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer fw.Close()
	_, err = io.Copy(fw, fr)
	if err != nil {
		return err
	}
	return nil
}

// 是否为可执行文件
func isExecutable(file fs.FileInfo) bool {
	if runtime.GOOS == "windows" {
		return !file.IsDir() && strings.HasSuffix(file.Name(), ".exe")
	}
	return !file.IsDir() && (file.Mode().Perm()&os.FileMode(0111)) == 0111
}
