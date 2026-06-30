package bloom

import (
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"

	"short-url/database"
	"short-url/logger"
	"short-url/models"
)

var (
	filter *bloom.BloomFilter
	once   sync.Once
)

// InitBloomFilter 初始化布隆过滤器 并从数据库预热加载已有code
func InitBloomFilter() {
	once.Do(func() {
		//创建布隆过滤器
		//参数包含预估元素数量: 10万 期望误判率: 0.01 这两个参数决定位数组大小和哈希函数数量
		filter = bloom.NewWithEstimates(100000, 0.01)
		logger.Info("布隆过滤器创建成功",
			zap.Uint("位数组大小", filter.Cap()),
			zap.Uint("哈希函数数量", filter.K()),
		)
		//从数据库预热加载所有已有code
		warmupFromDB()

	})
}

// warmuoFromDB 从数据库加载所有code到布隆过滤器
func warmupFromDB() {
	db := database.DB

	//查询所有code
	var codes []string
	if err := db.Model(&models.URL{}).Pluck("code", &codes).Error; err != nil {
		logger.Error("预热布隆过滤器失败", zap.Error(err))
		return
	}

	//将所有code添加到布隆过滤器
	for _, code := range codes {
		filter.Add([]byte(code))
	}

	logger.Info("布隆过滤器预热完成", zap.Int("加载数量", len(codes)))
}

// Add 添加code到布隆过滤器
func Add(code string) {
	if filter != nil {
		filter.Add([]byte(code))
	}
}

// Exists检查code是否可能存在
// 返回true表示可能存在,返回false表示一定不存在
func Exists(code string) bool {
	if filter == nil {
		return true
	}
	return filter.Test([]byte(code))
}
