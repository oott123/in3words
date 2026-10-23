package wpfmt

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// phpArray 是 PHP 数组的有序表示。键统一保存为字符串（整数键转为十进制字符串，与 PHP 的键规范化一致）。
type phpArray struct {
	keys   []string
	values map[string]any
}

func newPHPArray() *phpArray {
	return &phpArray{values: map[string]any{}}
}

// set 写入一个键值；已存在的键保留原有位置。
func (a *phpArray) set(key string, value any) {
	if _, ok := a.values[key]; !ok {
		a.keys = append(a.keys, key)
	}
	a.values[key] = value
}

func (a *phpArray) get(key string) (any, bool) {
	if a == nil {
		return nil, false
	}
	v, ok := a.values[key]
	return v, ok
}

func (a *phpArray) len() int {
	if a == nil {
		return 0
	}
	return len(a.keys)
}

// errUnserialize 表示序列化数据格式不受支持或已损坏。
var errUnserialize = errors.New("invalid php serialized data")

// phpUnserialize 解析 PHP serialize() 的输出，支持 N、b、i、d、s、a 类型；遇到对象等其他类型时返回错误。
func phpUnserialize(data string) (any, error) {
	p := &phpUnserializer{data: data}
	v, err := p.value()
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.data) {
		return nil, fmt.Errorf("unserialize trailing data at offset %d: %w", p.pos, errUnserialize)
	}
	return v, nil
}

type phpUnserializer struct {
	data string
	pos  int
}

func (p *phpUnserializer) fail() error {
	return fmt.Errorf("unserialize at offset %d: %w", p.pos, errUnserialize)
}

func (p *phpUnserializer) expect(s string) error {
	if !strings.HasPrefix(p.data[p.pos:], s) {
		return p.fail()
	}
	p.pos += len(s)
	return nil
}

// until 读取到分隔符 sep 之前的内容，并跳过分隔符。
func (p *phpUnserializer) until(sep byte) (string, error) {
	end := strings.IndexByte(p.data[p.pos:], sep)
	if end < 0 {
		return "", p.fail()
	}
	s := p.data[p.pos : p.pos+end]
	p.pos += end + 1
	return s, nil
}

func (p *phpUnserializer) value() (any, error) {
	if p.pos+1 >= len(p.data) {
		return nil, p.fail()
	}
	typ := p.data[p.pos]
	if typ == 'N' {
		p.pos++
		return nil, p.expect(";")
	}
	p.pos++
	if err := p.expect(":"); err != nil {
		return nil, err
	}
	switch typ {
	case 'b':
		s, err := p.until(';')
		if err != nil {
			return nil, err
		}
		return s == "1", nil
	case 'i':
		s, err := p.until(';')
		if err != nil {
			return nil, err
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, p.fail()
		}
		return n, nil
	case 'd':
		s, err := p.until(';')
		if err != nil {
			return nil, err
		}
		return parsePHPFloat(s)
	case 's':
		return p.str()
	case 'a':
		return p.array()
	default:
		return nil, p.fail()
	}
}

func parsePHPFloat(s string) (float64, error) {
	switch s {
	case "INF":
		return math.Inf(1), nil
	case "-INF":
		return math.Inf(-1), nil
	case "NAN":
		return math.NaN(), nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float %q: %w", s, errUnserialize)
	}
	return f, nil
}

func (p *phpUnserializer) str() (string, error) {
	lenStr, err := p.until(':')
	if err != nil {
		return "", err
	}
	n, err := strconv.Atoi(lenStr)
	if err != nil || n < 0 || p.pos+n+2 > len(p.data) || p.data[p.pos] != '"' {
		return "", p.fail()
	}
	s := p.data[p.pos+1 : p.pos+1+n]
	p.pos += n + 1
	if expectErr := p.expect(`";`); expectErr != nil {
		return "", expectErr
	}
	return s, nil
}

func (p *phpUnserializer) array() (*phpArray, error) {
	countStr, err := p.until(':')
	if err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 0 {
		return nil, p.fail()
	}
	if err = p.expect("{"); err != nil {
		return nil, err
	}
	arr := newPHPArray()
	for range count {
		var key, val any
		if key, err = p.value(); err != nil {
			return nil, err
		}
		var k string
		switch kv := key.(type) {
		case int64:
			k = strconv.FormatInt(kv, 10)
		case string:
			k = kv
		default:
			return nil, p.fail()
		}
		if val, err = p.value(); err != nil {
			return nil, err
		}
		arr.set(k, val)
	}
	if err = p.expect("}"); err != nil {
		return nil, err
	}
	return arr, nil
}

// phpToFloat 对应 PHP 中把标量当作数字使用时的转换。
func phpToFloat(v any) float64 {
	switch x := v.(type) {
	case int64:
		return float64(x)
	case float64:
		return x
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		num := phpNumericPrefix.FindString(strings.TrimLeft(x, " \t\n\r\v\f"))
		f, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

// phpToInt 对应 PHP 的 (int) 转换。
func phpToInt(v any) int {
	switch x := v.(type) {
	case int64:
		return int(x)
	case string:
		return phpIntval(x)
	default:
		f := phpToFloat(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0
		}
		return int(f)
	}
}

// phpToString 对应 PHP 中把标量转换为字符串。
func phpToString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		if x == math.Trunc(x) && math.Abs(x) < 1e15 {
			return strconv.FormatFloat(x, 'f', -1, 64)
		}
		return strconv.FormatFloat(x, 'G', 14, 64)
	case bool:
		if x {
			return "1"
		}
		return ""
	default:
		return ""
	}
}

// phpArrayKey 返回值作为 PHP 数组键时的规范形式：整数与十进制整数字符串都变成整数键，浮点数截断为整数。
func phpArrayKey(v any) string {
	switch x := v.(type) {
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case bool:
		if x {
			return "1"
		}
		return "0"
	case string:
		return x
	default:
		return ""
	}
}
