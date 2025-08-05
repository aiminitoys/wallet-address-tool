package main

import (
	"fmt"
	"sync"
	"time"
)

// MatchingResult 匹配结果
type MatchingResult struct {
	Wallets  []MultiChainWallet
	Duration time.Duration
}

// MatchingService 地址匹配服务
type MatchingService struct {
	config    *Config
	generator *WalletGenerator
	matcher   *AddressMatcher
}

// NewMatchingService 创建地址匹配服务
func NewMatchingService(config *Config) *MatchingService {
	generator := NewWalletGenerator(config)
	matcher, _ := NewAddressMatcher(config)

	return &MatchingService{
		config:    config,
		generator: generator,
		matcher:   matcher,
	}
}

// RunMatching 运行地址匹配
func (ms *MatchingService) RunMatching() *MatchingResult {
	fmt.Println("🎯 地址匹配模式")
	fmt.Printf("匹配规则: 前缀=%v, 后缀=%v, 包含=%v\n",
		ms.config.AddressMatching.Rules.Prefixes,
		ms.config.AddressMatching.Rules.Suffixes,
		ms.config.AddressMatching.Rules.Contains)

	if ms.config.AddressMatching.Rules.Regex != "" {
		fmt.Printf("正则表达式: %s\n", ms.config.AddressMatching.Rules.Regex)
	}

	fmt.Printf("目标链: %v\n", ms.config.AddressMatching.TargetChains)
	fmt.Printf("最大尝试次数: %d\n", ms.config.AddressMatching.MaxAttempts)

	var matchedWallets []MultiChainWallet
	workerCount := ms.config.GetOptimalWorkerCount()
	// 创建输出配置器
	var output *OutputConfig
	if ms.config != nil {
		output = &ms.config.Output
	}

	fmt.Printf("\n开始匹配，使用 %d 个协程...\n", workerCount)
	start := time.Now()

	// 使用协程池进行匹配
	walletChan := make(chan MultiChainWallet, 100)
	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	// 启动工作协程
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					if ms.matcher.ShouldStop() {
						return
					}

					wallet, err := ms.generator.GenerateRandomWallet()
					if err != nil {
						continue
					}

					if ms.matcher.MatchWallet(wallet) {
						select {
						case walletChan <- wallet:
						case <-stopChan:
							return
						}
					}
				}
			}
		}()
	}

	// 收集匹配的钱包
	var mu sync.Mutex
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for wallet := range walletChan {
			mu.Lock()
			matchedWallets = append(matchedWallets, wallet)
			count := len(matchedWallets)
			mu.Unlock()

			fmt.Printf("✅ 找到匹配地址! (#%d)\n", count)
			// 储存地址
			if output != nil && output.SaveToFile && output.OutputFile != "" {
				// 如果启用了输出配置，保存到文件
				if ms.matcher != nil && ms.matcher.IsEnabled() {
					err := saveWalletsToFile(ms.config.Generator.UseMnemonic, ms.matcher.config.AddressMatching.TargetChains[0], wallet, output.OutputFile)
					if err != nil {
						fmt.Printf("保存钱包到文件失败: %v\n", err)
					}
				} else {
					err := saveWalletsToFile(ms.config.Generator.UseMnemonic, "all", wallet, output.OutputFile)
					if err != nil {
						fmt.Printf("保存钱包到文件失败: %v\n", err)
					}
				}
			}
			PrintWalletSimple(wallet)

			// 如果找到足够的匹配，停止搜索
			if count >= ms.config.AddressMatching.MaxMatch && ms.config.AddressMatching.MaxMatch > 0 {
				close(stopChan)
				return
			}
		}
	}()

	// 定期显示统计信息
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if ms.matcher.ShouldStop() {
					return
				}
				ms.matcher.PrintStats()
			case <-doneChan:
				return
			}
		}
	}()

	// 等待停止条件
	for !ms.matcher.ShouldStop() {
		mu.Lock()
		count := len(matchedWallets)
		mu.Unlock()

		if count >= ms.config.AddressMatching.MaxMatch && ms.config.AddressMatching.MaxMatch > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 停止所有协程
	close(stopChan)
	wg.Wait() // 等待所有 worker goroutine 结束
	close(walletChan)
	<-doneChan // 等待收集器 goroutine 结束

	duration := time.Since(start)

	return &MatchingResult{
		Wallets:  matchedWallets,
		Duration: duration,
	}
}
