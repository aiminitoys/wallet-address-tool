package main

import (
	"fmt"
	"log"
	"runtime"
)

func main() {
	// 加载配置
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}
	GlobalConfig = config

	fmt.Println("🚀 多链钱包生成器 v2.0")
	fmt.Printf("💻 CPU核心数: %d, 推荐协程数: %d\n", runtime.NumCPU(), config.GetOptimalWorkerCount())

	// 显示地址匹配状态
	if config.AddressMatching.Enabled {
		fmt.Printf("🎯 地址匹配已启用 - 目标链: %v\n", config.AddressMatching.TargetChains)
	}

	// 创建应用实例
	app := NewApp(config)

	// 运行应用
	app.Run()
}
