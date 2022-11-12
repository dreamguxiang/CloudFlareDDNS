# CloudFlareDDNS
自动更新CloudFlare DNS配置

## 用法
1. 首次运行生成config.json

```json
{
	"intervalTime": 60, //更新间隔，建议300s
	"Email": "Example@outlook.com",// CF邮箱
	"APIKey": "CLOUDFLARE_API_KEY",// 通用密钥，不知道的百度
	"Zones": [//这是个数组，可以匹配多个
		{
			"Name": "example.com",//区域名（域名）
			"Records": [//这是个数组，可以匹配多个
				{
					"Name": "home.example.com"//需要动态更新公网IP进行更新的dns记录，支持V4,V6
				}
			]
		}
	]
}

```

2. 配置完Json再次运行即可，建议先调成10s测试一遍~

## 备

- Win平台可以利用计划任务实现开机自启
- Release文件进行了upx压缩，如有问题请运行Build.cmd重新编译
- 已测试平台 `Win10`,`Win11`,`linux arm`

