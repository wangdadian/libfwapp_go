@set DLL=libfwapp_go
@set DLLFILE=%DLL%.dll
@set DLLHEADER=%DLL%.h
@set BINPATH_D=bin\debug\
@set BINPATH_R=bin\release\

@set BINPATH=%BINPATH_R%
@if "%1"=="debug" set BINPATH=%BINPATH_D%

@set LDFLAGS=-ldflags "-s -w"
@if "%1"=="debug" set LDFLAGS=

@rem delete old bin files
@if exist %BINPATH%%DLLFILE%   del /F %BINPATH%%DLLFILE%
@if exist %BINPATH%%DLLHEADER% del /F %BINPATH%%DLLHEADER%

@echo building...
@rem build go, generate a dll file
go build -o %BINPATH%%DLLFILE% %LDFLAGS% -buildmode=c-shared fwapp\fwapp.go

@rem delete c header file, no need in windows(compiled with errors, so, fuck windows)
@rem @del /F %BINPATH%%DLLHEADER%

@echo Over