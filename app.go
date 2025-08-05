package main

import (
	"fmt"
	"os"
)

// App 应用程序结构
type App struct {
	config *Config
}

// NewApp 创建新的应用实例
func NewApp(config *Config) *App {
	return &App{
		config: config,
	}
}

// Run 运行应用程序
func (app *App) Run() {
	fmt.Println("\n选择生成模式:")
	fmt.Println("1. 单个钱包 (随机)")
	fmt.Println("2. 单个钱包 (助记词)")
	fmt.Println("3. 批量生成 (并发模式)")
	fmt.Println("4. 从指定助记词派生多个地址")
	fmt.Println("5. 地址匹配模式")
	fmt.Println("6. 性能基准测试")

	var choice int
	fmt.Print("请选择 (1-6): ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		app.generateSingleWallet(false)
	case 2:
		app.generateSingleWallet(true)
	case 3:
		app.generateBatchWallets()
	case 4:
		app.deriveFromMnemonic()
	case 5:
		app.runAddressMatching()
	case 6:
		app.runPerformanceBenchmark()
	default:
		fmt.Println("无效选择")
	}
}

// generateSingleWallet 生成单个钱包
func (app *App) generateSingleWallet(useMnemonic bool) {
	generator := NewWalletGenerator(app.config)

	config := GeneratorConfig{
		Count:       1,
		UseMnemonic: useMnemonic,
	}

	wallets := generator.GenerateWallets(config)
	if len(wallets) > 0 {
		PrintWallet(wallets[0])
	}
}

// generateBatchWallets 批量生成钱包
func (app *App) generateBatchWallets() {
	generator := NewWalletGenerator(app.config)

	var count, workers int
	var useMnemonic string

	fmt.Printf("生成数量 (默认: %d): ", app.config.Generator.BatchDefaultCount)
	if _, err := fmt.Scanln(&count); err != nil {
		count = app.config.Generator.BatchDefaultCount
	}

	optimalWorkers := app.config.GetOptimalWorkerCount()
	fmt.Printf("并发协程数 (建议: %d): ", optimalWorkers)
	if _, err := fmt.Scanln(&workers); err != nil {
		workers = optimalWorkers
	}

	fmt.Print("使用助记词? (y/n): ")
	fmt.Scanln(&useMnemonic)

	config := GeneratorConfig{
		Count:          count,
		UseMnemonic:    useMnemonic == "y" || useMnemonic == "Y",
		ConcurrentMode: true,
		WorkerCount:    workers,
	}

	result := generator.GenerateBatch(config)

	fmt.Printf("\n✅ 成功生成 %d 个钱包，耗时: %v\n", len(result.Wallets), result.Duration)
	fmt.Printf("平均每个钱包耗时: %v\n", result.AvgTime)
	fmt.Printf("生成速度: %.2f 钱包/秒\n", result.WalletsPerSec)

	// 显示前N个钱包
	previewCount := app.config.Output.PreviewCount
	for i, wallet := range result.Wallets {
		if i >= previewCount {
			fmt.Printf("... 还有 %d 个钱包\n", len(result.Wallets)-previewCount)
			break
		}
		PrintWalletSimple(wallet)
	}
}

// 存储钱包到文件
func saveWalletsToFile(isMnemonic bool, chain string, multiChainWallet MultiChainWallet, s string) error {
	file, err := os.OpenFile(s, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		// 如果文件不存在，创建新文件
		file, err = os.Create(s)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
	}
	// 文件已存在,逐行写入
	defer file.Close()
	address := ""
	switch chain {
	case "eth":
	case "polygon":
	case "bsc":
		address = multiChainWallet.EthAddress
	case "btc":
		address = multiChainWallet.BtcAddress
	case "tron":
		address = multiChainWallet.TronAddress
	case "all":
		address = multiChainWallet.EthAddress + " " + multiChainWallet.BtcAddress + " " + multiChainWallet.TronAddress
	default:
		return fmt.Errorf("未知链类型: %s", chain)
	}

	if isMnemonic {
		// 如果是助记词模式，写入地址和助记词
		_, err = file.WriteString(fmt.Sprintf("钱包地址: %s>>>助记词: %s\n", address, multiChainWallet.Mnemonic))
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	} else {
		// 如果是随机钱包模式，只写入地址和私钥
		_, err = file.WriteString(fmt.Sprintf("钱包地址: %s>>>私钥: %s\n", address, multiChainWallet.PrivateKey))
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}
	return nil
}

// deriveFromMnemonic 从指定助记词派生多个地址
func (app *App) deriveFromMnemonic() {
	generator := NewWalletGenerator(app.config)

	var mnemonic string
	var count int

	fmt.Print("输入助记词: ")
	fmt.Scanln(&mnemonic)
	fmt.Print("派生地址数量: ")
	fmt.Scanln(&count)

	config := GeneratorConfig{
		Count:       count,
		UseMnemonic: true,
		Mnemonic:    mnemonic,
	}

	wallets := generator.GenerateWallets(config)
	fmt.Printf("\n✅ 从助记词派生了 %d 个地址\n", len(wallets))

	for _, wallet := range wallets {
		PrintWalletSimple(wallet)
	}
}

// runAddressMatching 运行地址匹配模式
func (app *App) runAddressMatching() {
	if !app.config.AddressMatching.Enabled {
		fmt.Println("❌ 地址匹配功能未启用，请在config.yaml中配置")
		return
	}

	matcher := NewMatchingService(app.config)
	result := matcher.RunMatching()

	fmt.Printf("\n🏁 匹配完成，耗时: %v\n", result.Duration)
	fmt.Printf("找到 %d 个匹配地址\n", len(result.Wallets))

	for _, wallet := range result.Wallets {
		PrintWalletSimple(wallet)
	}
}

// runPerformanceBenchmark 运行性能基准测试
func (app *App) runPerformanceBenchmark() {
	fmt.Println("🔥 性能基准测试模式")

	tester := NewPerformanceTester(app.config)

	fmt.Println("选择测试模式:")
	fmt.Println("1. 完整基准测试")
	fmt.Println("2. 快速测试当前配置")
	fmt.Println("3. 自动选择最优协程数")

	var choice int
	fmt.Print("请选择 (1-3): ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		result, err := tester.RunBenchmark()
		if err != nil {
			fmt.Printf("❌ 基准测试失败: %v\n", err)
			return
		}
		fmt.Printf("\n🏆 推荐使用 %d 个协程 (%.2f 钱包/秒)\n",
			result.WorkerCount, result.WalletsPerSec)

	case 2:
		workerCount := app.config.GetOptimalWorkerCount()
		result, err := tester.QuickBenchmark(workerCount)
		if err != nil {
			fmt.Printf("❌ 快速测试失败: %v\n", err)
			return
		}
		fmt.Printf("当前配置性能: %.2f 钱包/秒\n", result.WalletsPerSec)

	case 3:
		optimalCount := tester.GetOptimalWorkerCountByBenchmark()
		fmt.Printf("🎯 建议协程数: %d\n", optimalCount)

		// 更新配置
		app.config.WorkerPool.ManualCount = optimalCount
		app.config.WorkerPool.AutoDetect = false

		fmt.Println("配置已更新，是否保存到文件? (y/n):")
		var save string
		fmt.Scanln(&save)
		if save == "y" || save == "Y" {
			if err := SaveConfig(app.config, "config.yaml"); err != nil {
				fmt.Printf("❌ 保存配置失败: %v\n", err)
			} else {
				fmt.Println("✅ 配置已保存")
			}
		}

	default:
		fmt.Println("无效选择")
	}
}
