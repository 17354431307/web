package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// 解决大多数人的需求

type Context struct {
	Req *http.Request

	// Resp 如果用户直接使用这个
	// 那么他们就绕开了 RespData 和 RespStatusCode 这两个
	// 那么部分 middleware 无法运作
	Resp http.ResponseWriter

	// 这个注意是为了给 middleware 读写用的
	RespData       []byte
	RespStatusCode int

	PathParams map[string]string

	queryValues url.Values

	MatchedRoute string
}

func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	c.Resp.WriteHeader(status)
	c.Resp.Header().Set("Content-Type", "application/json")
	//c.Resp.Header().Set("Content-Length", strconv.Itoa(len(data)))
	c.RespData = data
	c.RespStatusCode = status
	return nil
}

func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(200, val)
}

func (c *Context) BindJSON(val any) error {
	if val == nil {
		return errors.New("web: 输入为 nil")
	}

	if c.Req.Body == nil {
		return errors.New("web: body 为 nil")
	}

	decoder := json.NewDecoder(c.Req.Body)
	// useNumber => 数字就是用 Number 来表示
	// 否则默认是 float64
	//decoder.UseNumber()

	// DisallowUnknownFields 如果要是有一个未知的字段, 就会报错
	// 比如说你 User 只有 Name 和 Email 两个字段
	// JSON 里面额外多了一个 Age 字段, 那么就会报错
	//decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

func (c *Context) FormValue(key string) (string, error) {
	err := c.Req.ParseForm()
	if err != nil {
		return "", err
	}
	vals, ok := c.Req.Form[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return vals[0], nil
}

// Query 和 表达比起来, 它没有缓存
func (c *Context) QueryValue(key string) (string, error) {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}

	// 用户区别不出来是真的有值, 但是值恰好是空字符串
	// 还是没有值(go 的零值问题)

	vals, ok := c.queryValues[key]
	if !ok || len(vals) == 0 {
		return "", errors.New("web: key 不存在")
	}
	return vals[0], nil
}

func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}

	return val, nil
}

// 返回结构体的好处就是可以给结构体加方法, 就可以连续调用转类型的方法
func (c *Context) PathValueV1(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{
			val: "",
			err: errors.New("web: key 不存在"),
		}
	}

	return StringValue{
		val: val,
		err: nil,
	}
}

type StringValue struct {
	val string
	err error
}

func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.ParseInt(s.val, 10, 64)
}

// 原本的 Context 不是线程安全的, 我们也没必要设计成线程安全
// 如果用户非要怎么办? 让他自己使用装饰器模式用锁包装 Context
type SafeContext struct {
	Context
	mutex sync.RWMutex
}

func (c *SafeContext) RespJSON(val any) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Context.RespJSONOK(val)
}
