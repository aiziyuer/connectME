package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/gogf/gf/util/gconv"
	"go.uber.org/zap"
)

type Strategy struct {
	name string
}

var GLOBAL_DNS_STRATEGY_MAP map[string]*Strategy

func init() {

	_ = `

	// toml
	// 各类DNS的策略定义
	[strategies.doh]
	kind=doh
	proxy=socks5://xxxx:xxxx
	endpoint=https://dns.google/resolve?edns_client_subnet=115.192.208.0/24

	[strategies.trandition]
	kind=tcp
	proxy=
	endpoint=114.114.114.114:53

	// 某些域名固定结果
	[rules.address]
	dns.google=8.8.8.8
	dns.google.com=8.8.8.8

	// 某些域名使用不同的DNS解析策略
	[rules.server]
	.=doh
	.aiziyuer.familyds.com.=trandition
	.familyds.com.=trandition

	`

	// 解析出支持的策略
	doh := &Strategy{name: "doh"}
	trandition := &Strategy{name: "trandition"}

	// 解析基于域名的策略并形成映射表
	GLOBAL_DNS_STRATEGY_MAP = map[string]*Strategy{
		".":                     doh, // default
		"aiziyuer.familyds.com": trandition,
	}

}

func GetStrategy(queries ...string) *Strategy {

	for _, query := range queries {
		value, ok := GLOBAL_DNS_STRATEGY_MAP[query]
		if ok {
			return value
		}
	}

	return GLOBAL_DNS_STRATEGY_MAP["."]
}

func TestDNSStrategy(*testing.T) {

	// code.aiziyuer.familyds.com
	domain := `code.aiziyuer.familyds.com`
	domain = regexp.MustCompile(`[\.]{2,}`).ReplaceAllString(fmt.Sprintf(".%s.", domain), ".")

	// 生成域名查询(从长到短)
	queries := make([]string, 0)
	for q := domain; len(q) > 0; {
		queries = append(queries, q)
		q = regexp.MustCompile(`^\.[^.]*`).ReplaceAllString(q, "")
	}

	zap.S().Info(queries)

	// 查询策略
	currentStrategy := GetStrategy(queries...)
	zap.S().Info(gconv.String(currentStrategy))

}
