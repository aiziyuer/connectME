connectDNS 
---

`connectDNS` is a tool help you get dns record by doh server.

## ⚙ Installation

``` bash
CGO_ENABLED=0 \
GOBIN=/usr/bin \
go get -u -v github.com/aiziyuer/connectDNS
```


## 服务测试

```

# 测试DNS解析
dig @127.0.0.1 -p1053 www.google.com +short
dig www.google.com +short

# 用tcp可以测试出正确定制
nslookup -vc www.google.com 8.8.8.8

```