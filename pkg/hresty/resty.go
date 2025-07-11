package hresty

import (
	"errors"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/vaynedu/hollow/internal/logger"
	"go.uber.org/zap"
)

func NewRestyClient() *resty.Client {
	return resty.NewWithClient(newClient())
}

func newClient() *http.Client {
	return &http.Client{
		Transport: NewTransport(),
		Timeout:   10 * time.Second,
	}
}

func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		// DialContext 它决定了 HTTP 客户端如何拨号并建立底层的 TCP 连接
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true, // 是否启用 IPv4 和 IPv6 双栈支持（true 表示同时支持）
		}).DialContext,
		ForceAttemptHTTP2:     false,                      // 是否强制尝试 HTTP/2
		MaxConnsPerHost:       runtime.GOMAXPROCS(0) * 64, // 每个主机的最大连接数，0 表示不限制
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) * 64, // 每个主机的最大空闲连接数
		IdleConnTimeout:       30 * time.Second,           // 空闲连接超时时间
		TLSHandshakeTimeout:   3 * time.Second,            // TLS 握手超时时间
		ExpectContinueTimeout: 1 * time.Second,            // 100-continue 握手超时时间
		DisableKeepAlives:     false,                      // 是否开启长链接
	}
}

// RequestTrace 定义请求跟踪信息的结构化数据
type RequestTrace struct {
	DNSLookup      time.Duration // DNS查询时间
	ConnTime       time.Duration // 连接建立时间(含TCP+TLS)
	TCPConnTime    time.Duration // TCP连接时间
	TLSHandshake   time.Duration // TLS握手时间
	ServerTime     time.Duration // 服务器处理时间
	ResponseTime   time.Duration // 响应总时间
	TotalTime      time.Duration // 请求总耗时
	ResponseSize   int           // 响应体大小
	IsConnReused   bool          // 连接是否复用
	IsConnWasIdle  bool          // 连接是否为空闲连接
	ConnIdleTime   time.Duration // 连接空闲时间
	RequestAttempt int           // 请求尝试次数
	RemoteAddr     string        // 远程服务器地址
}

// GetTraceInfo 从响应中提取跟踪信息并返回结构化数据
func GetTraceInfo(resp *resty.Response) (*RequestTrace, error) {
	if resp == nil || resp.Request == nil {
		return nil, errors.New("resp or request is nil")
	}

	traceInfo := resp.Request.TraceInfo()
	return &RequestTrace{
		DNSLookup:      traceInfo.DNSLookup,
		ConnTime:       traceInfo.ConnTime,
		TCPConnTime:    traceInfo.TCPConnTime,
		TLSHandshake:   traceInfo.TLSHandshake,
		ServerTime:     traceInfo.ServerTime,
		ResponseTime:   traceInfo.ResponseTime,
		TotalTime:      traceInfo.TotalTime,
		ResponseSize:   len(resp.Body()),
		IsConnReused:   traceInfo.IsConnReused,
		IsConnWasIdle:  traceInfo.IsConnWasIdle,
		ConnIdleTime:   traceInfo.ConnIdleTime,
		RequestAttempt: traceInfo.RequestAttempt,
		RemoteAddr:     traceInfo.RemoteAddr.String(),
	}, nil
}

// PrintTraceInfo 打印请求跟踪信息(保持向后兼容)
func PrintTraceInfo(resp *resty.Response) {
	trace, err := GetTraceInfo(resp)
	if err != nil {
		logger.GetLogger().Error("获取跟踪信息失败", zap.Error(err))
		return
	}
	PrintStructuredTrace(trace)
}

// PrintStructuredTrace 打印结构化的请求跟踪信息
func PrintStructuredTrace(trace *RequestTrace) {
	logger.GetLogger().Info("请求跟踪信息",
		zap.Duration("dns查询时间", trace.DNSLookup),
		zap.Duration("连接建立时间", trace.ConnTime),
		zap.Duration("tcp连接时间", trace.TCPConnTime),
		zap.Duration("tls握手时间", trace.TLSHandshake),
		zap.Duration("服务器处理时间", trace.ServerTime),
		zap.Duration("响应总时间", trace.ResponseTime),
		zap.Duration("请求总耗时", trace.TotalTime),
		zap.Int("响应体大小(字节)", trace.ResponseSize),
		zap.Bool("连接是否复用", trace.IsConnReused),
		zap.Bool("是否空闲连接", trace.IsConnWasIdle),
		zap.Duration("连接空闲时间", trace.ConnIdleTime),
		zap.Int("请求尝试次数", trace.RequestAttempt),
		zap.String("远程服务器地址", trace.RemoteAddr),
	)
}
