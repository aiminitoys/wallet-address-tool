#!/bin/bash

# 多链钱包生成器构建脚本

echo "🔨 构建多链钱包生成器 v2.0..."

# 清理之前的构建
echo "🧹 清理旧的构建文件..."
rm -f wallet_generator

# 检查依赖
echo "📦 检查依赖..."
go mod tidy

# 运行测试
echo "🧪 运行测试..."
go test ./... -v

# 构建项目
echo "🏗️  构建项目..."
echo "🔨 编译 本机 版本..."
go build -o wallet_generator *.go

echo "🔨 编译 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -o wallet_generator_windows.exe *.go

echo "🔨 编译 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -o wallet_generator_linux *.go

echo "🔨 编译 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -o wallet_generator_macos *.go

echo "✅ 所有平台编译完成！"
if [ $? -eq 0 ]; then
    echo "✅ 构建成功！"
    echo "🚀 运行: ./wallet_generator"
    echo ""
    echo "📁 项目结构:"
    echo "├── main.go          # 程序入口"
    echo "├── app.go           # 应用主逻辑"
    echo "├── types.go         # 类型定义"
    echo "├── config.go        # 配置管理"
    echo "├── generator.go     # 钱包生成器"
    echo "├── matcher.go       # 地址匹配器"
    echo "├── matching.go      # 匹配服务"
    echo "├── benchmark.go     # 性能测试"
    echo "├── printer.go       # 输出管理"
    echo "└── config.yaml      # 配置文件"
else
    echo "❌ 构建失败！"
    exit 1
fi
