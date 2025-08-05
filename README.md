# 多链钱包生成器 v2.0 🚀

一个功能强大的多链钱包生成器，支持以太坊、比特币、波场、BSC、Polygon等多个区块链网络。

## ✨ 新功能特性

### 🔧 配置管理
- **YAML配置文件**: 通过 `config.yaml` 统一管理所有配置
- **智能默认值**: 自动生成默认配置文件
- **灵活配置**: 支持运行时修改和保存配置

### 🎯 地址匹配功能
- **前缀匹配**: 生成指定前缀的地址（如 0x888...）
- **后缀匹配**: 生成指定后缀的地址（如 ...888）
- **包含匹配**: 生成包含特定字符串的地址
- **正则表达式**: 支持复杂的模式匹配
- **多链支持**: 可指定匹配特定区块链的地址

### 🚀 智能协程池
- **自动检测**: 基于CPU核心数自动选择最优协程数
- **性能测试**: 内置基准测试，找出最佳性能配置
- **灵活配置**: 支持手动设置协程数和倍数

### 📊 性能优化
- **基准测试**: 全面的性能测试工具
- **智能调优**: 自动找出最优协程配置
- **实时统计**: 显示生成速度和匹配率

## 🛠️ 安装和运行

### 环境要求
- Go 1.18+
- 支持的操作系统：Windows、macOS、Linux

### 快速开始

1. **克隆项目**
```bash
git clone https://github.com/aiminitoys/wallet-address-tool
cd wallet_create_address
```

2. **安装依赖**
```bash
go mod tidy
```

3. **运行程序**
```bash
go run *.go
```

## 📖 功能说明

### 1. 单个钱包生成
- 随机生成单个钱包
- 从助记词生成钱包

### 2. 批量生成模式
- 高性能并发生成
- 自动优化协程数
- 支持大批量生成

### 3. 助记词派生
- 从指定助记词派生多个地址
- 遵循BIP44标准

### 4. 地址匹配模式 🎯
生成符合特定规则的靓号地址：

```yaml
# 配置示例
address_matching:
  enabled: true
  rules:
    prefixes: ["888", "666"]  # 前缀匹配
    suffixes: ["888", "999"]  # 后缀匹配
    contains: ["love"]        # 包含匹配
    regex: "(\\d)\\1{3,}"     # 正则匹配
  target_chains: ["eth"]      # 目标链
  max_attempts: 10000         # 最大尝试次数
```

### 5. 性能基准测试 📊
- 完整基准测试：测试不同协程数的性能
- 快速测试：测试当前配置性能
- 自动优化：找出最佳协程配置

## ⚙️ 配置文件说明

### config.yaml 完整配置

```yaml
# 生成器配置
generator:
  default_count: 1           # 默认生成数量
  use_mnemonic: false        # 默认是否使用助记词
  batch_default_count: 100   # 批量生成默认数量

# 协程池配置
worker_pool:
  auto_detect: true          # 自动检测最优协程数
  manual_count: 4            # 手动指定协程数
  cpu_multiplier: 2          # CPU倍数
  min_workers: 1             # 最小协程数
  max_workers: 32            # 最大协程数

# 地址匹配配置
address_matching:
  enabled: false             # 是否启用地址匹配
  rules:
    ignore_case: false # 是否忽略大小写
    prefixes: ["888"]        # 前缀匹配
    suffixes: ["888"]        # 后缀匹配
    suffix_same: 4        # 后缀连续匹配
    contains: ["love"]       # 包含匹配
    regex: ""                # 正则表达式
  target_chains: ["eth"]     # 目标区块链
  max_attempts: 10000        # 最大尝试次数

# 性能测试配置
performance:
  auto_benchmark: false      # 启动时自动基准测试
  test_samples: 1000         # 测试样本数
  worker_range:              # 协程数测试范围
    min: 1
    max: 16
    step: 2

# 输出配置
output:
  verbose: true              # 详细输出
  preview_count: 5           # 预览钱包数量
  save_to_file: false        # 保存到文件
  output_file: "wallets.json" # 输出文件路径
```

## 🎯 地址匹配示例

### 1. 生成前缀为888的地址
```yaml
address_matching:
  enabled: true
  rules:
    prefixes: ["888"]
  target_chains: ["eth"]
```

### 2. 生成包含"love"的地址
```yaml
address_matching:
  enabled: true
  rules:
    contains: ["love"]
  target_chains: ["eth"]
```

### 3. 生成连续数字的地址（正则）
```yaml
address_matching:
  enabled: true
  rules:
    regex: "(\\d)\\1{3,}"  # 匹配连续4个或更多相同数字
  target_chains: ["eth"]
```

## 🔧 性能优化建议

### 协程数设置
- **CPU密集型**: 协程数 = CPU核心数 × 1-2
- **I/O密集型**: 协程数 = CPU核心数 × 2-4
- **建议**: 使用内置基准测试找出最优值

### 地址匹配优化
- 简单规则（前缀/后缀）性能最好
- 正则表达式性能较低，但功能最强
- 合理设置最大尝试次数

## 📊 性能参考

测试环境：8核CPU，16GB内存

| 协程数 | 生成速度 | CPU使用率 | 推荐场景 |
|--------|----------|----------|----------|
| 4      | ~38000/s  | 60%      | 普通生成 |
| 8      | ~78000/s  | 85%      | 高性能生成 |
| 16     | ~96000/s  | 95%      | 极限性能 |

## 🔒 安全提醒

⚠️ **重要提醒**：
- 妥善保管私钥和助记词
- 不要在不安全的环境中运行
- 生产使用前请充分测试
- 建议在离线环境中生成钱包

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

感谢以下开源项目：
- go-ethereum
- btcd
- go-bip39
- go-bip32
