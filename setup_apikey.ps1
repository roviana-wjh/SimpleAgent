# Agent Runtime - DeepSeek API Key Setup Script (Windows)

Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║    Agent Runtime - DeepSeek API Key Setup             ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# 检查是否已经设置了 API Key
if ($env:DEEPSEEK_API_KEY) {
    Write-Host "✓ DEEPSEEK_API_KEY is already set" -ForegroundColor Green
    Write-Host ""
    $masked = $env:DEEPSEEK_API_KEY.Substring(0, [Math]::Min(10, $env:DEEPSEEK_API_KEY.Length)) + "..."
    Write-Host "Current API Key: $masked"
    Write-Host ""
    $update = Read-Host "Do you want to update it? (y/N)"
    if ($update -ne "y" -and $update -ne "Y") {
        Write-Host "Setup cancelled." -ForegroundColor Yellow
        exit 0
    }
}

# 提示用户输入 API Key
Write-Host "Please enter your DeepSeek API Key:"
Write-Host "(Format: sk-xxxxxxxxxxxxxxxx)" -ForegroundColor Gray
Write-Host ""
$apiKey = Read-Host "API Key"

# 验证 API Key 格式
if (-not $apiKey.StartsWith("sk-")) {
    Write-Host ""
    Write-Host "❌ Error: Invalid API Key format" -ForegroundColor Red
    Write-Host "API Key should start with 'sk-'" -ForegroundColor Red
    exit 1
}

# 设置环境变量（当前会话）
$env:DEEPSEEK_API_KEY = $apiKey

Write-Host ""
Write-Host "✓ API Key set successfully for current session!" -ForegroundColor Green
Write-Host ""
Write-Host "To make it permanent, you have two options:" -ForegroundColor Yellow
Write-Host ""
Write-Host "Option 1: Set User Environment Variable (Recommended)" -ForegroundColor Cyan
Write-Host "  [System.Environment]::SetEnvironmentVariable('DEEPSEEK_API_KEY', '$apiKey', 'User')" -ForegroundColor Gray
Write-Host ""
Write-Host "Option 2: Add to PowerShell Profile" -ForegroundColor Cyan
Write-Host "  Add this line to your `$PROFILE file:" -ForegroundColor Gray
Write-Host "  `$env:DEEPSEEK_API_KEY = '$apiKey'" -ForegroundColor Gray
Write-Host ""
Write-Host "Now you can run the demo:" -ForegroundColor Green
Write-Host "  go run cmd/deepseek_demo/main.go" -ForegroundColor White
Write-Host ""
