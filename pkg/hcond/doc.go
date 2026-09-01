// Package hcond 提供 SQL WHERE 条件构造器,支持原子条件
// (=, !=, >, <, >=, <=, IN)与逻辑组合 (AND, OR)。
//
// 输出 SQL 片段 + 占位参数,可直接喂给 database/sql 或 GORM。
//
// 安全注意:Condition.LHS 直接拼接进 SQL(列名无法参数化),
// 调用方必须确保 LHS 不来自不可信用户输入,以避免 SQL 注入。
// 列名建议用常量或白名单约束。
package hcond
