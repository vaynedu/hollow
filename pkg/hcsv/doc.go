// Package hcsv 提供 CSV 文件读取的简单封装。
//
// 当前只支持读取,且一次性加载到内存,适合配置 / 小数据。
// 处理大文件应直接用标准库 encoding/csv 的流式 Reader。
package hcsv
