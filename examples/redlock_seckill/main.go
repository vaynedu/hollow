package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/murmur3"
)

const (
	// 全局库存key
	globalStockKey = "global_stock"
	globalStockNum = 20
	// 库存总数量
	totalStock = 80
	// 库存分段数量
	segments = 10
	// 每段库存数量
	stockPerSegment = totalStock / segments
	// 锁超时时间（毫秒）
	lockExpiry = 200 * time.Millisecond
)

// 初始化Redis客户端
func initRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 无密码
		DB:       0,  // 默认数据库
	})
}

// 初始化库存分段
func initStockSegments(ctx context.Context, redisClient *redis.Client) error {
	pipe := redisClient.TxPipeline()
	for i := 0; i < segments; i++ {
		segmentKey := fmt.Sprintf("stock_segment_%d", i+1)
		pipe.Set(ctx, segmentKey, stockPerSegment, 0)
	}
	// 设置一个全局库存
	pipe.Set(ctx, globalStockKey, globalStockNum, 0)
	_, err := pipe.Exec(ctx)
	return err
}

// 获取库存段ID
func getSegmentID(userID string) int {

	// 去掉异常字符
	cleanID := strings.ToLower(strings.ReplaceAll(userID, "_", ""))
	// MurmurHash 高效、低碰撞、适合非加密场景 	github.com/spaolacci/murmur3
	// xxHash 极速、适合大数据量	github.com/cespare/xxhash
	hash := murmur3.Sum32([]byte(cleanID))
	return int(hash%uint32(segments)) + 1

}

// 尝试秒杀商品
func seckill(ctx context.Context, redisClient *redis.Client, rs *redsync.Redsync, userID, productID string) (bool, error) {
	// 获取用户对应的库存段ID
	segmentID := getSegmentID(userID)
	segmentKey := fmt.Sprintf("stock_segment_%d", segmentID)

	// 获取分布式锁
	mutexName := fmt.Sprintf("stock_lock_%s_%d", productID, segmentID)
	// 创建互斥锁并设置过期时间
	mutex := rs.NewMutex(mutexName, redsync.WithExpiry(lockExpiry))
	// 尝试加锁
	if err := mutex.Lock(); err != nil {
		return false, fmt.Errorf("获取锁失败: %w", err)
	}
	defer mutex.Unlock() // 确保锁最终被释放

	// 检查库存
	stock, err := redisClient.Get(ctx, segmentKey).Int64()
	if err != nil {
		return false, fmt.Errorf("获取库存失败: %w, segmentKey: %s", err, segmentKey)
	}

	if stock <= 0 {
		log.Printf("用户 %s 秒杀失败（库存不足）, segmentKey: %s, stock: %d", userID, segmentKey, stock)
		return false, nil // 库存不足
	}

	// 扣减库存
	newStock, err := redisClient.Decr(ctx, segmentKey).Result()
	if err != nil {
		return false, fmt.Errorf("扣减库存失败: %w", err)
	}

	if newStock < 0 {
		// 库存不足，回滚操作
		redisClient.Incr(ctx, segmentKey)
		return false, nil
	}

	return true, nil // 秒杀成功
}

// selectStockSegment 随机选择库存段
func selectStockSegment(ctx context.Context, redisClient *redis.Client, productID string) (int, error) {
	// 获取所有库存段的剩余库存
	var segmentSlice []int
	var avaliaStockSlice []int64
	for i := 0; i < segments; i++ {
		segmentKey := fmt.Sprintf("stock_segment_%d", i+1)
		stock, err := redisClient.Get(ctx, segmentKey).Int64()
		if err != nil {
			log.Printf("获取库存段 %d 失败: %v", i+1, err)
			continue
		}
		if stock > 0 {
			segmentSlice = append(segmentSlice, i+1)
			avaliaStockSlice = append(avaliaStockSlice, stock)
		}
	}
	if len(segmentSlice) == 0 {
		return 0, fmt.Errorf("没有库存段,所有库存都售罄了") // 所有库存都售罄了
	}
	// 简单实现：按库存比例随机选择（库存越多的段被选中概率越大）
	totalStock := int64(0)
	for _, stock := range avaliaStockSlice {
		totalStock += stock
	}
	rand.Seed(time.Now().UnixNano())
	r := rand.Int63n(totalStock)
	selectedSegment := int64(0)
	for i, stock := range avaliaStockSlice {
		selectedSegment += stock
		if r < selectedSegment {
			return segmentSlice[i], nil
		}
	}

	return segmentSlice[len(segmentSlice)-1], nil
}

