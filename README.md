# libfwapp_go.dll/so动态库源码文件

----

## 1. 编译说明

1.1 搭建go编译环境，将源码目录libfwapp_go、dadian拷贝至src目录下

1.2 进入libfwapp_go目录下：

1.2.1 windows环境下

​		执行build.bat（需要搭建mingGW-64编译环境，具体百度）

​		build.bat  debug是编译调试信息版本

​		会在上层目录bin下生成dll文件

​		使用此dll文件中的接口时，只支持LoadLibrary形式加载接口

1.2.2 linux环境下

​		执行make（make DEBUG=1是编译调试版本）

​		会在/bin目录下生成so文件

​		使用此so文件时，#include 此目录下的.h头文件即可，并在编译时，指定动态库链接：

​		gcc/g++    ……    <font face="微软雅黑" size=5 style="color:red">**-lfwapp_go  -lpthread**</font>



