## 说明
需要将registry文件（2.6.2版本）放在/home/bins目录下才能编译通过。
目前使用的golang版本是v1.12.2  并且使用了go mod作为依赖的管理工具，
gomod的命令如下：
```
cd /src
rm go.sum
rm vendor -rf
go mod vendor -v ##重新生成依赖
``

编译命令：
```
make compile
make build

or

make install
```
