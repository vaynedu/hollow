package hlark

import (
	"context"
	"testing"
)

func TestSendSmsToFeiShu(t *testing.T) {
	webHookUrl := "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxxxxxxxxxx"
	req := map[string]string{
		"text": "这是一条测试消息",
	}
	err := SendSmsToFeiShu(context.Background(), req, webHookUrl)
	if err != nil {
		t.Errorf("SendSmsToFeiShu failed: %v", err)
	}
}
