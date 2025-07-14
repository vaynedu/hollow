package hresty

import (
	"fmt"
	"testing"
	"time"

	"github.com/avast/retry-go/v4"
)

func TestGet(t *testing.T) {
	err := retry.Do(
		func() error {
			client := NewRestyClient()
			resp, err := client.R().EnableTrace().
				SetQueryParam("name", "value").
				SetQueryParams(map[string]string{
					"vayne": "du",
				}).
				SetHeader("Content-Type", "application/json").
				//Get("https://httpbin.org/get")
				Get("https://www.qq.com/")
			if err != nil {
				t.Fatal(err)
			}
			if resp == nil {
				t.Fatal("resp is nil")
			}

			// 打印请求的url
			t.Log(resp.Request.URL)
			// 打印请求方法
			t.Log(resp.Request.Method)
			// 打印请求头
			t.Log(resp.Request.Header)
			// 打印请求体
			// t.Log(resp.Request.Body)
			// 打印响应头
			t.Log(resp.Header())
			// 打印响应体
			// t.Log(resp.String())
			// 打印trace信息
			// 比较好的一张图，表示时间的关系
			// https://vearne.cc/archives/39953
			PrintTraceInfo(resp)
			traceInfo := resp.Request.TraceInfo()
			t.Logf("Reqeust trace info")
			t.Logf("DNSLookup: %v", traceInfo.DNSLookup)
			t.Logf("ConnTime: %v", traceInfo.ConnTime)
			t.Logf("TCPConnTime: %v", traceInfo.TCPConnTime)
			t.Logf("TLSHandshake: %v", traceInfo.TLSHandshake)
			t.Logf("ServerTime: %v", traceInfo.ServerTime)
			t.Logf("Responsetime: %v", traceInfo.ResponseTime)
			t.Logf("TotalTime: %v", traceInfo.TotalTime)
			t.Logf("Response size: %v", len(resp.Body()))
			t.Logf("IsConnReused: %v", traceInfo.IsConnReused)
			t.Logf("IsConnWasIdle: %v", traceInfo.IsConnWasIdle)
			t.Logf("ConnIdleTime: %v", traceInfo.ConnIdleTime)
			t.Logf("RequestAttempt: %v", traceInfo.RequestAttempt)
			t.Logf("RemoteAddr: %v", traceInfo.RemoteAddr.String())

			ti := traceInfo
			fmt.Println("Request Trace Info:")
			fmt.Println("DNSLookup:", ti.DNSLookup)
			fmt.Println("ConnTime:", ti.ConnTime)
			fmt.Println("TCPConnTime:", ti.TCPConnTime)
			fmt.Println("TLSHandshake:", ti.TLSHandshake)
			fmt.Println("ServerTime:", ti.ServerTime)
			fmt.Println("ResponseTime:", ti.ResponseTime)
			fmt.Println("TotalTime:", ti.TotalTime)
			fmt.Println("IsConnReused:", ti.IsConnReused)
			fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
			fmt.Println("ConnIdleTime:", ti.ConnIdleTime)
			fmt.Println("RequestAttempt:", ti.RequestAttempt)
			fmt.Println("RemoteAddr:", ti.RemoteAddr.String())

			return nil
		},
		retry.Attempts(2),                   // 重试2次 3-1
		retry.Delay(1*time.Second),          // 初始延时 1s
		retry.DelayType(retry.BackOffDelay), // 指数退避策略（0.5s -> 1s）
		retry.LastErrorOnly(true),           // 只返回最后一次错误
	)
	if err != nil {
		t.Fatal(err)
	}

}
