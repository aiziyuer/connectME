connectME
[![CI](https://github.com/aiziyuer/connectME/workflows/CI/badge.svg)](https://github.com/aiziyuer/connectME/actions?query=workflow%3ACI) 
[![Release](https://github.com/aiziyuer/connectME/workflows/Release/badge.svg)](https://github.com/aiziyuer/connectME/releases/latest)
---

`connectME` is a tool that help you connect the internet freely.

## ⚙ Installation

``` bash
# 如果是升级, 最好先把老版本删除
# rm -rf /usr/bin/connectME
CGO_ENABLED=0 \
GOBIN=/usr/bin \
go get -u -v github.com/aiziyuer/connectME
```
> 为了避免goproxy的缓存问题, 可以`export GOPROXY=direct`强制从源站下载

## 🧼 Serve DNS

```
# start it
➜  ~ connectME dns --port 53
ednsSubnet: 122.235.189.0/24
http_proxy: http://127.0.0.1:3128
https_proxy: http://127.0.0.1:3128
tcp_server: 0.0.0.0:53
udp_server: 0.0.0.0:53

# test
dig @127.0.0.1 -p53 www.google.com +short

```

## 🚪 Serve GW

网关服务安装

``` bash
# start it
➜  ~ export http_proxy=127.0.0.1:3128
➜  ~ export https_proxy=127.0.0.1:3128
➜  ~ connectME gw --port 1081
http_proxy: 127.0.0.1:3128
https_proxy: 127.0.0.1:3128
gw_server: 0.0.0.0:1081

```

由于透明网关需要防火墙上面将流量引入, 所以有如下推荐的防火墙配置:

``` bash
# install ipset
yum install ipset
# apt-get install ipset

# clean
# ipset flush NO_PROXY && ipset destory NO_PROXY || true

# create, ref: https://zh.wikipedia.org/wiki/%E4%BF%9D%E7%95%99IP%E5%9C%B0%E5%9D%80
ipset create NO_PROXY hash:net comment
ipset -exist add NO_PROXY 0.0.0.0/8 comment "IANA"
ipset -exist add NO_PROXY 10.0.0.0/8 comment "Class C IP address"
ipset -exist add NO_PROXY 172.16.0.0/12 comment "Class C IP address"
ipset -exist add NO_PROXY 192.168.0.0/16 comment "Class C IP address"
ipset -exist add NO_PROXY 127.0.0.0/8 comment "Loopback Address"
ipset -exist add NO_PROXY 169.254.0.0/16 comment "Link local address"
ipset -exist add NO_PROXY 224.0.0.0/16 comment "Multicast Address"
# ipset -exist add NO_PROXY xxxx/32 comment "Your Proxy Server"

# apply chain
iptables -t nat -N PROXY &>/dev/null; iptables -t nat -F PROXY
iptables -t nat -A PROXY -m set --match-set NO_PROXY dst -j RETURN
iptables -t nat -A PROXY -p tcp -j REDIRECT --to-port 1081
iptables -t nat -F OUTPUT && iptables -t nat -A OUTPUT -j PROXY 
iptables -t nat -F PREROUTING && iptables -t nat -A PREROUTING -p tcp -j PROXY

# review chain
iptables -t nat -S

```

## 🎂 AutoStart

``` bash

# ipset auto restore at reboot
cat <<'EOF' >/etc/systemd/system/ipset.service
[Unit]
Description=ipset persistent rule service
Before=iptables.service
ConditionFileNotEmpty=/etc/sysconfig/ipset
[Service]
Type=oneshot
RemainAfterExit=true
ExecStart=/usr/sbin/ipset -exist -file /etc/sysconfig/ipset restore
ExecStop=/usr/sbin/ipset -file /etc/sysconfig/ipset save
[Install]
WantedBy=multi-user.target
EOF
# apply ipset
systemctl daemon-reload; systemctl enable ipset.service; systemctl start ipset.service

# add systemd service
cat <<'EOF' >/etc/systemd/system/connectME@.service
[Unit]
Description=connectME dns
Documentation=https://github.com/aiziyuer/connectME
After=network.target

[Service]
Type=simple
Environment="HTTP_PROXY=127.0.0.1:3128"
Environment="HTTPS_PROXY=127.0.0.1:3128"
ExecStart=/usr/bin/connectME %i
RestartSec=1
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# apply dns
systemctl daemon-reload; systemctl enable connectME@dns.service; systemctl start connectME@dns.service

# apply gw
systemctl daemon-reload; systemctl enable connectME@gw.service; systemctl start connectME@gw.service
```

## 🙏 FAQ

- [Using Cobra With Golang](https://o-my-chenjian.com/2017/09/20/Using-Cobra-With-Golang/)
- [goproxy](https://goproxy.io/zh/)
- [emojipedia.org](https://emojipedia.org/)
- [The @ symbol and systemctl and vsftpd](https://superuser.com/questions/393423/the-symbol-and-systemctl-and-vsftpd)
- [golang多版本管理神器gvm](https://github.com/moovweb/gvm)