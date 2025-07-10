package hresty

import (
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
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
	}
}

func PrintTraceInfo(resp *resty.Response) {
	if resp == nil || resp.Request == nil {
		zap.L().Info("Resp or Req is nil")
		return
	}
	traceInfo := resp.Request.TraceInfo()
	zap.L().Info("Reqeust trace info")
	zap.L().Info("DNSLookup:", zap.Duration("DNSLookup", traceInfo.DNSLookup))
	zap.L().Info("ConnTime:", zap.Duration("ConnTime", traceInfo.ConnTime))
	zap.L().Info("TCPConnTime:", zap.Duration("TCPConnTime", traceInfo.TCPConnTime))
	zap.L().Info("TLSHandshake:", zap.Duration("TLSHandshake", traceInfo.TLSHandshake))
	zap.L().Info("ServerTime:", zap.Duration("ServerTime", traceInfo.ServerTime))
	zap.L().Info("Responsetime:", zap.Duration("Response time", traceInfo.ResponseTime))
	zap.L().Info("TotalTime:", zap.Duration("TotalTime", traceInfo.TotalTime))
	zap.L().Info("Response size:", zap.Int("Response size", len(resp.Body())))
	zap.L().Info("IsConnReused:", zap.Bool("IsConnReused", traceInfo.IsConnReused))
	zap.L().Info("IsConnWasIdle:", zap.Bool("IsConnWasIdle", traceInfo.IsConnWasIdle))
	zap.L().Info("ConnIdleTime:", zap.Duration("ConnIdleTime", traceInfo.ConnIdleTime))
	zap.L().Info("RequestAttempt:", zap.Int("RequestAttempt", traceInfo.RequestAttempt))
	zap.L().Info("RemoteAddr:", zap.String("RemoteAddr", traceInfo.RemoteAddr.String()))
}
