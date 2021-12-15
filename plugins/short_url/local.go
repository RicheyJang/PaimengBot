package short_url

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/client"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// FindShortURLOrgByLocal 无依赖地将短链接解析成原始链接
func FindShortURLOrgByLocal(short string) (long string) {
	var lastReq *http.Request
	redirect := func(req *http.Request, via []*http.Request) error {
		if req != nil {
			lastReq = req
			log.Infof("new req : method=%v,url=%v,header=%v\n", req.Method, req.URL, req.Header)
		}
		return nil
	}
	// 初始化
	short = completeURL(short)
	jar, _ := cookiejar.New(nil)
	c := http.Client{
		CheckRedirect: redirect,
		Jar:           jar,
		Timeout:       5 * time.Second,
	}
	u, err := url.Parse(short)
	if err != nil {
		log.Errorf("Parse URL for %v err: %v", short, err)
		return
	}
	defer func() {
		// 最后一次请求不为空，且最后一次请求的URL与short不同，则取最后一次请求的URL为准
		if len(long) == 0 && !(lastReq == nil || lastReq.URL.String() == short || lastReq.URL.String() == u.String()) {
			long = lastReq.URL.String()
		}
		// 解析失败，尝试https，出现于新浪短链接
		if len(long) == 0 && strings.HasPrefix(short, "http://") {
			long = FindShortURLOrgByLocal(strings.Replace(short, "http://", "https://", 1))
		}
	}()
	// 模拟请求
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Warnf("NewRequest for %v err: %v", u.String(), err)
		return
	}
	req.Header.Set("Host", u.Host)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0")
	rsp, err := c.Do(req)
	if err != nil {
		log.Warnf("GET %v err: %v", u.String(), err)
		return
	}
	_ = rsp.Body.Close()
	return
}

// 为33h.co单独解析，因为它会有一层HTML转跳
func findShortURLOrgFor33hCo(short string) string {
	u, err := url.Parse(completeURL(short))
	if err != nil || u.Host != "33h.co" {
		return ""
	}
	// 为33h.co产生的短链接，请求
	c := client.NewHttpClient(&client.HttpOptions{TryTime: 2})
	res, err := c.Get(u.String())
	if err != nil {
		log.Warnf("GET for %v err: %v", u.String(), err)
		return ""
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Warnf("GET for %v code != 2XX : %v", u.String(), res.Status)
		return ""
	}
	// 解析
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Warnf("GET for %v NewDocument err: %v", u.String(), err)
		return ""
	}
	return doc.Find("p.url>a").First().Text()
}

func completeURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}
	return url
}
