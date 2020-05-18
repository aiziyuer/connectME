connectME
[![CI](https://github.com/aiziyuer/connectME/workflows/CI/badge.svg)](https://github.com/aiziyuer/connectME/actions?query=workflow%3ACI) 
[![Release](https://github.com/aiziyuer/connectME/workflows/Release/badge.svg)](https://github.com/aiziyuer/connectME/releases/latest)
---

`connectME` is a tool that help you connect the internet freely.

## ⚙ Installation

``` bash
CGO_ENABLED=0 \
GOBIN=/usr/bin \
go get -u -v github.com/aiziyuer/connectME
```


## 🧼 DNS Server

```
# start it
➜  connectME git:(master) ✗ ./connectME dns
ednsSubnet: 122.235.189.0/24
tcp_server: 0.0.0.0:1053
udp_server: 0.0.0.0:1053

# test with udp
dig @127.0.0.1 -p10053 www.google.com +short
dig www.google.com +short

# test via tcp
nslookup -vc www.google.com 8.8.8.8

```

## ☕ AutoStart

```bash
cat <<'EOF' >/etc/systemd/system/connectME@.service
[Unit]
Description=connectDNS
Documentation=https://github.com/aiziyuer/connectME
After=network.target

[Service]
Type=simple
Environment="HTTP_PROXY=127.0.0.1:3128"
Environment="HTTPS_PROXY=127.0.0.1:3128"
ExecStart=/usr/bin/connectME @ --port 53
ProtectSystem=strict
RestartSec=1
Restart=always

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable connectME@dns
systemctl start connectME@dns
```

## 🙏 FAQ

- [Using Cobra With Golang](https://o-my-chenjian.com/2017/09/20/Using-Cobra-With-Golang/)
- [goproxy](https://goproxy.io/zh/)