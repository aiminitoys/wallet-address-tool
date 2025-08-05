package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

// BenchmarkResult 性能测试结果
type BenchmarkResult struct {
	WorkerCount   int           `json:"worker_count"`
	TotalTime     time.Duration `json:"total_time"`
	AvgTime       time.Duration `json:"avg_time"`
	WalletsPerSec float64       `json:"wallets_per_sec"`
	MemoryMB      float64       `json:"memory_mb"`
}

// PerformanceTester 性能测试器
type PerformanceTester struct {
	config *Config
	mutex  sync.RWMutex
}

// NewPerformanceTester 创建性能测试器
func NewPerformanceTester(config *Config) *PerformanceTester {
	return &PerformanceTester{
		config: config,
	}
}

// RunBenchmark 运行性能基准测试
func (pt *PerformanceTester) RunBenchmark() (*BenchmarkResult, error) {
	fmt.Println("🔥 开始性能基准测试...")

	bestResult := &BenchmarkResult{}
	var results []BenchmarkResult

	workerRange := pt.config.Performance.WorkerRange
	testSamples := pt.config.Performance.TestSamples

	// 测试不同协程数的性能
	for workers := workerRange.Min; workers <= workerRange.Max; workers += workerRange.Step {
		fmt.Printf("测试 %d 个协程... ", workers)

		result, err := pt.benchmarkWorkerCount(workers, testSamples)
		if err != nil {
			fmt.Printf("❌ 失败: %v\n", err)
			continue
		}

		results = append(results, *result)
		fmt.Printf("✅ %.2f 钱包/秒\n", result.WalletsPerSec)

		// 更新最佳结果
		if result.WalletsPerSec > bestResult.WalletsPerSec {
			*bestResult = *result
		}
	}

	// 显示测试结果
	pt.displayResults(results)

	fmt.Printf("\n🏆 最佳性能: %d 协程, %.2f 钱包/秒\n",
		bestResult.WorkerCount, bestResult.WalletsPerSec)

	return bestResult, nil
}

// benchmarkWorkerCount 测试指定协程数的性能
func (pt *PerformanceTester) benchmarkWorkerCount(workerCount, sampleCount int) (*BenchmarkResult, error) {
	generator := NewWalletGenerator(pt.config)

	config := GeneratorConfig{
		Count:          sampleCount,
		UseMnemonic:    false,
		ConcurrentMode: true,
		WorkerCount:    workerCount,
	}

	result := generator.GenerateBatch(config)

	return &BenchmarkResult{
		WorkerCount:   workerCount,
		TotalTime:     result.Duration,
		AvgTime:       result.AvgTime,
		WalletsPerSec: result.WalletsPerSec,
	}, nil
}

// displayResults 显示测试结果
func (pt *PerformanceTester) displayResults(results []BenchmarkResult) {
	if len(results) == 0 {
		return
	}

	// 按协程数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].WorkerCount < results[j].WorkerCount
	})

	fmt.Println("\n📊 性能测试报告:")
	fmt.Println("================================================================")
	fmt.Printf("%-10s %-12s %-12s %-15s\n", "协程数", "总时间", "平均时间", "钱包/秒")
	fmt.Println("================================================================")

	for _, result := range results {
		fmt.Printf("%-10d %-12s %-12s %-15.2f\n",
			result.WorkerCount,
			formatDuration(result.TotalTime),
			formatDuration(result.AvgTime),
			result.WalletsPerSec,
		)
	}
	fmt.Println("================================================================")
}

// GetOptimalWorkerCountByBenchmark 通过基准测试获取最优协程数
func (pt *PerformanceTester) GetOptimalWorkerCountByBenchmark() int {
	if !pt.config.Performance.AutoBenchmark {
		return pt.config.GetOptimalWorkerCount()
	}

	result, err := pt.RunBenchmark()
	if err != nil {
		fmt.Printf("⚠️  性能测试失败，使用默认协程数: %v\n", err)
		return pt.config.GetOptimalWorkerCount()
	}

	return result.WorkerCount
}

// QuickBenchmark 快速性能测试
func (pt *PerformanceTester) QuickBenchmark(workerCount int) (*BenchmarkResult, error) {
	// 使用较少的样本进行快速测试
	return pt.benchmarkWorkerCount(workerCount, 100)
}

// formatDuration 格式化时间显示
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.0fns", float64(d.Nanoseconds()))
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// EstimateOptimalWorkers 估算最优协程数（基于CPU和内存）
func EstimateOptimalWorkers() int {
	cpuCount := getCPUCount()

	// 基于经验公式：CPU核心数 * 2 到 CPU核心数 * 4 之间
	// 对于I/O密集型任务，可以设置更高的倍数
	// 对于CPU密集型任务，设置较低的倍数

	// 钱包生成主要是CPU密集型（加密计算），所以使用较低倍数
	optimalWorkers := cpuCount * 2

	// 限制范围
	if optimalWorkers < 1 {
		optimalWorkers = 1
	}
	if optimalWorkers > 32 {
		optimalWorkers = 32
	}

	return optimalWorkers
}

// getCPUCount 获取CPU核心数
func getCPUCount() int {
	return runtime.NumCPU()
}
