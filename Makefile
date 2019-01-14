DLL=libfwapp_go
DLLFILE=$(DLL).so
DLLHEADER=$(DLL).h
BINPATH_R=bin/release/
BINPATH_D=bin/debug/
BINPATH=$(BINPATH_R)

ifndef DEBUG
DEBUG=0
endif

LDFLAGS=-ldflags "-s -w"
ifeq ($(DEBUG),1)
LDFLAGS=
BINPATH=$(BINPATH_D)
endif

all:
	@echo "building..."
	@go build -o $(BINPATH)$(DLLFILE) $(LDFLAGS) -buildmode=c-shared fwapp/fwapp.go
	@cp fwapp/libfwapp_go-config.json $(BINPATH) -f
clean:
	@rm -rfv $(BINPATH_R)$(DLLFILE) $(BINPATH_R)$(DLLHEADER) $(BINPATH_R)fwserver-config.json
	@rm -rfv $(BINPATH_D)$(DLLFILE) $(BINPATH_D)$(DLLHEADER) $(BINPATH_D)fwserver-config.json
