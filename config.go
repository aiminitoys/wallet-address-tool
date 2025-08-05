package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"

	"gopkg.in/yaml.v2"
)

// Config 主配置结构
type Config struct {
	Generator       ConfigGeneratorConfig `yaml:"generator"`
	WorkerPool      WorkerPoolConfig      `yaml:"worker_pool"`
	AddressMatching AddressMatchingConfig `yaml:"address_matching"`
	Performance     PerformanceConfig     `yaml:"performance"`
	Output          OutputConfig          `yaml:"output"`
}

// ConfigGeneratorConfig 生成器配置
type ConfigGeneratorConfig struct {
	DefaultCount      int  `yaml:"default_count"`
	UseMnemonic       bool `yaml:"use_mnemonic"`
	BatchDefaultCount int  `yaml:"batch_default_count"`
}

// WorkerPoolConfig 协程池配置
type WorkerPoolConfig struct {
	AutoDetect    bool `yaml:"auto_detect"`
	ManualCount   int  `yaml:"manual_count"`
	CPUMultiplier int  `yaml:"cpu_multiplier"`
	MinWorkers    int  `yaml:"min_workers"`
	MaxWorkers    int  `yaml:"max_workers"`
}

// AddressMatchingConfig 地址匹配配置
type AddressMatchingConfig struct {
	Enabled      bool          `yaml:"enabled"`
	Rules        MatchingRules `yaml:"rules"`
	TargetChains []string      `yaml:"target_chains"`
	MaxAttempts  int           `yaml:"max_attempts"`
	MaxMatch     int           `yaml:"max_match"`
}

// MatchingRules 匹配规则
type MatchingRules struct {
	IgnoreCase   bool     `yaml:"ignore_case"`
	Prefixes     []string `yaml:"prefixes"`
	Suffixes     []string `yaml:"suffixes"`
	SuffixesSame int      `yaml:"suffix_same"`
	Contains     []string `yaml:"contains"`
	Regex        string   `yaml:"regex"`
}

// PerformanceConfig 性能测试配置
type PerformanceConfig struct {
	AutoBenchmark bool        `yaml:"auto_benchmark"`
	TestSamples   int         `yaml:"test_samples"`
	WorkerRange   WorkerRange `yaml:"worker_range"`
}

// WorkerRange 协程数测试范围
type WorkerRange struct {
	Min  int `yaml:"min"`
	Max  int `yaml:"max"`
	Step int `yaml:"step"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Verbose      bool   `yaml:"verbose"`
	PreviewCount int    `yaml:"preview_count"`
	SaveToFile   bool   `yaml:"save_to_file"`
	OutputFile   string `yaml:"output_file"`
}

// 全局配置实例
var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  配置文件 %s 不存在，创建默认配置...\n", configPath)
		config := getDefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("创建默认配置失败: %v", err)
		}
		return config, nil
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return &config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		Generator: ConfigGeneratorConfig{
			DefaultCount:      1,
			UseMnemonic:       false,
			BatchDefaultCount: 100,
		},
		WorkerPool: WorkerPoolConfig{
			AutoDetect:    true,
			ManualCount:   4,
			CPUMultiplier: 2,
			MinWorkers:    1,
			MaxWorkers:    32,
		},
		AddressMatching: AddressMatchingConfig{
			Enabled: false,
			Rules: MatchingRules{
				Prefixes: []string{},
				Suffixes: []string{},
				Contains: []string{},
				Regex:    "",
			},
			TargetChains: []string{"eth"},
			MaxAttempts:  10000,
		},
		Performance: PerformanceConfig{
			AutoBenchmark: false,
			TestSamples:   1000,
			WorkerRange: WorkerRange{
				Min:  1,
				Max:  16,
				Step: 1,
			},
		},
		Output: OutputConfig{
			Verbose:      true,
			PreviewCount: 5,
			SaveToFile:   false,
			OutputFile:   "wallets.json",
		},
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证协程池配置
	if config.WorkerPool.MinWorkers < 1 {
		return fmt.Errorf("最小协程数不能小于1")
	}
	if config.WorkerPool.MaxWorkers < config.WorkerPool.MinWorkers {
		return fmt.Errorf("最大协程数不能小于最小协程数")
	}
	if config.WorkerPool.CPUMultiplier < 1 {
		return fmt.Errorf("CPU倍数不能小于1")
	}

	// 验证地址匹配配置
	if config.AddressMatching.Enabled {
		if config.AddressMatching.Rules.Regex != "" {
			if _, err := regexp.Compile(config.AddressMatching.Rules.Regex); err != nil {
				return fmt.Errorf("正则表达式无效: %v", err)
			}
		}
		if config.AddressMatching.MaxAttempts < 0 {
			return fmt.Errorf("最大尝试次数不能为负数")
		}
	}

	// 验证性能测试配置
	if config.Performance.WorkerRange.Min < 1 {
		return fmt.Errorf("性能测试最小协程数不能小于1")
	}
	if config.Performance.WorkerRange.Max < config.Performance.WorkerRange.Min {
		return fmt.Errorf("性能测试最大协程数不能小于最小协程数")
	}
	if config.Performance.WorkerRange.Step < 1 {
		return fmt.Errorf("性能测试步长不能小于1")
	}

	return nil
}

// GetOptimalWorkerCount 获取最优协程数
func (c *Config) GetOptimalWorkerCount() int {
	if !c.WorkerPool.AutoDetect {
		return c.WorkerPool.ManualCount
	}

	cpuCount := runtime.NumCPU()
	workerCount := cpuCount * c.WorkerPool.CPUMultiplier

	// 限制在最小最大范围内
	if workerCount < c.WorkerPool.MinWorkers {
		workerCount = c.WorkerPool.MinWorkers
	}
	if workerCount > c.WorkerPool.MaxWorkers {
		workerCount = c.WorkerPool.MaxWorkers
	}

	return workerCount
}

// MatchesAddress 检查地址是否匹配规则
func (c *Config) MatchesAddress(address string, chain string) bool {
	if !c.AddressMatching.Enabled {
		return true
	}

	// 检查目标链
	if len(c.AddressMatching.TargetChains) > 0 {
		found := false
		for _, targetChain := range c.AddressMatching.TargetChains {
			if targetChain == "all" || targetChain == chain {
				found = true
				break
			}
		}
		if !found {
			return true // 不是目标链，直接通过
		}
	}

	rules := c.AddressMatching.Rules

	// 检查前缀
	for _, prefix := range rules.Prefixes {
		if len(prefix) > 0 && !hasPrefix(address, prefix) {
			return false
		}
	}

	// 检查后缀
	for _, suffix := range rules.Suffixes {
		if len(suffix) > 0 && !hasSuffix(address, suffix) {
			return false
		}
	}

	// 检查包含
	for _, contain := range rules.Contains {
		if len(contain) > 0 && !contains(address, contain) {
			return false
		}
	}

	// 检查正则表达式
	if rules.Regex != "" {
		matched, err := regexp.MatchString(rules.Regex, address)
		if err != nil || !matched {
			return false
		}
	}

	return true
}

// 辅助函数
func hasPrefix(s, prefix string) bool {
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:] // 去掉0x前缀
	}
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:] // 去掉0x前缀
	}
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func contains(s, substr string) bool {
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:] // 去掉0x前缀
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
