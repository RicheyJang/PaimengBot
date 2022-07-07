package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

type HttpClient struct {
	HttpOptions
	client  *http.Client
	header  map[string]string
	cookies []*http.Cookie
}

type HttpOptions struct {
	TryTime int
	Timeout time.Duration
}

// NewHttpClient 创建新Http请求器
func NewHttpClient(option *HttpOptions) *HttpClient {
	if option == nil {
		option = new(HttpOptions)
	}
	if option.TryTime == 0 {
		option.TryTime = 1
	}
	if option.Timeout == 0 {
		option.Timeout = 10 * time.Second
	}
	return &HttpClient{
		HttpOptions: *option,
		client:      &http.Client{Timeout: option.Timeout},
		header:      make(map[string]string),
	}
}

func ParseReader(reader io.Reader) gjson.Result {
	rspBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return gjson.Result{}
	}
	return gjson.Parse(string(rspBody))
}

func (c *HttpClient) AddCookie(cookie ...*http.Cookie) {
	c.cookies = append(c.cookies, cookie...)
}

func (c *HttpClient) SetHeader(key string, val string) {
	c.header[key] = val
}

func (c *HttpClient) SetUserAgent() {
	c.SetHeader("User-Agent",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36")
}

func (c HttpClient) Do(req *http.Request) (*http.Response, error) {
	var res *http.Response
	err := errors.New("TryTime is zero, send no http request")
	if req == nil {
		return nil, errors.New("req is nil")
	}
	for key, val := range c.header { // 设置 header
		req.Header.Add(key, val)
	}
	for _, cookie := range c.cookies { // 添加 cookie
		if cookie == nil {
			continue
		}
		req.AddCookie(cookie)
	}
	for i := 0; i < c.TryTime; i++ { // 进行指定次数的重试
		res, err = c.client.Do(req)
		if err == nil {
			break
		}
	}
	return res, err
}

func (c HttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c HttpClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c HttpClient) PostFormByMap(URL string, data map[string]string) (resp *http.Response, err error) {
	body := make(url.Values)
	for k, v := range data {
		body.Add(k, v)
	}
	return c.PostForm(URL, body)
}

func (c HttpClient) PostJson(url string, data interface{}) (gjson.Result, error) {
	dataB, err := json.Marshal(data)
	if err != nil {
		return gjson.Result{}, err
	}
	body := bytes.NewReader(dataB)
	rsp, err := c.Post(url, "application/json", body)
	if err != nil {
		return gjson.Result{}, err
	}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(string(rspBody)), nil
}

func (c HttpClient) PostMarshal(url string, data interface{}, response interface{}) error {
	dataB, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(dataB)
	rsp, err := c.Post(url, "application/json", body)
	if err != nil {
		return err
	}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rspBody, response)
	return err
}

func (c HttpClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// GetReader 通过Get请求获取回包Body（io.Reader）
func (c HttpClient) GetReader(url string) (io.ReadCloser, error) {
	res, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

// GetGJson 通过Get请求获取回包Body（gjson.Result）
func (c HttpClient) GetGJson(url string) (gjson.Result, error) {
	res, err := c.GetReader(url)
	if err != nil {
		return gjson.Result{}, err
	}
	rspBody, err := ioutil.ReadAll(res)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(string(rspBody)), nil
}

// Head 发送Head请求
func (c HttpClient) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
