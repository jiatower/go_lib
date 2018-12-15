package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//要求to必须已经分配好空间
func Uint32ToBytes(from uint32) (to []byte) {
	to = make([]byte, 4)
	b_buf := new(bytes.Buffer)
	binary.Write(b_buf, binary.BigEndian, from)
	copy(to, b_buf.Bytes()[0:4])
	return
}

func BytesToUint32(from []byte) (to uint32) {
	b_buf := bytes.NewBuffer(from)
	binary.Read(b_buf, binary.BigEndian, &to)
	return
}

//把interface类型转换成string类型
func ToString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func ToBool(v interface{}) (bool, error) {
	switch value := v.(type) {
	case bool:
		return value, nil
	case string:
		switch value {
		case "true", "True":
			return true, nil
		case "false", "False":
			return false, nil
		default:
			return false, errors.New("cannot convert " + value + " to bool")
		}
	case float32:
		return value != 0, nil
	case float64:
		return value != 0, nil
	case int8:
		return value != 0, nil
	case int16:
		return value != 0, nil
	case int32:
		return value != 0, nil
	case int:
		return value != 0, nil
	case int64:
		return value != 0, nil
	case uint8:
		return value != 0, nil
	case uint16:
		return value != 0, nil
	case uint32:
		return value != 0, nil
	case uint:
		return value != 0, nil
	case uint64:
		return value != 0, nil
	default:
		return false, errors.New(fmt.Sprintf("cannot convert %v(%v) to bool", v, reflect.TypeOf(v)))
	}
}
func ToFloat64(v interface{}) (float64, error) {
	switch value := v.(type) {
	case string:
		return strconv.ParseFloat(value, 10)
	case float32:
		return float64(value), nil
	case float64:
		return value, nil
	case int8:
		return float64(value), nil
	case int16:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case int:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case uint8:
		return float64(value), nil
	case uint16:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	default:
		return 0, errors.New(fmt.Sprintf("cannot convert %v(%v) to float64", v, reflect.TypeOf(v)))
	}
}

func ToUint8(v interface{}) (uint8, error) {
	i, e := ToFloat64(v)
	return uint8(i), e
}
func ToUint16(v interface{}) (uint16, error) {
	i, e := ToFloat64(v)
	return uint16(i), e
}
func ToUint32(v interface{}) (uint32, error) {
	i, e := ToFloat64(v)
	return uint32(i), e
}
func ToUint(v interface{}) (uint, error) {
	i, e := ToFloat64(v)
	return uint(i), e
}
func ToUint64(v interface{}) (uint64, error) {
	i, e := ToFloat64(v)
	return uint64(i), e
}
func ToInt8(v interface{}) (int8, error) {
	i, e := ToFloat64(v)
	return int8(i), e
}
func ToInt16(v interface{}) (int16, error) {
	i, e := ToFloat64(v)
	return int16(i), e
}
func ToInt(v interface{}) (int, error) {
	i, e := ToFloat64(v)
	return int(i), e
}
func ToInt32(v interface{}) (int32, error) {
	i, e := ToFloat64(v)
	return int32(i), e
}
func ToInt64(v interface{}) (int64, error) {
	i, e := ToFloat64(v)
	return int64(i), e
}
func ToFloat32(v interface{}) (float32, error) {
	i, e := ToFloat64(v)
	return float32(i), e
}
func ToStringSlice(v interface{}) ([]string, error) {
	switch slice := v.(type) {
	case []string:
		return slice, nil
	case []interface{}:
		r := make([]string, 0, len(slice))
		for _, item := range slice {
			r = append(r, fmt.Sprintf("%v", item))
		}
		return r, nil
	default:
		return nil, errors.New(fmt.Sprintf("cannot convert %v(%v) to []string", v, reflect.TypeOf(v)))
	}
}

func BirthdayToAge(birthday time.Time) int {
	if birthday.After(time.Now()) {
		return 0
	}
	return int(time.Now().Year() - birthday.Year())
}

func AgeToBirthday(Age int) time.Time {
	return time.Now().AddDate(-Age, 0, 0)
}

func Join(v interface{}, sep string) (string, error) {
	switch slice := v.(type) {
	case []string:
		return strings.Join(slice, sep), nil
	case []uint32:
		if len(slice) == 0 {
			return "", nil
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%v", slice[0])
		for i := 1; i < len(slice); i++ {
			fmt.Fprintf(&buf, "%s%v", sep, slice[i])
		}
		return buf.String(), nil
	case []uint64:
		if len(slice) == 0 {
			return "", nil
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%v", slice[0])
		for i := 1; i < len(slice); i++ {
			fmt.Fprintf(&buf, "%s%v", sep, slice[i])
		}
		return buf.String(), nil
	case []interface{}:
		if len(slice) == 0 {
			return "", nil
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%v", slice[0])
		for i := 1; i < len(slice); i++ {
			fmt.Fprintf(&buf, "%s%v", sep, slice[i])
		}
		return buf.String(), nil
	default:
		return "", errors.New(fmt.Sprintf("cannot convert %v(%v) to Slice", v, reflect.TypeOf(v)))
	}
}
