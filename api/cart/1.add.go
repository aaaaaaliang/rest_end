package cart

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/state"
	"rest/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddCart 添加购物车
func addCart(c *gin.Context) {
	type Req struct {
		ProductCode string `json:"product_code" binding:"required"`
		ProductNum  int    `json:"product_num" binding:"required,min=1"`
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode, ok := utils.GetUserCode(c)
	if !ok {
		return
	}

	// 1. 获取商品信息 (优先从 Redis 中获取)
	ctx := context.Background()
	product, err := getProductFromCache(ctx, req.ProductCode)
	if err != nil || product == nil {
		response.Success(c, response.QueryFail, fmt.Errorf("商品不存在 或获取商品信息失败: %v", err))
		return
	}

	// 2. 获取用户购物车中该商品的当前数量
	existingCart := model.UserCart{}
	hasCart, err := config.DB.Where("user_code = ? AND product_code = ? AND is_ordered = ?", userCode, req.ProductCode, 0).Get(&existingCart)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查询购物车失败: %v", err))
		return
	}

	// 3. 计算增量
	var increment int
	if hasCart {
		// 如果购物车已经有该商品，计算增量（当前传递的数量 - 购物车中的数量）
		increment = req.ProductNum - existingCart.ProductNum
	} else {
		// 如果购物车中没有该商品，则直接插入
		increment = req.ProductNum
	}

	log.Println("addCart 增量", increment, "productNum", product.Count)

	// 4. 如果库存不足，返回库存不足错误
	if product.Count < int64(increment) {
		response.Success(c, response.ServerError, fmt.Errorf("库存不足, 当前库存: %d", product.Count))
		return
	}

	// 5. 分布式锁，防止并发超卖
	lockKey := fmt.Sprintf(state.RedisCartLock, req.ProductCode)
	if !acquireLock(ctx, lockKey) {
		response.Success(c, response.ServerError, fmt.Errorf("系统繁忙，请稍后再试"))
		return
	}
	defer releaseLock(ctx, lockKey) // 确保锁会被释放

	// 6. 扣减 Redis 中的库存
	stockKey := fmt.Sprintf(state.RedisStockKey, req.ProductCode)
	newStock, err := config.R.DecrBy(ctx, stockKey, int64(increment)).Result()
	if err != nil || newStock < 0 {
		response.Success(c, response.ServerError, fmt.Errorf("库存扣减失败，当前库存: %d", newStock))
		return
	}

	log.Println("当前增量:", newStock)

	// 7. 开启数据库事务
	session := config.DB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 8. 更新购物车记录
	if hasCart {
		// 更新购物车数量
		existingCart.ProductNum = req.ProductNum
		existingCart.TotalPrice = float64(existingCart.ProductNum) * product.Price

		if _, err := session.Where("user_code = ? AND product_code = ?", userCode, req.ProductCode).
			Cols("product_num", "total_price").Update(&existingCart); err != nil {
			rollbackTransaction(session, c, err)
			return
		}
	} else {
		log.Println("123")
		// 插入新商品到购物车
		newCart := model.UserCart{
			UserCode:    userCode,
			ProductCode: req.ProductCode,
			ProductNum:  req.ProductNum,
			TotalPrice:  float64(req.ProductNum) * product.Price,
		}
		affectRow, err := session.Insert(&newCart)
		if affectRow == 0 || err != nil {
			rollbackTransaction(session, c, err)
			return
		}
	}

	// 9. 更新库存
	products := model.Products{
		Count: product.Count - int64(increment),
	}
	affectRow, err := session.Table(model.Products{}).Where("code = ?", req.ProductCode).Cols("count").Update(&products)
	if affectRow == 0 || err != nil {
		rollbackTransaction(session, c, err)
		return
	}

	// 10. 提交事务
	if err := session.Commit(); err != nil {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}

// getProductFromCache 优先从 Redis 获取商品信息，如果不存在则查询数据库
func getProductFromCache(ctx context.Context, productCode string) (*model.Products, error) {
	stockKey := fmt.Sprintf(state.RedisStockKey, productCode)
	priceKey := fmt.Sprintf(state.RedisPriceKey, productCode)

	stockStr, err := config.R.Get(ctx, stockKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil // Redis 中没有缓存
	}
	if err != nil {
		return nil, err
	}

	priceStr, err := config.R.Get(ctx, priceKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// 转换库存和价格为合适的数据类型
	stock, _ := strconv.ParseInt(stockStr, 10, 64)
	price, _ := strconv.ParseFloat(priceStr, 64)

	return &model.Products{
		BasicModel: model.BasicModel{Code: productCode},
		Count:      stock,
		Price:      price,
	}, nil
}
