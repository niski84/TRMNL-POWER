@echo off
REM TRMNL-POWER Launcher
REM Automatically sets up Playwright on first run, then starts the server

echo ========================================
echo TRMNL-POWER Launcher
echo ========================================
echo.

REM Check if setup was run (node_modules/playwright exists)
if not exist "node_modules\playwright" (
    echo First-time setup detected.
    echo Running setup script...
    echo.
    powershell.exe -ExecutionPolicy Bypass -File setup-playwright.ps1
    if errorlevel 1 (
        echo.
        echo Setup failed. Please check errors above.
        echo.
        pause
        exit /b 1
    )
    echo.
)

REM Check if executable exists
if not exist "trmnl-power.exe" (
    echo ERROR: trmnl-power.exe not found!
    echo Please make sure you extracted all files from the ZIP.
    echo.
    pause
    exit /b 1
)

REM Run the server
echo Starting TRMNL-POWER server...
echo.
trmnl-power.exe

REM If server exits, pause so user can see any error messages
if errorlevel 1 (
    echo.
    echo Server exited with an error.
    pause
)

