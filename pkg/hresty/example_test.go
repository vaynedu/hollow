package hresty

import "fmt"

// Example 不带 // Output:,因为需要外网,仅展示调用形态。
func ExampleNewRestyClient() {
	client := NewRestyClient()
	resp, err := client.R().EnableTrace().Get("https://example.com")
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}

	trace, err := GetTraceInfo(resp)
	if err == nil {
		fmt.Println("DNS lookup:", trace.DNSLookup)
		fmt.Println("Total time:", trace.TotalTime)
	}
}
