package hresty

import (
	"errors"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
)

// NewRestyClient 创建带连接池与超时配置的 resty 客户端
func NewRestyClient() *resty.Client {
	return resty.NewWithClient(newClient())
}

func newClient() *http.Client {
	return &http.Client{
		Transport: NewTransport(),
		Timeout:   10 * time.Second,
	}
}

// NewTransport 默认 transport 配置,可被外部直接使用
func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxConnsPerHost:       runtime.GOMAXPROCS(0) * 64,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) * 64,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
	}
}

// RequestTrace 结构化的请求追踪信息
// 调用方拿到后可自行用任意 logger(zap/logrus/std log)输出
type RequestTrace struct {
	DNSLookup      time.Duration
	ConnTime       time.Duration
	TCPConnTime    time.Duration
	TLSHandshake   time.Duration
	ServerTime     time.Duration
	ResponseTime   time.Duration
	TotalTime      time.Duration
	ResponseSize   int
	IsConnReused   bool
	IsConnWasIdle  bool
	ConnIdleTime   time.Duration
	RequestAttempt int
	RemoteAddr     string
}

// GetTraceInfo 从响应中提取跟踪信息,要求请求开启了 EnableTrace()
func GetTraceInfo(resp *resty.Response) (*RequestTrace, error) {
	if resp == nil || resp.Request == nil {
		return nil, errors.New("hresty: response or request is nil")
	}
	ti := resp.Request.TraceInfo()
	addr := ""
	if ti.RemoteAddr != nil {
		addr = ti.RemoteAddr.String()
	}
	return &RequestTrace{
		DNSLookup:      ti.DNSLookup,
		ConnTime:       ti.ConnTime,
		TCPConnTime:    ti.TCPConnTime,
		TLSHandshake:   ti.TLSHandshake,
		ServerTime:     ti.ServerTime,
		ResponseTime:   ti.ResponseTime,
		TotalTime:      ti.TotalTime,
		ResponseSize:   len(resp.Body()),
		IsConnReused:   ti.IsConnReused,
		IsConnWasIdle:  ti.IsConnWasIdle,
		ConnIdleTime:   ti.ConnIdleTime,
		RequestAttempt: ti.RequestAttempt,
		RemoteAddr:     addr,
	}, nil
}
