package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var client http.Client

const (
	//DefaultTimeout 默认超时时间（秒）
	DefaultTimeout = 10
)

func init() {
	client.Timeout = DefaultTimeout * time.Second
}

//SendHTTP 发送http的GET或POST请求
//data如果为nil，则发GET请求，否则发POST请求
func SendHTTP(host string, path string, params map[string]string, cookies map[string]string, data []byte) (body []byte, e error) {
	return Send("http", host, path, params, nil, cookies, data)
}

//GetHTTP GET模式请求
func GetHTTP(host string, path string, params map[string]string, timeout int) (body []byte, e error) {
	return send("http", host, path, params, nil, nil, nil, timeout)
}

//SendHTTPS 发送https的GET或POST请求k
//data如果为nil，则发GET请求，否则发POST请求
func SendHTTPS(host string, path string, params map[string]string, cookies map[string]string, data []byte) (body []byte, e error) {
	return Send("https", host, path, params, nil, cookies, data)
}

//SendForJSON 参数和返回值都是json格式的http(s)GET或POST请求
//data如果为nil，则发GET请求，否则发POST请求
func SendForJSON(protocal, host string, path string, params map[string]string, cookies map[string]string, data interface{}, result interface{}) (e error) {
	b, e := json.Marshal(data)
	if e != nil {
		return e
	}
	body, e := Send(protocal, host, path, params, nil, cookies, b)
	if e != nil {
		return
	}
	return json.Unmarshal(body, result)
}

//GetHTTPS GET模式发送HTTPS请求
func GetHTTPS(host string, path string, params map[string]string, timeout int) (body []byte, e error) {
	return send("https", host, path, params, nil, nil, nil, timeout)
}

//Send 发送http(s)的GET或POST请求
//data如果为nil，则发GET请求，否则发POST请求
func Send(protocal string, host string, path string, params map[string]string, header map[string]string, cookies map[string]string, data []byte, timeout ...int) (body []byte, e error) {
	to := DefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	return send(protocal, host, path, params, header, cookies, data, to)
}

//send http(s)的内部实现
//protocal 就是url的最前面的协议名称，即http/https等
func send(protocal string, host string, path string, params map[string]string, header map[string]string, cookies map[string]string, data []byte, timeout int) (body []byte, e error) {
	m := "GET"
	if data != nil {
		m = "POST"
	}
	v := url.Values{}
	for key, value := range params {
		v.Set(key, value)
	}
	reqURL := &url.URL{
		Host:     host,
		Scheme:   protocal,
		Path:     path,
		RawQuery: v.Encode(),
	}
	req, e := http.NewRequest(m, reqURL.String(), bytes.NewBuffer(data))
	if e != nil {
		return nil, e
	}
	req.Close = true
	for k, v := range header {
		req.Header.Add(k, v)
	}
	for k, v := range cookies {
		var cookie http.Cookie
		cookie.Name = k
		cookie.Value = v
		req.AddCookie(&cookie)
	}
	c := &client
	if timeout != DefaultTimeout {
		c = &http.Client{}
		c.Timeout = time.Duration(timeout) * time.Second
	}
	resp, e := c.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	body, e = ioutil.ReadAll(resp.Body)
	return
}
