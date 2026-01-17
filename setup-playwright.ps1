# TRMNL-POWER Setup Script
# Auto-installs Playwright if needed

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TRMNL-POWER Setup" -ForegroundColor Cyan
Write-Host "Checking prerequisites..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check prerequisites status
Write-Host "Prerequisites Status:" -ForegroundColor Yellow
Write-Host "--------------------" -ForegroundColor Yellow
Write-Host ""

# Check if Node.js is installed
$nodeInstalled = $false
$nodeVersion = ""
try {
    $nodeOutput = node --version 2>$null
    if ($LASTEXITCODE -eq 0 -and $nodeOutput) {
        $nodeVersion = $nodeOutput.Trim()
        $nodeInstalled = $true
        Write-Host "  ✓ Node.js: Found ($nodeVersion)" -ForegroundColor Green
    } else {
        throw "Node.js not found"
    }
} catch {
    Write-Host "  ✗ Node.js: NOT FOUND" -ForegroundColor Red
}

# Check if npm is installed
$npmInstalled = $false
$npmVersion = ""
if ($nodeInstalled) {
    try {
        $npmOutput = npm --version 2>$null
        if ($LASTEXITCODE -eq 0 -and $npmOutput) {
            $npmVersion = $npmOutput.Trim()
            $npmInstalled = $true
            Write-Host "  ✓ npm: Found ($npmVersion)" -ForegroundColor Green
        } else {
            throw "npm not found"
        }
    } catch {
        Write-Host "  ✗ npm: NOT FOUND" -ForegroundColor Red
    }
} else {
    Write-Host "  ⚠ npm: Cannot check (Node.js not found)" -ForegroundColor Yellow
}

# Check if Playwright is installed
$playwrightInstalled = $false
if (Test-Path "node_modules\playwright\package.json") {
    try {
        $playwrightVersion = (Get-Content "node_modules\playwright\package.json" | ConvertFrom-Json).version
        $playwrightInstalled = $true
        Write-Host "  ✓ Playwright: Installed ($playwrightVersion)" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Playwright: Found but version unknown" -ForegroundColor Yellow
        $playwrightInstalled = $true
    }
} else {
    Write-Host "  ✗ Playwright: NOT INSTALLED" -ForegroundColor Red
}

Write-Host ""

# If Node.js is missing, provide clear instructions
if (-not $nodeInstalled) {
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "MISSING PREREQUISITE" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host ""
    Write-Host "Node.js is required but not found on your system." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "To install Node.js:" -ForegroundColor Cyan
    Write-Host "  1. Download from: https://nodejs.org/" -ForegroundColor White
    Write-Host "  2. Run the installer (LTS version recommended)" -ForegroundColor White
    Write-Host "  3. Restart your command prompt/PowerShell" -ForegroundColor White
    Write-Host "  4. Run this setup script again" -ForegroundColor White
    Write-Host ""
    Write-Host "Note: Node.js installer may require administrator privileges." -ForegroundColor Yellow
    Write-Host "      You'll only need to do this once." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 1
}

# If npm is missing (unlikely if Node.js is installed, but check anyway)
if (-not $npmInstalled) {
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "ERROR: npm not found" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host ""
    Write-Host "npm should come with Node.js. Please reinstall Node.js." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 1
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if package.json exists, create minimal one if not
if (-not (Test-Path "package.json")) {
    Write-Host "Creating package.json..." -ForegroundColor Yellow
    @{
        "name" = "trmnl-power"
        "version" = "1.0.0"
        "private" = $true
    } | ConvertTo-Json | Out-File -FilePath "package.json" -Encoding utf8
}

# Install Playwright if needed
if (-not $playwrightInstalled) {
    Write-Host "Installing Playwright..." -ForegroundColor Yellow
    Write-Host "This may take a few minutes (downloading browser binaries)..." -ForegroundColor Yellow
    Write-Host ""
    
    # Check if package.json exists, create minimal one if not
    if (-not (Test-Path "package.json")) {
        Write-Host "Creating package.json..." -ForegroundColor Gray
        @{
            "name" = "trmnl-power"
            "version" = "1.0.0"
            "private" = $true
        } | ConvertTo-Json | Out-File -FilePath "package.json" -Encoding utf8
    }
    
    npm install playwright
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "========================================" -ForegroundColor Red
        Write-Host "ERROR: Failed to install Playwright" -ForegroundColor Red
        Write-Host "========================================" -ForegroundColor Red
        Write-Host ""
        Write-Host "Please check:" -ForegroundColor Yellow
        Write-Host "  • Your internet connection" -ForegroundColor White
        Write-Host "  • npm is working: npm --version" -ForegroundColor White
        Write-Host "  • Try running: npm install playwright" -ForegroundColor White
        Write-Host ""
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        exit 1
    }
    
    Write-Host ""
    Write-Host "✓ Playwright installed successfully!" -ForegroundColor Green
} else {
    Write-Host "Playwright already installed - skipping." -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "You can now run: trmnl-power.exe" -ForegroundColor Cyan
Write-Host "Or use: run.bat" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

