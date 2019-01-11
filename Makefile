DLL=libfwapp_go
DLLFILE=$(DLL).so
DLLHEADER=$(DLL).h
BINPATH=bin/linux64/

all:
	@go build -o $(BINPATH)$(DLLFILE) -ldflags "-w" -buildmode=c-shared fwapp/fwapp.go
	@cp fwapp/libfwapp_go-config.json bin/linux64/ -fv
clean:
	@rm -rfv $(BINPATH)$(DLLFILE) $(BINPATH)$(DLLHEADER) $(BINPATH)fwserver-config.json
