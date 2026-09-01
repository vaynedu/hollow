package hlark

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type feiShuTextContent struct {
	Text string `json:"text"`
}

type feiShuRequest struct {
	Timestamp string            `json:"timestamp"`
	Sign      string            `json:"sign"`
	MsgType   string            `json:"msg_type"`
	Content   feiShuTextContent `json:"content"`
}

type feiShuResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SendTextToFeiShu 通过自定义机器人 webhook 发送文本消息
// url:    open.feishu.cn/open-apis/bot/v2/hook/{token}
// secret: 机器人安全设置中的"签名校验"密钥(若未开启,可传空,GenSign 会跳过)
func SendTextToFeiShu(ctx context.Context, url, secret, text string) error {
	ts := time.Now().Unix()
	sign, err := GenSign(ts, secret)
	if err != nil {
		return fmt.Errorf("hlark: gen sign: %w", err)
	}

	payload := feiShuRequest{
		Timestamp: fmt.Sprintf("%d", ts),
		Sign:      sign,
		MsgType:   "text",
		Content:   feiShuTextContent{Text: text},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("hlark: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("hlark: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("hlark: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hlark: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var parsed feiShuResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("hlark: decode response: %w, body=%s", err, string(respBody))
	}
	if parsed.Code != 0 {
		return fmt.Errorf("hlark: code=%d msg=%s", parsed.Code, parsed.Msg)
	}
	return nil
}

// GenSign 飞书自定义机器人签名算法:HMAC-SHA256(key=timestamp+\n+secret, data=""),再 base64
// secret 为空时返回空串(适配未开启签名校验的机器人)
func GenSign(timestamp int64, secret string) (string, error) {
	if secret == "" {
		return "", nil
	}
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	if _, err := h.Write(nil); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
