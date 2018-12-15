/*
简易的Http框架，以json为传输格式。

错误返回值：
	{
		"code":2001,	//错误码
		"detail":"uid not provided",	//内部使用的错误详情
		"msg":"参数错误",	//客户端显示的错误原因
		"status":"fail"
	}
*/
package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
	"yh_pkg/log"

	_ "github.com/go-sql-driver/mysql"
)

type Server struct {
	modules      map[string]Module
	sysLog       *log.MLogger
	conf         *Config
	parseBody    bool //是否把POST的内容解析为json对象
	customResult bool //返回结果中是否包含result和tm项
}

func New(conf *Config, args ...bool) (server *Server, err error) {
	sysLog, err := log.NewMLogger(conf.LogDir+"/system", 10000, conf.LogLevel)
	if err != nil {
		return nil, err
	}
	server = &Server{make(map[string]Module), sysLog, conf, true, false}
	server.AddModule("default", &DefaultModule{})
	if len(args) >= 1 {
		server.parseBody = args[0]
	}
	if len(args) >= 2 {
		server.customResult = args[1]
	}
	return server, nil
}

func (server *Server) AddModule(name string, module Module) (err error) {
	fmt.Printf("add module %s... ", name)
	mlog, err := log.NewMLogger(server.conf.LogDir+"/"+name, 10000, server.conf.LogLevel)
	if err != nil {
		fmt.Println("failed")
		return err
	}
	env := server.conf.GetEnv(name)
	env.Log = mlog
	err = module.Init(env)
	if err != nil {
		return
	}
	fmt.Println("ok")
	mlog.Append("add module success", log.NOTICE)
	server.modules[name] = module
	return
}

func (server *Server) StartService() error {
	handler := http.NewServeMux()
	//用户验证
	handler.HandleFunc("/s/", server.secureHandler)
	//用户验证+集群绑定验证
	handler.HandleFunc("/sc/", server.secureHandler)
	//用户验证+集群绑定+管理员
	handler.HandleFunc("/sca/", server.secureHandler)
	handler.HandleFunc("/", server.nonSecureHandler)
	handler.HandleFunc("/stream", server.nonSecureHandler)
	s := &http.Server{
		Addr:           server.conf.IpPort,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 0,
	}
	l := fmt.Sprint("service start at ", server.conf.IpPort, " ...")
	server.sysLog.Append(l, log.NOTICE)
	fmt.Println(l)
	return s.ListenAndServe()
}

func (server *Server) writeBackErr(r *http.Request, w http.ResponseWriter, reqBody []byte, err Error, duration int64, req_res interface{}) {
	var result = Result{"fail", err.Show, err.Desc, err.Code, req_res, map[string]interface{}{}}
	res, _ := json.Marshal(result)
	server.writeBack(r, w, reqBody, res, false, duration)
}

func (server *Server) writeBack(r *http.Request, w http.ResponseWriter, reqBody []byte, result []byte, success bool, duration int64) {
	w.Write(result)
	var l string
	var response string
	/*	uidCookie, e := r.Cookie("uid")
		if e != nil {
			uid = ""
		} else {
			uid = uidCookie.Value
		}
	*/
	session := []string{""}
	for _, c := range r.Cookies() {
		session = append(session, c.String())
	}

	if reqBody != nil {
		response = strings.Replace(string(reqBody), " ", "", -1)
		response = strings.Replace(response, "\n", "", -1)
	}

	/*	format := "\nduration: %.2fms\n"
		format += "session: %s\n"
		format += "uri: %s\n"
		format += "param: %s\n"
		format += "response:\n%s\n"
		format += "------------------------------------------------------------------"
	*/
	req := &HttpRequest{nil, nil, r, nil}

	format := " | " + req.IP()
	format += " | uri: %s"
	format += " | duration: %.2fms"
	format += " | session: %s"
	format += " | param: %s"
	format += " | response:%s\n"
	format += "------------------------------------------------------------------"

	l = fmt.Sprintf(format, r.URL.String(), float64(duration)/1000000, session, response, string(result))
	if !success {
		server.sysLog.Append(l, log.ERROR)
	}

	url := r.URL.String()
	//配置不需要在System.debug.log中打印的url
	ex_url_map := map[string]interface{}{"/sc/storage/Punch": "", "/s/p2p_storage/ListUpdatedFiles": "", "/s/p2p_storage/Report": "", "/s/p2p_storage/GetDelegates": "", "/s/p2p_storage/UpdateNode": "", "/s/p2p_storage/UpdateNode2": "", "/sys_test/ErrorReport": "", "/sys_test/TimeoutRequest": "", "/sys_test/ReportLog": "", "/s/p2p_storage/Download": ""}
	if _, exist := ex_url_map[url]; exist {
		return
	}
	server.sysLog.Append(l, log.DEBUG)

}

