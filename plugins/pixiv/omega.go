package pixiv

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// 从Omega图库随机获取图片，返回图片URL切片
func getPicturesFromOmega(tags []string, num int, isR18 bool) (pics []PictureInfo) {
	nsfw := 0
	if isR18 {
		nsfw = 2
	}
	// 构建查询
	var models []OmegaPixivIllusts
	tx := proxy.GetDB()
	// nsfw Flag
	if proxy.GetConfigBool("omega.setu") && nsfw == 0 {
		tx = tx.Where("nsfw_tag IN ?", []int{0, 1})
	} else {
		tx = tx.Where("nsfw_tag = ?", nsfw)
	}
	// 筛选标签 AND
	for _, tag := range tags {
		tx = tx.Where("tags LIKE ?", "%"+tag+"%")
	}
	// 查询
	err := tx.Scopes(proxy.SQLRandomOrder).Limit(num).Find(&models).Error
	if err != nil {
		log.Warnf("Query OmegaPixivIllusts err: %v", err)
	}
	// 构成图片信息
	for _, m := range models {
		if strings.Contains(m.URL, "www.pixiv.net") { // 无法下载
			m.URL = ""
		}
		pics = append(pics, PictureInfo{
			Title:  m.Title,
			URL:    m.URL,
			Tags:   strings.Split(m.Tags, ","),
			PID:    m.PID,
			P:      0,
			Author: m.Uname,
			UID:    m.UID,
		})
	}
	// 刷新图片URL
	if len(models) > 0 {
		go flushOmegaURL(models)
	}
	return
}

func flushOmegaURL(models []OmegaPixivIllusts) {
	time.Sleep(10 * time.Second) // 防止抢占网络带宽
	var count int64
	for _, m := range models {
		if len(m.URL) > 0 && !strings.Contains(m.URL, "www.pixiv.net") {
			continue
		}
		p := PictureInfo{
			PID: m.PID,
			P:   0,
		}
		_ = p.GetURLByPID()
		count += proxy.GetDB().Model(&OmegaPixivIllusts{}).Where("id = ?", m.ID).Update("url", p.URL).RowsAffected
	}
	log.Infof("成功更新Omega URL：%v条记录", count)
}
