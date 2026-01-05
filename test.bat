@echo off
REM Pre-deployment test script for Windows
REM Runs all tests for frontend and backend

echo =========================================
echo   Folda Finances - Pre-Deployment Tests
echo =========================================
echo.

set OVERALL_STATUS=0

REM Backend Tests
echo Running Backend Tests...
echo ----------------------------------------
cd backend
set CGO_ENABLED=1
go test ./... -v
if %ERRORLEVEL% NEQ 0 (
    echo [FAILED] Backend tests failed
    echo.
    echo Note: Backend tests require CGO (C compiler) on Windows.
    echo If you see CGO errors, see TESTING.md for setup instructions.
    echo You can still deploy if frontend tests pass - backend tests
    echo will run in CI/CD on Linux.
    set OVERALL_STATUS=1
) else (
    echo [PASSED] Backend tests passed
)
cd ..
echo.

REM Frontend Tests
echo Running Frontend Tests...
echo ----------------------------------------
cd frontend
call npm test -- --run
if %ERRORLEVEL% NEQ 0 (
    echo [FAILED] Frontend tests failed
    set OVERALL_STATUS=1
) else (
    echo [PASSED] Frontend tests passed
)
cd ..
echo.

REM Summary
echo =========================================
if %OVERALL_STATUS% EQU 0 (
    echo [SUCCESS] All tests passed! Ready to deploy.
    exit /b 0
) else (
    echo [ERROR] Some tests failed. Fix issues before deploying.
    exit /b 1
)
echo =========================================