// seckillV2 选择动态库存
func seckillV2(ctx context.Context, redisClient *redis.Client, rs *redsync.Redsync, userID, productID string) (bool, error) {
	// 获取用户对应的库存段ID
	segmentID, err := selectStockSegment(ctx, redisClient, productID)
	if err != nil {
		return false, fmt.Errorf("选择库存段失败: %w", err)
	}

	segmentKey := fmt.Sprintf("stock_segment_%d", segmentID)

	// 获取分布式锁
	mutexName := fmt.Sprintf("stock_lock_%s_%d", productID, segmentID)
	// 创建互斥锁并设置过期时间
	mutex := rs.NewMutex(mutexName, redsync.WithExpiry(lockExpiry))
	// 尝试加锁
	if err = mutex.Lock(); err != nil {
		return false, fmt.Errorf("获取锁失败: %w", err)
	}
	defer mutex.Unlock() // 确保锁最终被释放

	// 检查库存
	stock, err := redisClient.Get(ctx, segmentKey).Int64()
	if err != nil {
		return false, fmt.Errorf("获取库存失败: %w, segmentKey: %s", err, segmentKey)
	}

	if stock <= 0 {
		log.Printf("用户 %s 秒杀失败（库存不足）, segmentKey: %s, stock: %d", userID, segmentKey, stock)
		return false, nil // 库存不足
	}

	// 扣减库存
	newStock, err := redisClient.Decr(ctx, segmentKey).Result()
	if err != nil {
		return false, fmt.Errorf("扣减库存失败: %w", err)
	}

	if newStock < 0 {
		// 库存不足，回滚操作
		redisClient.Incr(ctx, segmentKey)
		return false, nil
	}

	return true, nil // 秒杀成功
}

func seckillV3(ctx context.Context, redisClient *redis.Client, rs *redsync.Redsync, userID, productID string) (bool, error) {
	// 获取用户对应的库存段ID
	segmentID, err := selectStockSegment(ctx, redisClient, productID)
	if err != nil {
		return false, fmt.Errorf("选择库存段失败: %w", err)
	}

	segmentKey := fmt.Sprintf("stock_segment_%d", segmentID)

	// 获取分布式锁
	mutexName := fmt.Sprintf("stock_lock_%s_%d", productID, segmentID)
	mutex := rs.NewMutex(mutexName, redsync.WithExpiry(lockExpiry))
	if err = mutex.Lock(); err != nil {
		return false, fmt.Errorf("获取锁失败: %w", err)
	}
	defer mutex.Unlock()
	// 检查分段库存
	stock, err := redisClient.Get(ctx, segmentKey).Int64()
	if err != nil {
		return false, fmt.Errorf("获取库存失败: %w", err)
	}
	if stock <= 0 {
		// 库存不足， 尝试从全局库存调拨
		globalMutex := rs.NewMutex(fmt.Sprintf("global_stock_lock_%s", productID), redsync.WithExpiry(lockExpiry))
		if err = globalMutex.Lock(); err != nil {
			return false, fmt.Errorf("获取全局锁失败: %w", err)
		}
		defer globalMutex.Unlock()

		// 检查全局库存
		globalStock, err := redisClient.Get(ctx, globalStockKey).Int64()
		if err != nil {
			return false, fmt.Errorf("获取全局库存失败: %w", err)
		}
		if globalStock <= 0 {
			log.Printf("用户 %s 秒杀失败（全局库存不足）, segmentKey: %s, stock: %d", userID, segmentKey, stock)
			return false, nil // 全局库存不足
		}

		// 从全局库存调拨部分到当前分段, 这里也要判断全局库存的数量
		transferAmount := int64(math.Min(float64(globalStock), float64(5))) // 每次调拨5个
		redisClient.DecrBy(ctx, globalStockKey, transferAmount)
		redisClient.IncrBy(ctx, segmentKey, transferAmount)
		// 更新当前分段库存
		stock = transferAmount
		log.Printf("用户 %s 从全局库存调拨 %d 个库存到分段 %d", userID, transferAmount, segmentID)
	}
	// 扣减库存
	newStock, err := redisClient.Decr(ctx, segmentKey).Result()
	if err != nil {
		return false, fmt.Errorf("扣减库存失败: %w", err)
	}
	if newStock < 0 {
		// 库存不足，回滚操作
		redisClient.Incr(ctx, segmentKey)
		return false, nil
	}
	return true, nil
}

