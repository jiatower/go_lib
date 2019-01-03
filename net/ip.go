package net

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
)

//Server 获取ip地址信息的url地址
const Server string = "http://ip.taobao.com/service/getIpInfo.php?ip="

//IPInfo IP信息
type IPInfo struct {
	Code int `json:"code"`
	Data IP  `json:"data"`
}

//IP IP信息
type IP struct {
	Country   string `json:"country"`
	CountryID string `json:"country_id"`
	Area      string `json:"area"`
	AreaID    string `json:"area_id"`
	Region    string `json:"region"`
	RegionID  string `json:"region_id"`
	City      string `json:"city"`
	CityID    string `json:"city_id"`
	Isp       string `json:"isp"`
}

//TaobaoAPI 淘宝获取IP信息的接口
func TaobaoAPI(ip string) *IPInfo {
	url := Server + ip

	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var result IPInfo
	if err := json.Unmarshal(out, &result); err != nil {
		return nil
	}

	return &result
}

//GetIntranetIPv4 获取本机IPv4地址
func GetIntranetIPv4() (ipList []net.IP, e error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return nil, err
	}

	ipList = make([]net.IP, 0)
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			if ipnet.IP.IsInterfaceLocalMulticast() || ipnet.IP.IsLinkLocalMulticast() || ipnet.IP.IsLinkLocalUnicast() || ipnet.IP.IsLoopback() || ipnet.IP.IsMulticast() || ipnet.IP.IsUnspecified() {
				continue
			}
			ipList = append(ipList, ipnet.IP)
		}
	}
	return
}
