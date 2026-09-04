@echo off
cls
echo Please wait, Orbit is currently building...
go run github.com/tc-hib/go-winres@latest simply --icon internal\assets\icon.ico --manifest gui --out rsrc >nul 2>nul
go build -ldflags="-H windowsgui -s -w" -o Orbit.exe .
del /f /q rsrc_windows_*.syso >nul 2>nul
