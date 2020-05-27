gw模式
===

透明代理

### 测试

```
# 启动容器环境
docker run --privileged --rm -ti golang bash

# 安装iptables
apt-get update && apt-get -y install iptables

# 安装
CGO_ENABLED=0 \
GOBIN=/usr/bin \
go get -u -v github.com/aiziyuer/connectME

# 启动服务(当前只支持http代理)
export http_proxy=10.10.10.254:3128
export https_proxy=10.10.10.254:3128
connectME gw --port 11081 &

# 设置防火墙策略
iptables -t nat -N PROXY
iptables -t nat -A PROXY -d 127.0.0.0/8 -j RETURN
iptables -t nat -A PROXY -d 10.0.0.0/8 -j RETURN
iptables -t nat -A PROXY -d 169.254.0.0/16 -j RETURN
iptables -t nat -A PROXY -d 172.16.0.0/12 -j RETURN
iptables -t nat -A PROXY -d 192.168.0.0/16 -j RETURN
iptables -t nat -A PROXY -d 224.0.0.0/4 -j RETURN
iptables -t nat -A PROXY -d 240.0.0.0/4 -j RETURN
iptables -t nat -A PROXY -p tcp -j LOG --log-prefix "PROXY " --log-level 6
iptables -t nat -A PROXY -p tcp -j REDIRECT --to-ports 11081

iptables -t nat -A OUTPUT -p tcp -j PROXY

# 测试服务
curl -v -H "Host: dns.google.com" https://8.8.8.8/resolve?name=www.google.com&type=A


```