package hlark

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"time"
)

type FeiShuMessageContent struct {
	Text string `json:"text"`
}

type FeiShuMessage struct {
	Timestamp string               `json:"timestamp"`
	Sign      string               `json:"sign"`
	MsgType   string               `json:"msg_type"`
	Content   FeiShuMessageContent `json:"content"`
}

func SendSmsToFeiShu(ctx context.Context, req interface{}, url string) error {
	timeStamp := time.Now().Unix()
	sg, err := genSign(timeStamp)
	if err != nil {
		return err
	}
	bts, _ := json.Marshal(req)
	msg := FeiShuMessage{
		Sign:      sg,
		Timestamp: fmt.Sprintf("%v", timeStamp),
		MsgType:   "text",
		Content: FeiShuMessageContent{
			Text: string(bts),
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	feiShuReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	feiShuReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(feiShuReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func genSign(timeStamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timeStamp) + "\n" + "xxxxxxxxxxxxxxxxxxxxxxxxxxx"
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
