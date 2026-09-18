#!/bin/bash

# Agent Runtime - DeepSeek API Key Setup Script

echo "╔════════════════════════════════════════════════════════╗"
echo "║    Agent Runtime - DeepSeek API Key Setup             ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# 检查是否已经设置了 API Key
if [ -n "$DEEPSEEK_API_KEY" ]; then
    echo "✓ DEEPSEEK_API_KEY is already set"
    echo ""
    echo "Current API Key: ${DEEPSEEK_API_KEY:0:10}..."
    echo ""
    read -p "Do you want to update it? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled."
        exit 0
    fi
fi

# 提示用户输入 API Key
echo "Please enter your DeepSeek API Key:"
echo "(Format: sk-xxxxxxxxxxxxxxxx)"
echo ""
read -p "API Key: " api_key

# 验证 API Key 格式
if [[ ! $api_key =~ ^sk- ]]; then
    echo ""
    echo "❌ Error: Invalid API Key format"
    echo "API Key should start with 'sk-'"
    exit 1
fi

# 设置环境变量
export DEEPSEEK_API_KEY="$api_key"

echo ""
echo "✓ API Key set successfully!"
echo ""
echo "To make it permanent, add this line to your shell config:"
echo "  ~/.bashrc or ~/.zshrc:"
echo "  export DEEPSEEK_API_KEY=\"$api_key\""
echo ""
echo "Now you can run the demo:"
echo "  go run cmd/deepseek_demo/main.go"
echo ""
