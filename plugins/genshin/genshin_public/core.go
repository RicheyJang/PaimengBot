package genshin_public

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MiyoRequest struct {
	url     string
	headers map[string]string
}

type MiyoResponse struct {
	RetCode int             `json:"retcode"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func NewMiyoRequest(url string) *MiyoRequest {
	return &MiyoRequest{
		url,
		map[string]string{
			"Accept": "application/json",
		},
	}
}

func (th *MiyoRequest) Execute() ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, th.url, nil)
	if err != nil {
		return []byte{}, err
	}
	for k, v := range th.headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var miyoResp MiyoResponse
	err = dec.Decode(&miyoResp)
	if err != nil {
		return []byte{}, err
	}
	if miyoResp.RetCode != 0 {
		return []byte{}, fmt.Errorf("errcode: %d, errmsg: %s", miyoResp.RetCode, miyoResp.Message)
	}
	return miyoResp.Data, nil
}

func (th *MiyoRequest) SetHeader(k, v string) {
	th.headers[k] = v
}
