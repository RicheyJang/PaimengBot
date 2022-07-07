package inspection

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	log "github.com/sirupsen/logrus"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/RicheyJang/PaimengBot/utils/client"
)

func getLatestVersion() (string, error) {
	cli := client.NewHttpClient(&client.HttpOptions{TryTime: 3, Timeout: 15 * time.Second})
	rsp, err := cli.GetGJson("https://api.github.com/repos/RicheyJang/PaimengBot/releases/latest")
	if err != nil {
		return unknownVersion, err
	}
	return rsp.Get("tag_name").String(), nil
}

func downloadAndReplace(version string) error {
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
	filename := downloadURL[index+1:]
	// 下载文件
	downloadDir := filepath.Join(consts.TempRootDir, "update")
	if _, err := utils.MakeDir(downloadDir); err != nil {
		return err
	}
	// defer utils.RemovePath(downloadDir) // TODO 下载更新完成后删除临时文件夹
	downloadPath := filepath.Join(downloadDir, filename)
	if err := client.DownloadToFile(downloadPath, downloadURL, 2); err != nil {
		return err
	}
	// 解压出可执行文件
	var dest string
	if strings.HasSuffix(filename, ".tar.gz") {
		dest, err = tarGzFile(downloadPath, downloadDir)
	} else {
		dest, err = unzipFile(downloadPath, downloadDir)
	}
	if err != nil {
		return fmt.Errorf("decompress err: %v", err)
	}
	// TODO 替换旧的可执行文件
	log.Infof("解压至%v", dest)
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

// unzipFile 解压缩zip文件到源目录
func unzipFile(zipPath, dest string) (string, error) {
	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer rc.Close()
	// 遍历压缩包
	for _, f := range rc.File {
		if !f.FileInfo().IsDir() && (f.Mode().Perm()&os.FileMode(0111)) == 0111 { // 只需解压出可执行文件
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
		if !h.FileInfo().IsDir() && (h.FileInfo().Mode().Perm()&os.FileMode(0111)) == 0111 {
			dest = filepath.Join(dest, filepath.Base(h.Name))
			return dest, decompressFileTo(tr, dest)
		}
	}
	return "", fmt.Errorf("executable file not found")
}

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