// transferStock 库存调拨函数
func transferStock(ctx context.Context, redisClient *redis.Client,
	fromSegment, toSegment int, amount int64) error {

	// 使用Lua脚本保证原子性
	script := `
        local from_key = KEYS[1]
        local to_key = KEYS[2]
        local amount = ARGV[1]
        
        local from_stock = tonumber(redis.call('GET', from_key))
        if from_stock >= amount then
            redis.call('DECRBY', from_key, amount)
            redis.call('INCRBY', to_key, amount)
            return 1
        else
            return 0
        end
    `

	keys := []string{
		fmt.Sprintf("stock_segment_%d", fromSegment),
		fmt.Sprintf("stock_segment_%d", toSegment),
	}

	result, err := redisClient.Eval(ctx, script, keys, amount).Int64()
	if err != nil {
		return err
	}

	if result == 0 {
		return fmt.Errorf("调拨失败：源库存不足")
	}

	return nil
}

func main() {
	ctx := context.Background()
	redisClient := initRedisClient()

	// 初始化库存分段
	if err := initStockSegments(ctx, redisClient); err != nil {
		log.Fatalf("初始化库存失败: %v", err)
	}

	// 创建redsync实例用于分布式锁
	pool := goredis.NewPool(redisClient)
	rs := redsync.New(pool)

	// 模拟100个并发用户请求
	var wg sync.WaitGroup
	successCount := 0
	failCount := 0

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()

			// 随机延迟，模拟不同用户请求时间
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

			// success, err := seckill(ctx, redisClient, rs, userID, "product_1")
			//success, err := seckillV2(ctx, redisClient, rs, userID, "product_1")
			success, err := seckillV3(ctx, redisClient, rs, userID, "product_1")

			if err != nil {
				log.Printf("用户 %s 秒杀出错: %v", userID, err)
				return
			}

			if success {
				log.Printf("用户 %s 秒杀成功", userID)
				successCount++
			} else {
				log.Printf("用户 %s 秒杀失败（库存不足）", userID)
				failCount++
			}
		}(fmt.Sprintf("user_%d", i+1))
	}

	wg.Wait()

	// 验证最终库存
	var totalRemaining int64
	for i := 0; i < segments; i++ {
		segmentKey := fmt.Sprintf("stock_segment_%d", i+1)
		remaining, err := redisClient.Get(ctx, segmentKey).Int64()
		if err != nil {
			log.Printf("获取库存段 %d 失败: %v", i+1, err)
			continue
		}
		totalRemaining += remaining
	}

	log.Printf("秒杀结果: 成功 %d 次, 失败 %d 次", successCount, failCount)
	log.Printf("剩余总库存: %d", totalRemaining)
}
