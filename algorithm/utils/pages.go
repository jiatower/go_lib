package utils

import "strconv"

const DEFAULT_PS = 10
const DEFAULT_CUR = 1
const PAGE_HASNEXT_KEY = "hasNext"
const PAGE_RES_NAME = "list"

type Pages struct {
	Cur   int `json:"cur"`   // 当前页
	Total int `json:"total"` // 总条数
	Ps    int `json:"ps"`    // 每页显示数量
	Pn    int `json:"pn"`    // 总页数
}

type PageRes struct {
	HasNext int         `json:"hasNext"`
	Start   string      `json:"start"`
	Ps      int         `json:"ps"`
	List    interface{} `json:"list"`
}

// 封装返回，获取分页信息
func PageInfo(total, cur, ps int) *Pages {
	pages := new(Pages)
	if total < 0 {
		pages.Cur = cur
		pages.Total = -1
		pages.Ps = ps
		pages.Pn = -1
		return pages
	}
	pages.Total = total
	pages.Ps = ps
	// 计算共有分页数....
	if cur <= 1 {
		cur = 1
	}
	pn := total / ps
	if total%ps != 0 {
		pn += 1
	}
	pages.Pn = pn

	if cur > pn {
		cur = pn
	}
	pages.Cur = cur
	return pages
}

// 生成 分页limit 10，10 字符串
func BuildLimit(cur, ps int) string {
	begin := 0
	if cur > 1 {
		begin = (cur - 1) * ps
	}
	lsql := " limit " + strconv.Itoa(begin) + "," + strconv.Itoa(ps)
	return lsql
}

// 生成redis分页
func BuildRange(cur, ps, total int) (int, int) {
	begin := 0
	if cur > 1 {
		begin = (cur - 1) * ps
	}
	end := begin + ps - 1
	if total > 0 && end >= total {
		end = total - 1
	}
	return begin, end
}

//简要分页，获取是否有下一页，简单判断获取数据是否等于请求数量
func CheckHashNext(ps, len int) (hasNext int) {
	if len >= ps {
		hasNext = 1
	}
	return
}

//构建包含list数据的分页结果
func FormatPageRes(ps, len int, start string, data interface{}) (res PageRes) {
	var hasNext int
	if len >= ps {
		hasNext = 1
	}
	res.HasNext = hasNext
	res.Ps = ps
	res.Start = start
	res.List = data
	return
}

//获取分页结构
func FormatPages(ps, len int) (pages PageRes) {
	pages = FormatPageRes(ps, len, "", []interface{}{})
	return
}
