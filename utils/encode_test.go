package utils

import (
	"testing"
)

// 测试1 测试Base62编码
func TestEncodeBase62(t *testing.T) {
	//表格驱动测试
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"数字 0", 0, "0"},
		{"数字 10", 10, "a"},
		{"数字 61", 61, "Z"},
		{"数字 62", 62, "10"},
		{"数字 100", 100, "1C"},
		{"数字 1000", 1000, "g8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeBase62(tt.input)
			if result != tt.expected {
				t.Errorf("encodeBase62(%d) = %s, 期望 %s", tt.input, result, tt.expected)
			}
		})
	}
}

// 测试2 测试短链码生成(幂等性)
func TestGenerateShortCode(t *testing.T) {
	url := "https://www.google.com"

	//同一时间 URL 多次生成, 结果应该相同(幂等性)
	code1 := GenerateShortCode(url)
	code2 := GenerateShortCode(url)
	code3 := GenerateShortCode(url)

	if code1 != code2 || code2 != code3 {
		t.Errorf("幂等性失败: %s, %s, %s", code1, code2, code3)
	}

	//不同URL应该生成不同的code
	code4 := GenerateShortCode("https://www.baidu.com")
	if code1 == code4 {
		t.Errorf("不同URL不应生成相同的 code: %s == %s", code1, code4)
	}

	t.Logf("Google 的 code: %s", code1)
	t.Logf("百度的 code: %s", code4)
}

// 测试3 测试短链码长度
func TestShortCodeLength(t *testing.T) {
	urls := []string{
		"https://www.google.com",
		"https://www.baidu.com",
		"https://github.com",
		"https://stackoverflow.com/questions/ask",
	}

	for _, url := range urls {
		code := GenerateShortCode(url)
		length := len(code)
		if length < 5 || length > 7 {
			t.Errorf("URL: %s 生成的 code %s 长度 %d 不在5-7之间", url, code, length)
		} else {
			t.Logf("URL: %s -> code: %s(长度: %d)", url, code, length)
		}
	}
}

// 测试4 测试冲突重试机制
func TestGenerateShortCodeWithRetry(t *testing.T) {
	//模拟一个总是返回trye的检查器(模拟全部冲突)
	alwaysExists := func(code string) bool {
		return true
	}
	//第一次调用 应该触发重试 最终返回兜底的tmp值
	code := GenerateShortCodeWithRetry("https://test.com", alwaysExists)

	//因为检查其总是返回true,所以最终会走到兜底逻辑, 返回以"tmp"开头的code
	if len(code) < 3 || code[:3] != "tmp" {
		t.Errorf("冲突重试失败: 期望以 'tmp'开头, 实际得到 %s", code)
	} else {
		t.Logf("冲突重试成功, 返回兜底code: %s", code)
	}

	//模拟一个从不冲突的检查器
	neverExists := func(code string) bool {
		return false
	}
	code2 := GenerateShortCodeWithRetry("https://test2.com", neverExists)
	if code2 == "" {
		t.Error("无冲突时不应该返回空字符串")
	} else {
		t.Logf("无冲突时正常返回: %s", code2)
	}
}

// 测试5 基准测试(性能测试)
func BenchmarkGenerateShortCode(b *testing.B) {
	url := "https://www/google/com"
	for i := 0; i < b.N; i++ {
		GenerateShortCode(url)
	}
}

func BenchmarkEncodeBase62(b *testing.B) {
	num := uint64(123456789)
	for i := 0; i < b.N; i++ {
		encodeBase62(num)
	}
}
