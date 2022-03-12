package sign_model

type SignResponseInfo struct {
	*ResponseInfo
	Data SignResponseData `json:"data,omitempty"`
}

type SignResponseData struct {
	Code string `json:"code,omitempty"`
}
