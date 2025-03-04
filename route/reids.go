package route

import (
	"context"
	"fmt"
	"log"
	"rest/config"
	"rest/model"
	"rest/state"
	"time"
)

// LoadData 初始化时加载产品数据并缓存到 Redis
func loadData() {
	log.Println("Initial loading of product data to Redis...")
	syncDataToRedis()
}

// StartDataSyncTask 启动定时任务，定期同步数据库数据到 Redis
func startDataSyncTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Starting scheduled data sync to Redis...")
		syncDataToRedis()
	}
}

// syncDataToRedis 执行数据同步，将数据库中的商品数据缓存到 Redis
func syncDataToRedis() {
	products, err := fetchProductsFromDB()
	if err != nil {
		log.Println("Failed to fetch products from DB:", err)
		return
	}

	if err := cacheProductsToRedis(products); err != nil {
		log.Println("Failed to cache products to Redis:", err)
	}
}

// fetchProductsFromDB 从数据库加载产品数据
func fetchProductsFromDB() ([]model.Products, error) {
	var products []model.Products
	if err := config.DB.Find(&products); err != nil {
		return nil, fmt.Errorf("fetchProductsFromDB err: %v", err)
	}
	return products, nil
}

// cacheProductsToRedis 将产品库存缓存到 Redis
func cacheProductsToRedis(products []model.Products) error {
	ctx := context.Background()
	pipe := config.R.Pipeline()

	for _, v := range products {
		stockKey := fmt.Sprintf(state.RedisStockKey, v.Code)
		priceKey := fmt.Sprintf(state.RedisPriceKey, v.Code)
		// 批量设置库存和价格缓存
		pipe.Set(ctx, stockKey, v.Count, 0)
		pipe.Set(ctx, priceKey, v.Price, 0)
	}

	// 执行 Pipeline 批量操作
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("cacheProductsToRedis pipeline exec err: %v", err)
	}

	log.Printf("Successfully cached %d products to Redis\n", len(products))
	return nil
}
