package sign_model

// ResponseInfo 基础序列化器
type ResponseInfo struct {
	Code int    `json:"retcode"`
	Msg  string `json:"message"`
}
