CXX=g++

ifndef DEBUG
DEBUG=1
endif

# so NAME
FWDLL=fwapp
FWDLLFILE=lib$(FWDLL).so

# version info
ifndef VERSION
VERSION="\"10.0.1\""
endif

ifndef BUILDDATE
BUILDDATE= "\"`date +%Y-%m-%d\ %k:%M:%S`\""
endif

ifndef BUILDINFO
BUILDINFO=-DVERSION=$(VERSION) -DBUILDDATE=$(BUILDDATE)
endif

#include dirs
INCLUDEDIR=-I/usr/local/include -I. -I../include

#include libs
link_lib_a=
link_lib_so=-L../lib -lpthread -lfwapp_go
LINK_ALL_LIB=$(link_lib_a) $(link_lib_so)


ifeq ($(DEBUG),1)
CXXFLAGS=-fPIC -g -Wall -Wno-unused-parameter -Wno-deprecated $(BUILDINFO) $(INCLUDEDIR)
else
CXXFLAGS=-fPIC -Wall -O3 -Wno-unused-parameter $(BUILDINFO) $(INCLUDEDIR)
endif


#objects
cpp_self=$(wildcard ../source/*.cpp)
objs_self=$(patsubst %.cpp, %.o, $(cpp_self))
OBJS=$(objs_self)
all:$(FWDLLFILE)

$(FWDLLFILE):$(OBJS)
	$(CXX) -fPIC -shared -o $@ $(OBJS) $(LINK_ALL_LIB)
	@mv -f $(FWDLLFILE) ../lib/
$%.o:%.cpp %.h
	$(CXX) $(CXXFLAGS) -c -o $@ $<
.PHONY: clean
clean:
	@rm -rfv ../source/*.o ../lib/$(FWDLLFILE)
	@echo "clean: [fwapp] clean over."
install:
