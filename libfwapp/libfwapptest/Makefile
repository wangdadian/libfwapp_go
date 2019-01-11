CXX=g++

ifndef DEBUG
DEBUG=1
endif

EXE=fwtest

cpp_self=$(wildcard ./*.cpp)
OBJS=$(patsubst %.cpp, %.o, $(notdir $(cpp_self)))

ifeq ($(DEBUG),1)
CXXFLAGS=-fPIC -g -Wall -Wno-unused-parameter -Wno-deprecated
else
CXXFLAGS=-fPIC -Wall -O3 -Wno-unused-parameter
endif

$(EXE):$(OBJS)
	$(CXX) -o $@ $< -L../libfwapp/lib -lfwapp -lfwapp_go

%.o: %.cpp %.h
	$(CXX) $(CXXFLAGS) -c -o $@ $<
.PHONY: clean
clean:
	@rm -rfv *.o $(EXE)