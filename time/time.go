/****
	高性能时间，精确到0.1秒
****/
package time

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jiatower/go_lib/utils"
)

const (
	TIME_LAYOUT_1 = "2006-01-02 15:04:05"
	TIME_LAYOUT_2 = "2006-01-02"
	TIME_LAYOUT_3 = "2006-01-02T15:04:05Z"
	TIME_LAYOUT_4 = "2006年01月02日"
	TIME_LAYOUT_5 = "2006年01月"
	TIME_LAYOUT_6 = "2006年01月02日 15:04"
	TIME_LAYOUT_7 = "20060102"
	TIME_LAYOUT_8 = "01月02日"
	TIME_LAYOUT_9 = "Mon, 02 Jan 2006 15:04:05 MST"
)

var Now time.Time
var Local *time.Location

func init() {
	Now = time.Now().Round(time.Second)
	Local, _ = time.LoadLocation("Local")
	go refresh()
}

func refresh() {
	for {
		Now = time.Now().Round(time.Second)
		Local, _ = time.LoadLocation("Local")
		time.Sleep(100 * time.Millisecond)
	}
}

//获取当前时间戳，单位秒
func GetTimeStamp() int64 {
	return Now.Unix()
}

//当前时间多少秒后
func After(diff int) int64 {
	return Now.Unix() + int64(diff)
}

// 获取时间界限，如：today  返回stm: 2015-05-01 00:00:00  etm: 2015-05-02: 00:00:00
func TmLime(tmflag string) (stm, etm string) {
	stm = "1970-01-01 00:00:00"
	etm = "2070-01-01 00:00:00"
	if "today" == tmflag {
		stm = Now.Format("2006-01-02") + " 00:00:00"
		etm_tm := Now.AddDate(0, 0, 1)
		etm = etm_tm.Format("2006-01-02") + " 00:00:00"
	} else if "yesterday" == tmflag {
		stm_tm := Now.AddDate(0, 0, -1)
		stm = stm_tm.Format("2006-01-02") + " 00:00:00"
		etm = Now.Format("2006-01-02") + " 00:00:00"
	}
	return
}

func FormatPrevLogin(tm time.Time) (st string) {
	if Now.Before(tm) {
		return "当前在线"
	}
	du := Now.Sub(tm)
	switch {
	case du.Minutes() < 60:
		return utils.ToString(int(du.Minutes())) + "分钟前"
	case du.Hours() < 24:
		return utils.ToString(int(du.Hours())) + "小时前"
	case du.Hours() < 24*7:
		return utils.ToString(int(du.Hours()/24)) + "天前"
	default:
		return "七天前"
	}
	return
}

//转换时间到当前时间是多少天多少小时
func FormatRunTime(tm time.Time) (st string) {
	du := Now.Sub(tm)
	switch {
	case du.Minutes() < 0:
		st = "1分钟"
	case du.Minutes() < 60:
		st = utils.ToString(int(du.Minutes())) + "分钟"
	case du.Hours() < 24:
		hour := int(du.Hours())
		st = utils.ToString(hour) + "小时" + utils.ToString(int(du.Minutes())-60*hour) + "分钟"
	default:
		day := int(du.Hours() / 24)
		st = utils.ToString(day) + "天" + utils.ToString(int(du.Hours())-24*day) + "小时"
	}
	return
}

//打印超过exceed时长的时间
//Prams:
// 	key: 用于识别的关键字
// 	start: 起始时间
// 	exceed: 持续时间超过多久才打印
func PrintDuration(key string, start time.Time, exceed time.Duration) {
	dur := time.Now().Sub(start)
	if dur >= exceed {
		fmt.Println("Duration", key, ":", dur.Seconds())
	}
}

//从现在到[days]天后的[H:M:S]时刻的时长
func DurationTo(days int, H, M, S int) time.Duration {
	fmt.Println(Now.Minute(), Now.Second())
	seconds := (days*24+H-Now.Hour())*3600 + (M-Now.Minute())*60 + S - Now.Second()
	return time.Duration(seconds) * time.Second
}

//可以把字符串、时间、数字（当成秒）转换成时间类型
func ToTime(v interface{}, layout ...string) (t time.Time, e error) {
	switch value := v.(type) {
	case string:
		l := TIME_LAYOUT_1
		if len(layout) > 0 {
			l = layout[0]
		}
		return time.ParseInLocation(l, value, Local)
	case time.Time:
		return value, nil
	default:
		sec, e := utils.ToInt64(value)
		if e != nil {
			return t, errors.New(fmt.Sprintf("cannot change %v(%v) to time.Time", v, reflect.TypeOf(v)))
		}
		return time.Unix(sec, 0), nil
	}
}

//获取当前时间戳，单位秒
func GetTimeStampFromStr(time_str string) (tms int64, e error) {
	if time_str == "" {
		return
	}
	tm, e := ToTime(time_str)
	if e != nil {
		return
	}
	tms = tm.Unix()
	return
}

func FormatSec2Hour(tm int64) string {
	if tm <= 0 {
		return "0.00秒"
	}
	var minute float64 = 60
	var hour float64 = 60 * 60
	tmf := float64(tm)
	if tmf/hour >= 1 {
		return fmt.Sprintf("%.1f", tmf/hour) + "小时"
	} else if tmf/minute >= 1 {
		return fmt.Sprintf("%.1f", tmf/minute) + "分钟"
	} else {
		return fmt.Sprintf("%.0f", tmf) + "秒"
	}

}

func FormatSec2HourInt(tm int64) string {
	if tm <= 0 {
		return "0.00秒"
	}
	var minute float64 = 60
	var hour float64 = 60 * 60
	tmf := float64(tm)
	if tmf/hour >= 1 {
		return fmt.Sprintf("%.0f", tmf/hour) + "小时"
	} else if tmf/minute >= 1 {
		return fmt.Sprintf("%.0f", tmf/minute) + "分钟"
	} else {
		return fmt.Sprintf("%.0f", tmf) + "秒"
	}

}
func FormatDay2Year(days int) string {
	if days <= 0 {
		return "0天"
	}
	if days < 365 {
		return fmt.Sprintf("%v", days) + "天"
	} else {
		var y string
		year := days / 365
		day := days % 365
		if day == 0 {
			y = fmt.Sprintf("%v年", year)
		} else {
			y = fmt.Sprintf("%v年%v天", year, day)
		}
		return y
	}
}

func GetWeekdayCN(t time.Time) (w string) {
	we := t.Weekday().String()
	we = strings.ToLower(we)
	switch we {
	case time.Sunday.String():
		w = "星期日"
	case time.Saturday.String():
		w = "星期六"
	case time.Friday.String():
		w = "星期五"
	case time.Thursday.String():
		w = "星期四"
	case time.Wednesday.String():
		w = "星期三"
	case time.Tuesday.String():
		w = "星期二"
	case time.Monday.String():
		w = "星期一"
	}
	return
}
