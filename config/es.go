package config

import (
	"github.com/elastic/go-elasticsearch/v8"
	"log"
)

type ESConfig struct {
	Url string `mapstructure:"url"`
}

var ESClient *elasticsearch.Client

func InitES() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			G.ES.Url,
		},
		// 如果有用户名密码：
		// Username: "elastic",
		// Password: "your_password",
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ 初始化 Elasticsearch 客户端失败: %s", err)
	}
	ESClient = es

	// 测试连接
	res, err := es.Info()
	if err != nil {
		log.Fatalf("❌ 无法连接 Elasticsearch: %s", err)
	}
	defer res.Body.Close()

	log.Println("✔️ 成功连接 Elasticsearch")
}
