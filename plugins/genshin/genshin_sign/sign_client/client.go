package sign_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_util/constant"
	"io/ioutil"

	"net/http"
)

type HTTPClient struct {
	*http.Client
}

func NewClient() (g *HTTPClient) {
	g = &HTTPClient{
		Client: &http.Client{},
	}
	return
}

// NewRequest 建立消息
//  param:
//    method	string		请求方法
//    url		string		地址
//    body		interface{}	Body数据
//  return:
//    req		*http.Request	请求
//    err		error			错误
func (g *HTTPClient) NewRequest(method, url string, body interface{}) (req *http.Request, err error) {
	var jsonByte []byte = nil
	if body != nil {
		jsonByte, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable convert request body to json:%v", err)
		}
	}

	req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonByte))
	if err != nil {
		return nil, fmt.Errorf("unable to new http request:%v", err)
	}

	return
}

// NewGetRequest 建立GET消息
//  param:
//    url		string		地址
//    body		interface{}	Body数据
//  return:
//    req		*http.Request	请求
//    err		error			错误
func (g *HTTPClient) NewGetRequest(url string, body interface{}) (req *http.Request, err error) {
	return g.NewRequest(http.MethodGet, url, body)
}

// NewPostRequest 建立post消息
//  param:
//    url		string		地址
//    body		interface{}	Body数据
//  return:
//    req		*http.Request	请求
//    err		error			错误
func (g *HTTPClient) NewPostRequest(url string, body interface{}) (req *http.Request, err error) {
	return g.NewRequest(http.MethodPost, url, body)
}

// Send 发送http请求
//  param:
//    req	*http.Request	请求
//    pRet	interface{}		返回模型，必须是指针
//  return:
//    err	error	错误
func (g *HTTPClient) Send(req *http.Request, pRet interface{}) (err error) {
	resp, err := g.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable convert response body io to json:%v", err)
	}

	if err = json.Unmarshal(respBody, pRet); err != nil {
		return fmt.Errorf("unable convert response body json to struct:%v", err)
	}

	return
}

// SendMessage 发送消息
//  param:
//    method	string		请求方法
//    url		string		地址
//    pRet		interface{}	返回模型，必须是指针
//    modify	func(req *http.Request) error	用于修改请求的函数
//  return:
//    err		error	错误
func (g *HTTPClient) SendMessage(method, url string, body, pRet interface{}, modify ...func(req *http.Request) error) (err error) {
	req, err := g.NewRequest(method, url, body)
	if err != nil {
		return
	}
	if len(modify) > 0 {
		for _, f := range modify {
			f(req)
		}
	}

	return g.Send(req, pRet)
}

// SendGetMessage 发送GET消息
//  param:
//    url		string		地址
//    pRet		interface{}	返回模型，必须是指针
//    modify	func(req *http.Request) error	用于修改请求的函数
//  return:
//    err		error	错误
func (g *HTTPClient) SendGetMessage(url string, body, pRet interface{}, modify ...func(req *http.Request) error) (err error) {
	return g.SendMessage(http.MethodGet, url, body, pRet, modify...)
}

// SendPostMessage 发送POST消息
//  param:
//    url		string		地址
//    pRet		interface{}	返回模型，必须是指针
//    modify	func(req *http.Request) error	用于修改请求的函数
//  return:
//    err		error	错误
func (g *HTTPClient) SendPostMessage(url string, body, pRet interface{}, modify ...func(req *http.Request) error) (err error) {
	return g.SendMessage(http.MethodPost, url, body, pRet, modify...)
}

// AddBasicHeader 添加一些基础的消息头
//  return:
//    func(req *http.Request) error	用于修改请求的函数
func AddBasicHeader() func(req *http.Request) error {
	return func(req *http.Request) error {
		req.Header.Add("Accept-Encoding", constant.AcceptEncoding)
		req.Header.Add("User-Agent", constant.UserAgent)
		req.Header.Add("Referer", constant.AcceptEncoding)
		req.Header.Add("Accept-Encoding", constant.AcceptEncoding)
		return nil
	}
}

// AddCookieHeader 添加Cookie记录消息头
//  param:
//    cookie	string	Cookie内容
//  return:
//    func(req *http.Request) error	用于修改请求的函数
func AddCookieHeader(cookie string) func(req *http.Request) error {
	return func(req *http.Request) error {
		req.Header.Add("Cookie", cookie)
		return nil
	}
}
