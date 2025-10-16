@echo off
REM Test script for Windows Credential Manager
echo Testing FDOT Credential Manager...
echo.

echo === Setting a test credential ===
credmgr.exe set fdot-test "Hello from FDOT credential manager!"
if %ERRORLEVEL% neq 0 goto error

echo.
echo === Reading the test credential ===
credmgr.exe get fdot-test
if %ERRORLEVEL% neq 0 goto error

echo.
echo === Deleting the test credential ===
credmgr.exe del fdot-test
if %ERRORLEVEL% neq 0 goto error

echo.
echo === Listing all credentials ===
credmgr.exe list
if %ERRORLEVEL% neq 0 goto error

echo.
echo === Verifying deletion (should fail) ===
credmgr.exe get fdot-test
if %ERRORLEVEL% equ 0 echo ERROR: Credential should have been deleted!

echo.
echo === Test complete! ===
goto end

:error
echo ERROR: Test failed with error code %ERRORLEVEL%
exit /b 1

:end
echo All tests completed successfully.