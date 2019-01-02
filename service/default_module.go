package service

type DefaultModule struct {
	env *Env
}

func (module *DefaultModule) Init(env *Env) error {
	module.env = env
	return nil
}

func (module *DefaultModule) Hello(req *HTTPRequest, result *Result) (e Error) {
	result.Res = "World"
	return
}
func (module *DefaultModule) SecHello(req *HTTPRequest, result *Result) (e Error) {
	result.Res = "Secure World!"
	return
}
func (module *DefaultModule) ErrorModule(req *HTTPRequest, result *Result) (e Error) {
	e.Desc = "Invalid Module Name"
	e.Code = ERR_INVALID_PARAM
	return
}
func (module *DefaultModule) ErrorMethod(req *HTTPRequest, result *Result) (e Error) {
	e.Desc = "Invalid Method Name"
	e.Code = ERR_INVALID_PARAM
	return
}
