package yaml

import (
	"fmt"
	"net"
	"os"

	yaml "gopkg.in/yaml.v2"
)

//Address IPV4地址
type Address struct {
	//Ip IPV4
	IP string `yaml:"ip"`
	//Port 端口
	Port int `yaml:"port"`
}

//Addresses IPV4地址列表
type Addresses []Address

//MysqlType mysql配置信息
type MysqlType struct {
	Master string   `yaml:"master"`
	Slave  []string `yaml:"slave"`
}

//RedisType redis配置信息
type RedisType struct {
	Master  Address   `yaml:"master"`
	Slave   Addresses `yaml:"slave"`
	MaxConn int       `yaml:"max_conn"`
}

func (a *Address) String(lookup ...bool) string {
	if a.IP == "0.0.0.0" {
		return fmt.Sprintf(":%v", a.Port)
	}
	if len(lookup) > 0 && lookup[0] == false {
		return fmt.Sprintf("%s:%v", a.IP, a.Port)
	}

	ip, e := net.LookupIP(a.IP)
	if e != nil {
		return fmt.Sprintf("%s:%v", a.IP, a.Port)
	} else {
		return fmt.Sprintf("%s:%v", ip[0].String(), a.Port)
	}
}

//StringSlice  把地址列表转换成字符串数组
func (as Addresses) StringSlice() []string {
	ret := make([]string, len(as))
	for i, v := range as {
		ret[i] = v.String()
	}
	return ret
}

//Load 读取配置文件到c数据结构
func Load(c interface{}, path string) error {
	file, e := os.Open(path)
	if e != nil {
		return e
	}
	info, e := file.Stat()
	if e != nil {
		return e
	}
	defer file.Close()
	data := make([]byte, info.Size())
	n, e := file.Read(data)
	if e != nil {
		return e
	}
	if int64(n) < info.Size() {
		return fmt.Errorf("cannot read %v bytes from %v", info.Size(), path)
	}

	e = yaml.Unmarshal([]byte(data), c)
	return e
}

//String 将配置转换成字符串
func String(c interface{}) (string, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
