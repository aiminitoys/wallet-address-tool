package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// AddressMatcher 地址匹配器
type AddressMatcher struct {
	config    *Config
	regex     *regexp.Regexp
	attempts  int64
	matched   int64
	startTime time.Time
	mutex     sync.RWMutex
}

// NewAddressMatcher 创建地址匹配器
func NewAddressMatcher(config *Config) (*AddressMatcher, error) {
	matcher := &AddressMatcher{
		config:    config,
		startTime: time.Now(),
	}

	// 预编译正则表达式
	if config.AddressMatching.Rules.Regex != "" {
		regex, err := regexp.Compile(config.AddressMatching.Rules.Regex)
		if err != nil {
			return nil, fmt.Errorf("编译正则表达式失败: %v", err)
		}
		matcher.regex = regex
	}

	return matcher, nil
}

// IsEnabled 检查是否启用地址匹配
func (am *AddressMatcher) IsEnabled() bool {
	return am.config.AddressMatching.Enabled
}

// MatchWallet 检查钱包是否匹配规则
func (am *AddressMatcher) MatchWallet(wallet MultiChainWallet) bool {
	if !am.IsEnabled() {
		return true
	}

	atomic.AddInt64(&am.attempts, 1)

	// 检查各个链的地址
	chains := []struct {
		name    string
		address string
	}{
		{"eth", wallet.EthAddress},
		{"btc", wallet.BtcAddress},
		{"tron", wallet.TronAddress},
		{"bsc", wallet.BscAddress},
		{"polygon", wallet.PolygonAddress},
	}
	switch am.config.AddressMatching.TargetChains[0] {
	case "eth":
		chains = chains[:1]
	case "btc":
		chains = chains[1:2]
	case "tron":
		chains = chains[2:3]
	case "bsc":
		chains = chains[3:4]
	case "polygon":
		chains = chains[4:5]
	case "all":
		// 保持所有链
	default:
		return false // 未指定有效链
	}

	if am.matchesChainAddress(chains[0].address, chains[0].name) {
		atomic.AddInt64(&am.matched, 1)
		return true
	}

	return false
}

// matchesChainAddress 检查特定链的地址是否匹配
func (am *AddressMatcher) matchesChainAddress(address, chainName string) bool {
	// 检查是否是目标链
	if !am.isTargetChain(chainName) {
		return false
	}

	rules := am.config.AddressMatching.Rules

	// 标准化地址（去除0x前缀进行匹配）
	normalizedAddr := normalizeAddress(address)

	// 检查后缀匹配 - 只要有一个匹配就返回true
	if am.checkSuffixes(normalizedAddr, rules.Suffixes) {
		log.Println(2, "后缀匹配成功:", normalizedAddr)
		return true
	}

	// 检查后缀相同匹配 - 如果满足条件就返回true
	if am.checkSuffixesSame(normalizedAddr, rules.SuffixesSame) {
		log.Println(3, "后缀相同连号配成功:", normalizedAddr)
		return true
	}

	// 如果所有规则都为空，则返回true（没有限制）
	if len(rules.Prefixes) == 0 && len(rules.Suffixes) == 0 &&
		len(rules.Contains) == 0 && rules.Regex == "" {
		return true
	}

	return false
}

// isTargetChain 检查是否是目标链
func (am *AddressMatcher) isTargetChain(chainName string) bool {
	targetChains := am.config.AddressMatching.TargetChains
	if len(targetChains) == 0 {
		return true
	}

	for _, target := range targetChains {
		if target == "all" || target == chainName {
			return true
		}
	}
	return false
}

