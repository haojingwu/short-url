package bloom

import (
	"testing"

	"github.com/bits-and-blooms/bloom/v3"
)

// 测试布隆过滤器添加和查询
func TestBloomFilterAddExists(t *testing.T) {
	//这里直接在测试中创建过滤器
	filter = NewBloomFilterForTest()

	tests := []struct {
		code      string
		shouldAdd bool
		expected  bool
	}{
		{"abc123", true, true},
		{"xyz789", true, true},
		{"notexist", false, false},
	}

	for _, tt := range tests {
		if tt.shouldAdd {
			filter.Add([]byte(tt.code))
		}
		result := filter.Test([]byte(tt.code))
		if result != tt.expected {
			t.Errorf("code=%s, 期望 %v, 实际 %v", tt.code, tt.expected, result)
		}
	}
}

func NewBloomFilterForTest() *bloom.BloomFilter {
	return bloom.NewWithEstimates(1000, 0.01)
}
