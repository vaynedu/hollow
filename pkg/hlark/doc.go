// Package hlark 提供飞书机器人 webhook 发送辅助。
//
// 签名算法用飞书官方:
//   stringToSign := timestamp + "\n" + secret
//   HMAC-SHA256(key=stringToSign, data="") 然后 base64 编码
//
// 当前支持文本消息(text)。富文本 / 卡片消息后续扩展。
package hlark
