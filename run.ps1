# Agent Runtime - 快速启动脚本

# 检查 API Key（从环境变量读取）
if (-not $env:DEEPSEEK_API_KEY) {
    Write-Host "⚠️  警告: 未设置 DEEPSEEK_API_KEY 环境变量" -ForegroundColor Yellow
    Write-Host "   运行真实 Demo 需要先设置 API Key:" -ForegroundColor Gray
    Write-Host '   $env:DEEPSEEK_API_KEY = "your-api-key-here"' -ForegroundColor Gray
    Write-Host ""
}

Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  Agent Runtime - Quick Start" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# 检查 Go 环境
Write-Host "检查环境..." -ForegroundColor Yellow
$goVersion = go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 错误: 未安装 Go 或不在 PATH 中" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Go 已安装: $goVersion" -ForegroundColor Green

# 进入项目目录
$projectRoot = "C:\Users\lenovo\agent-runtime"
if (-not (Test-Path $projectRoot)) {
    Write-Host "❌ 错误: 项目目录不存在: $projectRoot" -ForegroundColor Red
    exit 1
}
Set-Location $projectRoot
Write-Host "✓ 项目目录: $projectRoot" -ForegroundColor Green
Write-Host ""

# 显示菜单
Write-Host "请选择要运行的示例:" -ForegroundColor Yellow
Write-Host "  1. 运行 DeepSeek Demo (基础版)" -ForegroundColor White
Write-Host "  2. 运行 DeepSeek Demo (增强版)" -ForegroundColor White
Write-Host "  3. 运行测试 (使用 Fake LLM)" -ForegroundColor White
Write-Host "  4. 查看项目结构" -ForegroundColor White
Write-Host "  5. 退出" -ForegroundColor White
Write-Host ""

$choice = Read-Host "请输入选项 (1-5)"

switch ($choice) {
    "1" {
        Write-Host ""
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host "  运行 DeepSeek Demo (基础版)" -ForegroundColor Cyan
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host ""
        go run cmd/deepseek_demo/main.go
    }
    "2" {
        Write-Host ""
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host "  运行 DeepSeek Demo (增强版)" -ForegroundColor Cyan
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host ""
        go run cmd/deepseek_demo_v2/main.go
    }
    "3" {
        Write-Host ""
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host "  运行测试 (使用 Fake LLM)" -ForegroundColor Cyan
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host ""
        Set-Location pkg/runtime
        go test -v -run "Test(DirectAnswer|SingleToolCall|MultipleToolCalls|ToolFailure|UnknownTool|MaxSteps|SessionIsolation|FollowUp)"
        Set-Location ../..
    }
    "4" {
        Write-Host ""
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host "  项目结构" -ForegroundColor Cyan
        Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
        Write-Host ""
        Get-ChildItem -Recurse -Directory -Depth 2 | Select-Object FullName | Format-Table -AutoSize
    }
    "5" {
        Write-Host ""
        Write-Host "再见！" -ForegroundColor Green
        exit 0
    }
    default {
        Write-Host ""
        Write-Host "❌ 无效选项，请重新运行脚本" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  完成" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