// checkPrefixes 检查前缀匹配
func (am *AddressMatcher) checkPrefixes(address string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}

	for _, prefix := range prefixes {
		if prefix == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(address), strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// checkSuffixes 检查后缀匹配
func (am *AddressMatcher) checkSuffixes(address string, suffixes []string) bool {
	if len(suffixes) == 0 {
		return true
	}

	for _, suffix := range suffixes {
		if suffix == "" {
			continue
		}
		if am.config.AddressMatching.Rules.IgnoreCase {
			if strings.HasSuffix(strings.ToLower(address), strings.ToLower(suffix)) {
				return true
			}
		} else {
			if strings.HasSuffix(address, suffix) {
				return true
			}
		}
	}
	return false
}

// checkSuffixesSame 检查后缀相同匹配
func (am *AddressMatcher) checkSuffixesSame(address string, SuffixesSame int) bool {
	if SuffixesSame == 0 {
		return true
	}

	if len(address) < SuffixesSame {
		return false
	}

	// 获取地址最后一个字符
	lastChar := address[len(address)-1]

	// 检查末尾连续重复字符的数量是否大于等于 SuffixesSame
	count := 0
	for i := len(address) - 1; i >= 0; i-- {
		c := address[i]
		if c == lastChar {
			count++
		} else if am.config.AddressMatching.Rules.IgnoreCase &&
			((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) &&
			(c == lastChar+32 || c == lastChar-32) {
			// 忽略大小写比较
			count++
		} else {
			break
		}
	}

	return count >= SuffixesSame
}

// checkContains 检查包含匹配
func (am *AddressMatcher) checkContains(address string, contains []string) bool {
	if len(contains) == 0 {
		return true
	}

	addressLower := strings.ToLower(address)
	for _, contain := range contains {
		if contain == "" {
			continue
		}
		if strings.Contains(addressLower, strings.ToLower(contain)) {
			return true
		}
	}
	return false
}

// checkRegex 检查正则表达式匹配
func (am *AddressMatcher) checkRegex(address string) bool {
	if am.regex == nil {
		return true
	}
	return am.regex.MatchString(address)
}

// GetStats 获取匹配统计信息
func (am *AddressMatcher) GetStats() (attempts, matched int64, rate float64, duration time.Duration) {
	attempts = atomic.LoadInt64(&am.attempts)
	matched = atomic.LoadInt64(&am.matched)
	duration = time.Since(am.startTime)

	if attempts > 0 {
		rate = float64(matched) / float64(attempts) * 100
	}

	return
}

// PrintStats 打印匹配统计信息
func (am *AddressMatcher) PrintStats() {
	attempts, matched, rate, duration := am.GetStats()

	fmt.Printf("\n📊 地址匹配统计:\n")
	fmt.Printf("尝试次数: %d\n", attempts)
	fmt.Printf("匹配成功: %d\n", matched)
	fmt.Printf("匹配率: %.2f%%\n", rate)
	fmt.Printf("运行时间: %v\n", duration)

	if attempts > 0 {
		fmt.Printf("平均速度: %.2f 次/秒\n", float64(attempts)/duration.Seconds())
	}
}

// ShouldStop 检查是否应该停止生成
func (am *AddressMatcher) ShouldStop() bool {
	if !am.IsEnabled() {
		return false
	}

	maxAttempts := am.config.AddressMatching.MaxAttempts
	if maxAttempts <= 0 {
		return false // 无限制
	}

	return atomic.LoadInt64(&am.attempts) >= int64(maxAttempts)
}

// Reset 重置统计信息
func (am *AddressMatcher) Reset() {
	atomic.StoreInt64(&am.attempts, 0)
	atomic.StoreInt64(&am.matched, 0)
	am.startTime = time.Now()
}

// normalizeAddress 标准化地址（去除0x前缀，转小写）
func normalizeAddress(address string) string {
	if len(address) >= 2 && strings.ToLower(address[:2]) == "0x" {
		return address[2:]
	}
	return address
}

// ValidateMatchingRules 验证匹配规则
func ValidateMatchingRules(rules MatchingRules) error {
	// 验证正则表达式
	if rules.Regex != "" {
		if _, err := regexp.Compile(rules.Regex); err != nil {
			return fmt.Errorf("无效的正则表达式: %v", err)
		}
	}

	// 验证前缀格式
	for _, prefix := range rules.Prefixes {
		if len(prefix) > 40 { // 以太坊地址最长40字符（不含0x）
			return fmt.Errorf("前缀太长: %s", prefix)
		}
	}

	// 验证后缀格式
	for _, suffix := range rules.Suffixes {
		if len(suffix) > 40 {
			return fmt.Errorf("后缀太长: %s", suffix)
		}
	}

	return nil
}
