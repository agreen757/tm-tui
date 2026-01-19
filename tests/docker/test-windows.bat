@echo off
REM Test script for Windows container
REM Tests Task Master installation on Windows with Git Bash

echo ========================================
echo Task Master Windows Installation Test
echo ========================================
echo.

REM Check environment
echo Environment Information:
echo   OS: %OS%
echo   Processor: %PROCESSOR_ARCHITECTURE%
echo   User: %USERNAME%
echo.

REM Check Git Bash
echo Checking Git Bash installation...
where bash >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] Git Bash not found in PATH
    exit /b 1
) else (
    echo [PASS] Git Bash found
)

REM Check Node.js
echo Checking Node.js installation...
where node >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] Node.js not found in PATH
    exit /b 1
) else (
    echo [PASS] Node.js found
    node --version
)

REM Check npm
echo Checking npm installation...
where npm >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] npm not found in PATH
    exit /b 1
) else (
    echo [PASS] npm found
    npm --version
)

REM Check make
echo Checking make installation...
where make >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] make not found in PATH
    exit /b 1
) else (
    echo [PASS] make found
)

REM Test Task Master installation
echo.
echo Testing Task Master installation...
cd C:\workspace
make install-task-master
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] make install-task-master failed
    exit /b 1
) else (
    echo [PASS] make install-task-master succeeded
)

REM Verify installation
echo.
echo Verifying Task Master installation...
make check-task-master
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] Task Master verification failed
    exit /b 1
) else (
    echo [PASS] Task Master verification succeeded
)

echo.
echo ========================================
echo All Windows tests passed!
echo ========================================
exit /b 0
