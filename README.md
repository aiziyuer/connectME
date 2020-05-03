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

## 开机启动

```bash
cat <<'EOF' >/etc/systemd/system/connectDNS.service
[Unit]
Description=connectDNS
Documentation=https://github.com/aiziyuer/connectDNS
After=network.target

[Service]
Type=simple
Environment="HTTP_PROXY=127.0.0.1:3128"
Environment="HTTPS_PROXY=127.0.0.1:3128"
ExecStart=/usr/bin/connectDNS
ProtectSystem=strict

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable connectDNS
systemctl start connectDNS
```

## FAQ

- [Using Cobra With Golang](https://o-my-chenjian.com/2017/09/20/Using-Cobra-With-Golang/)
- [goproxy](https://goproxy.io/zh/)