package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// OrderExpiryHandler 处理订单过期事件
type OrderExpiryHandler struct {
	rdb        *redis.Client
	orderQueue chan string // 处理订单的队列
}

// NewOrderExpiryHandler 创建订单过期处理器
func NewOrderExpiryHandler(rdb *redis.Client) *OrderExpiryHandler {
	return &OrderExpiryHandler{
		rdb:        rdb,
		orderQueue: make(chan string, 1000), // 带缓冲的通道
	}
}

// Start 启动事件监听和处理
func (h *OrderExpiryHandler) Start(ctx context.Context) {
	// 启动事件监听协程
	go h.listenExpiryEvents(ctx)

	// 启动多个工作协程处理订单
	const workerCount = 5
	for i := 0; i < workerCount; i++ {
		go h.processOrders(ctx, i)
	}
}

// listenExpiryEvents 监听Redis过期事件
func (h *OrderExpiryHandler) listenExpiryEvents(ctx context.Context) {
	// 订阅过期事件频道
	pubsub := h.rdb.Subscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	// 接收消息
	ch := pubsub.Channel()
	log.Println("开始监听Redis过期事件...")

	for msg := range ch {
		// 过滤订单相关的过期事件（假设订单键格式为 order:{orderID}）
		if strings.HasPrefix(msg.Payload, "order_expiry:") {
			orderID := strings.TrimPrefix(msg.Payload, "order_expiry:")
			log.Printf("收到订单过期事件: %s(来自监控Key %s)", orderID, msg.Payload)

			// 将订单ID放入处理队列
			select {
			case h.orderQueue <- orderID:
				log.Printf("订单 %s 已加入处理队列", orderID)
			case <-ctx.Done():
				return
			}
		}
	}
}

// processOrders 处理过期订单
func (h *OrderExpiryHandler) processOrders(ctx context.Context, workerID int) {
	log.Printf("工作协程 %d 开始处理订单...", workerID)

	for {
		select {
		case orderID := <-h.orderQueue:
			// 处理过期订单
			err := h.handleExpiredOrder(ctx, orderID)
			if err != nil {
				log.Printf("处理订单 %s 失败: %v", orderID, err)

				// 可以在这里实现重试逻辑（如放入延迟队列）
			}

		case <-ctx.Done():
			log.Printf("工作协程 %d 已停止", workerID)
			return
		}
	}
}

// handleExpiredOrder 处理单个过期订单
func (h *OrderExpiryHandler) handleExpiredOrder(ctx context.Context, orderID string) error {
	// 1. 检查订单状态（防止重复处理或状态已变更）
	orderStatus, err := h.rdb.Get(ctx, fmt.Sprintf("order_status:%s", orderID)).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("订单 %s 状态不存在，可能已被处理", orderID)
			return nil
		}
		return fmt.Errorf("获取订单状态失败: %v", err)
	}

	// 2. 只处理未支付的订单
	if orderStatus != "pending" {
		log.Printf("订单 %s 状态不是pending,当前状态为%s,跳过处理", orderID, orderStatus)
		return nil
	}

	// 3. 开始事务处理
	err = h.rdb.Watch(ctx, func(tx *redis.Tx) error {
		// 再次检查订单状态（CAS乐观锁）
		status, err := tx.Get(ctx, fmt.Sprintf("order_status:%s", orderID)).Result()
		if err != nil {
			return err
		}

		if status != "pending" {
			return nil // 状态已变更，跳过处理
		}

		// 获取订单信息
		orderKey := fmt.Sprintf("order:%s", orderID)
		orderData, err := tx.HGetAll(ctx, orderKey).Result()
		if err != nil {
			return err
		}

		// 提取商品ID和数量
		productID, ok := orderData["product_id"]
		if !ok || productID == "" {
			return fmt.Errorf("订单 %s 缺少product_id", orderID)
		}
		quantityStr, ok := orderData["quantity"]
		if !ok || quantityStr == "" {
			return fmt.Errorf("订单 %s 缺少quantity", orderID)
		}
		quantity, err := stringToInt(quantityStr)
		if err != nil {
			return fmt.Errorf("订单 %s 的quantity格式错误(值：%s):%v", orderID, quantityStr, err)
		}

		// 开始事务操作
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			// a. 更新订单状态为已取消
			pipe.Set(ctx, fmt.Sprintf("order_status:%s", orderID), "cancelled", 0)

			// b. 释放库存（增加可用库存，减少锁定库存）
			pipe.IncrBy(ctx, fmt.Sprintf("stock:%s", productID), int64(quantity))
			pipe.DecrBy(ctx, fmt.Sprintf("locked_stock:%s", productID), int64(quantity))

			// c. 记录库存释放日志
			releaseLog := map[string]interface{}{
				"order_id":   orderID,
				"product_id": productID,
				"quantity":   quantity,
				"timestamp":  time.Now().Unix(),
				"reason":     "order_expired",
			}

			logJSON, _ := json.Marshal(releaseLog)
			pipe.RPush(ctx, "stock_release_logs", string(logJSON))

			return nil
		})

		return err
	}, fmt.Sprintf("order_status:%s", orderID))

	if err != nil {
		return fmt.Errorf("处理订单 %s 事务失败: %v", orderID, err)
	}

	log.Printf("订单 %s 已成功处理，库存已释放", orderID)
	return nil
}

// stringToInt 辅助函数：字符串转整数
func stringToInt(s string) (int, error) {
	if s == "" { // 避免空字符串
		return 0, fmt.Errorf("空字符串无法转换为整数")
	}
	return strconv.Atoi(s) // 直接解析字符串为整数（支持"1"、"100"等格式）
}

func main() {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "10.10.187.21:6379",
		Password: "123456", // 无密码
		DB:       0,        // 默认数据库
	})

	// 测试连接
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("连接Redis失败: %v", err)
	}
	log.Printf("Redis连接成功: %s", pong)

	// 创建订单过期处理器并启动
	handler := NewOrderExpiryHandler(rdb)
	handler.Start(ctx)

	// 模拟创建一些带过期时间的订单
	go simulateOrderCreation(ctx, rdb)

	// 保持主程序运行
	select {}
}

// simulateOrderCreation 模拟创建订单（仅用于测试）
func simulateOrderCreation(ctx context.Context, rdb *redis.Client) {
	for i := 0; i < 10; i++ {
		orderID := fmt.Sprintf("test_%d", i)
		productID := "iphone14"

		// 创建订单
		err := rdb.HSet(ctx, fmt.Sprintf("order:%s", orderID), map[string]interface{}{
			"order_id":    orderID,
			"user_id":     fmt.Sprintf("user_%d", i),
			"product_id":  productID,
			"quantity":    strconv.Itoa(1),
			"amount":      "9999",
			"create_time": strconv.FormatInt(time.Now().Unix(), 10),
		}).Err()

		if err != nil {
			log.Printf("创建订单 %s 失败: %v", orderID, err)
			continue
		}

		// 2. 设置订单状态（不设置过期）
		rdb.Set(ctx, fmt.Sprintf("order_status:%s", orderID), "pending", 0)

		// 3. 单独的过期监控Key（order_expiry:%s 设置15秒过期）
		expiryKey := fmt.Sprintf("order_expiry:%s", orderID)
		rdb.Set(ctx, expiryKey, "1", 15*time.Second) // 仅用于触发过期事件

		log.Printf("创建订单 %s,监控Key %s 将在15秒后过期", orderID, expiryKey)

		// 模拟用户间隔创建订单
		time.Sleep(2 * time.Second)
	}
}