func (server *Server) secureHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UnixNano()
	var result Result = NewResult()
	var err error
	s, e := server.conf.IsValidUser(r)
	var body []byte

	if e == nil {
		if s != nil && s.Uid > 0 {
			fields := strings.Split(r.URL.Path[1:], "/")
			if len(fields) >= 3 {
				pre_url := fields[0]
				if pre_url == "sc" || pre_url == "sca" {
					isOk, e := server.conf.CheckClusterUser(r, s)
					if e == nil {
						if isOk {
							if pre_url == "sc" {
								body, err = server.handleRequest(fields[1], "Sc"+fields[2], s, r, &result)
							} else {
								if s.IsAdmin || s.IsSuperAdmin {
									body, err = server.handleRequest(fields[1], "Sca"+fields[2], s, r, &result)
								} else {
									err = NewError(ERR_PERMISSION_DENIED, "no permission: "+r.URL.Path)
								}
							}
						} else {
							err = NewError(ERR_BINDDEV_INVALID, "not bind storage, or cluster_user is no ok: "+r.URL.Path)
						}
					} else {
						err = e
						result.Res = map[string]interface{}{"appid": s.AppId, "appuid": s.AppUid, "cluster": s.Cluster, "bind_tm": s.BindTm}
					}
				} else if pre_url == "s" {
					body, err = server.handleRequest(fields[1], "Sec"+fields[2], s, r, &result)
				}
			} else {
				err = NewError(ERR_INVALID_PARAM, "invalid url format : "+r.URL.Path)
			}
		} else {
			err = NewError(ERR_INVALID_USER, "invalid user")
		}
	} else {
		err = e
	}
	end := time.Now().UnixNano()
	server.processError(w, r, err, body, &result, end-start)
}

func (server *Server) nonSecureHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UnixNano()
	var result Result = NewResult()
	var err error

	s := new(Session)
	fields := strings.Split(r.URL.Path[1:], "/")
	var body []byte
	if len(fields) >= 2 {
		body, err = server.handleRequest(fields[0], fields[1], s, r, &result)
		if fields[0] == "stream" {
			body, err = server.handleStreamRequest(fields[1], fields[2], s, r, &result, w)
		}

	} else {
		err = NewError(ERR_INVALID_PARAM, "invalid url format : "+r.URL.Path)
	}
	end := time.Now().UnixNano()
	server.processError(w, r, err, body, &result, end-start)
}

