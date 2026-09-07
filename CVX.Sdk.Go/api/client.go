package api

// Client 接口定义
// 所有API模块都依赖此接口来执行HTTP请求
type Client interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
}
