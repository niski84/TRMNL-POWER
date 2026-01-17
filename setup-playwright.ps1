# TRMNL-POWER Setup Script
# Auto-installs Playwright if needed

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TRMNL-POWER Setup" -ForegroundColor Cyan
Write-Host "Installing Playwright if needed..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Node.js is installed
try {
    $nodeVersion = node --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Node.js found: $nodeVersion" -ForegroundColor Green
    } else {
        throw "Node.js not found"
    }
} catch {
    Write-Host "✗ ERROR: Node.js not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Node.js from: https://nodejs.org/" -ForegroundColor Yellow
    Write-Host "After installing Node.js, run this script again." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 1
}

# Check if package.json exists, create minimal one if not
if (-not (Test-Path "package.json")) {
    Write-Host "Creating package.json..." -ForegroundColor Yellow
    @{
        "name" = "trmnl-power"
        "version" = "1.0.0"
        "private" = $true
    } | ConvertTo-Json | Out-File -FilePath "package.json" -Encoding utf8
}

# Check if node_modules/playwright exists
if (-not (Test-Path "node_modules\playwright")) {
    Write-Host ""
    Write-Host "Playwright not found. Installing..." -ForegroundColor Yellow
    Write-Host "This may take a few minutes (downloading browser binaries)..." -ForegroundColor Yellow
    Write-Host ""
    
    npm install playwright
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "✗ ERROR: Failed to install Playwright" -ForegroundColor Red
        Write-Host "Please check your internet connection and try again." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        exit 1
    }
    
    Write-Host ""
    Write-Host "✓ Playwright installed successfully!" -ForegroundColor Green
} else {
    Write-Host "✓ Playwright already installed." -ForegroundColor Green
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "You can now run: trmnl-power.exe" -ForegroundColor Cyan
Write-Host "Or use: run.bat" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