func (server *Server) processError(w http.ResponseWriter, r *http.Request, err error, reqBody []byte, result *Result, duration int64) {
	var re Error
	switch e := err.(type) {
	case nil:
	case Error:
		re = e
	default:
		re = NewError(ERR_INTERNAL, e.Error(), "未知错误")
	}
	// 302跳转
	if re.Code == SERVER_REDIRECT {
		if url, found := result.GetString(SERVER_REDIRECT_KEY); found {
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
	}

	if re.Code == ERR_NOERR {
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		/*
			if !server.customResult {
				result["status"] = "ok"
				result["tm"] = utils.Now.Unix()
			}
		*/
		res, e := json.Marshal(result)
		//res, e := json.MarshalIndent(result, "", " ")
		if e == nil {
			server.writeBack(r, w, reqBody, res, true, duration)
		} else {
			server.writeBackErr(r, w, reqBody, NewError(ERR_INTERNAL, fmt.Sprintf("Marshal result error : %v", e.Error())), duration, result.Res)
		}
	} else if re.Code == SERVER_WRITE_IMAGE { //输出图片信息
		w.Header().Set("Content-Type", "image/png")
		switch b := result.Res.(type) {
		case nil:
		case []byte:
			w.Write(b)
		}
	} else if re.Code == SERVER_WRITE_MP4 { //输出视频信息
		w.Header().Set("Content-Type", "video/mp4")
		//w.Header().Set("connection", "keep-alive")
		switch b := result.Res.(type) {
		case nil:
		case []byte:
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
			w.Write(b)
		}
	} else if re.Code == SERVER_WRITE_TEXT { //输出文字信息
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		desc := "error"
		if re.Desc != "" {
			desc = re.Desc
		}
		w.Write([]byte(desc))
	} else if re.Code == SERVER_WRITE_XML { //输出xml信息
		w.Header().Set("Content-Type", "text/xml")
		desc := "xml"
		if re.Desc != "" {
			desc = re.Desc
		}
		w.Write([]byte(desc))

	} else if v, ok := DocContentTypeMap[re.Code]; ok { //输出文档类数据
		w.Header().Set("Content-Type", v)
		switch b := result.Res.(type) {
		case nil:
		case []byte:
			w.Write(b)
		}
	} else {
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		server.writeBackErr(r, w, reqBody, re, duration, result.Res)
	}

}

func (server *Server) handleRequest(moduleName string, methodName string, s *Session, r *http.Request, result *Result) ([]byte, error) {
	bodyBytes := make([]byte, 0, 10)
	var e error
	var body map[string]interface{}
	if moduleName == "upload" || moduleName == "wechat_msg" {

	} else {
		bodyBytes, e = ioutil.ReadAll(r.Body)
		if e != nil {
			return nil, NewError(ERR_INTERNAL, "read http data error : "+e.Error())
		}
		if len(bodyBytes) == 0 {
			//可能是Get请求
			body = make(map[string]interface{})
		} else if server.parseBody {
			e = json.Unmarshal(bodyBytes, &body)
			if e != nil {
				return bodyBytes, NewError(ERR_INVALID_PARAM, "read body error : "+e.Error())
			}
		}
	}
	var values []reflect.Value
	module, ok := server.modules[moduleName]
	if ok {
		method := reflect.ValueOf(module).MethodByName(methodName)
		if method.IsValid() {
			values = method.Call([]reflect.Value{reflect.ValueOf(&HttpRequest{body, bodyBytes, r, s}), reflect.ValueOf(result)})
		} else {
			method = reflect.ValueOf(server.modules["default"]).MethodByName("ErrorMethod")
			values = method.Call([]reflect.Value{reflect.ValueOf(&HttpRequest{body, bodyBytes, r, s}), reflect.ValueOf(result)})
		}
	} else {
		method := reflect.ValueOf(server.modules["default"]).MethodByName("ErrorModule")
		values = method.Call([]reflect.Value{reflect.ValueOf(&HttpRequest{body, bodyBytes, r, s}), reflect.ValueOf(result)})
	}
	if len(values) != 1 {
		return bodyBytes, NewError(ERR_INTERNAL, fmt.Sprintf("method %s.%s return value is not 2.", moduleName, methodName))
	}
	switch x := values[0].Interface().(type) {
	case nil:
		return bodyBytes, nil
	default:
		return bodyBytes, x.(error)
	}
}

//处理下载较大文件流的请求
func (server *Server) handleStreamRequest(moduleName string, methodName string, s *Session, r *http.Request, result *Result, w http.ResponseWriter) ([]byte, error) {
	bodyBytes := make([]byte, 0, 10)
	var e error
	var body map[string]interface{}
	if moduleName == "upload" || moduleName == "wechat_msg" {

	} else {
		bodyBytes, e = ioutil.ReadAll(r.Body)
		if e != nil {
			return nil, NewError(ERR_INTERNAL, "read http data error : "+e.Error())
		}
		if len(bodyBytes) == 0 {
			//可能是Get请求
			body = make(map[string]interface{})
		} else if server.parseBody {
			e = json.Unmarshal(bodyBytes, &body)
			if e != nil {
				return bodyBytes, NewError(ERR_INVALID_PARAM, "read body error : "+e.Error())
			}
		}
	}
	var values []reflect.Value
	module, ok := server.modules[moduleName]
	if ok {
		method := reflect.ValueOf(module).MethodByName(methodName)
		if method.IsValid() {
			values = method.Call([]reflect.Value{reflect.ValueOf(&HttpRequest{body, bodyBytes, r, s}), reflect.ValueOf(result), reflect.ValueOf(w)})
		} else {
			method = reflect.ValueOf(server.modules["default"]).MethodByName("ErrorMethod")
			values = method.Call([]reflect.Value{reflect.ValueOf(&HttpRequest{body, bodyBytes, r, s}), reflect.ValueOf(result)})
		}
	} else {
		method := reflect.ValueOf(server.modules["default"]).MethodByName("ErrorModule")
		values = method.Call([]reflect.Value{reflect.ValueOf(&HttpRequest{body, bodyBytes, r, s}), reflect.ValueOf(result)})
	}
	if len(values) != 1 {
		return bodyBytes, NewError(ERR_INTERNAL, fmt.Sprintf("method %s.%s return value is not 2.", moduleName, methodName))
	}
	switch x := values[0].Interface().(type) {
	case nil:
		return bodyBytes, nil
	default:
		return bodyBytes, x.(error)
	}
}
