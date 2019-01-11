@set DLL=libfwapp_go
@set DLLFILE=%DLL%.dll
@set DLLHEADER=%DLL%.h
@set BINPATH=bin\win64\

@rem delete old bin files
@if exist %BINPATH%%DLLFILE%   del /F %BINPATH%%DLLFILE%
@if exist %BINPATH%%DLLHEADER% del /F %BINPATH%%DLLHEADER%

@echo building...
@rem build go, generate a dll file
go build -o %BINPATH%%DLLFILE% -ldflags "-w" -buildmode=c-shared fwapp\fwapp.go

@rem delete c header file, no need in windows(compiled with errors, so fuck windows)
@del /F %BINPATH%%DLLHEADER%

@echo Over