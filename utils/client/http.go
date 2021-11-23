package client

import (
	"errors"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	HttpOptions
	client *http.Client
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
	}
}

func (c HttpClient) Do(req *http.Request) (*http.Response, error) {
	var res *http.Response
	err := errors.New("TryTime is zero, send no http request")
	if req == nil {
		return nil, errors.New("req is nil")
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

func (c HttpClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// GetReader 通过Get请求获取回包Body（io.Reader）
func (c HttpClient) GetReader(url string) (io.Reader, error) {
	res, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
