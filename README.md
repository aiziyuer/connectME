connectDNS
[![CI](https://github.com/aiziyuer/connectDNS/workflows/CI/badge.svg)](https://github.com/aiziyuer/connectDNS/actions?query=workflow%3ACI) 
[![Release](https://github.com/aiziyuer/connectDNS/workflows/Release/badge.svg)](https://github.com/aiziyuer/connectDNS/releases/latest)
---

`connectDNS` is a tool help you get dns record by doh server.

## ⚙ Installation

``` bash
CGO_ENABLED=0 \
GOBIN=/usr/bin \
go get -u -v github.com/aiziyuer/connectDNS
```


## Testing

```

# test with udp
dig @127.0.0.1 -p1053 www.google.com +short
dig www.google.com +short

# test via tcp
nslookup -vc www.google.com 8.8.8.8

```

## AutoStart

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
RestartSec=3
Restart=always

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