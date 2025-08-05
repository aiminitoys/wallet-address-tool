#!/bin/bash

# 多链钱包生成器 v2.0 - 快速启动脚本

echo "🚀 多链钱包生成器 v2.0 - 快速启动"
echo "=================================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go 1.18+"
    exit 1
fi

echo "✅ Go环境检查通过"

# 安装依赖
echo "📦 安装依赖包..."
go mod tidy

# 编译程序
if [ $? -ne 0 ]; then
    echo "❌ 错误: 依赖安装失败，请检查Go环境和网络连接"
    exit 1
fi
echo "🔨 编译本机版本..."
go build -o wallet_generator *.go



if [ $? -eq 0 ]; then
    echo "✅ 编译成功!"
    echo ""
    echo "🎯 可用功能:"
    echo "1. 单个钱包生成（随机/助记词）"
    echo "2. 批量并发生成"
    echo "3. 助记词派生"
    echo "4. 地址匹配（靓号生成）"
    echo "5. 性能基准测试"
    echo ""
    echo "📝 配置文件: config.yaml"
    echo "📖 使用文档: README.md"
    echo ""
    echo "🚀 启动程序:"
    echo "./wallet_generator"
    echo ""
    echo "🎯 地址匹配示例:"
    echo "cp config_matching_example.yaml config.yaml && ./wallet_generator"
else
    echo "❌ 编译失败!"
    exit 1
fi
