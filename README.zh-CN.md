# MyIPTV

[English Version](./README.md)

`MyIPTV` 是一个把 IPTV 节目流从 `UDP` 组播转换为 `HTTP` 的小工具，你可以把它视为一个加强的（带管理界面的）[udpxy](https://github.com/pcherenkov/udpxy)。

## 硬件

要使用 `MyIPTV`，电视需要支持安装第三方应用。

需要用一台软路由设备安装 `MyIPTV`，`R2S`、`H28K` 等 200 多块钱的就足够了（我使用的是 `R2S Plus`，外壳比较热，据说 `H28K` 发热量低些）；使用旧电脑也行，但最好有两个网口（支持 WIFI 的话，只有一个网口也能用）。

## 网络

我是按下图部署网络的，软路由设备和光猫 `iTV` 口之间必须用网线连接，其他可以用 `WIFI`，但建议用网线。

![Network](network.png)

软路由接光猫 LAN 口的网卡建议使用固定 IP，我用的是 `192.168.1.2`，后文也将以此为例。

## 软件

软路由上建议装 `Ubuntu` 之类的通用操作系统，不要装 `openwrt` 之类的专用软路由系统，我一开始用的是 `FriendlyWrt`，死活不能用。

从 [这里](https://github.com/localvar/myiptv/releases)下载一个与你的软路由 CPU 架构对应的 `MyIPTV` 到你的软路由，然后执行类似下面的命令就可以运行 `MyIPTV` 了。

```shell
$ mv myiptv-v0.1.0-linux.arm64 myiptv
$ chmod +x myiptv
$ ./myiptv
```

如果你的 LAN 口 IP 地址以 `192.168.` 开头，`MyIPTV` 一般能自动正确检测到网络相关的配置；如果没有，你需要手工准备一个 `myiptv.json` 的配置文件并重启 `MyIPTV`，配置文件里至少需要包含以下内容：

```json
{
	"config": {
		"serverAddr": "192.168.1.2:7709",
		"mastIface": "eth0"
	}
}
```

其中，`serverAddr` 是 `MyIPTV` 对外提供服务的地址，`mcastIface` 是软路由接光猫 `iTV` 口的网卡的名称，请都按你的实际情况填写。

`MyIPTV` 成功启动后，在电脑上用浏览器访问 “http://{serverAddr}” （如 `http://192.168.1.2:7709`）就可以看到管理界面了，如果你没有手工配置 `serverAddr`，那么它的默认值是 “{软路由LAN口IP}:7709”，如 `192.168.1.2:7709`。

> **注意**：为防止 SSH 会话关闭时 `MyIPTV` 进程被杀掉，需要使用 `nohup` 方式运行或最好把 `MyIPTV` 做成一个 `daemon`，具体方法请自行搜索。

## 频道导入和导出

在“频道管理”界面，可以导入和导出频道列表，对应的文件是 `csv` 格式，其中以 `#` 开头的行为注释，下面是一个示例文件：

```
#频道组,频道名称,显示名称,是否隐藏,台标,节目源
央视,CCTV1,CCTV-1 综合,否,http://epg.51zmt.top:8000/tb1/CCTV/CCTV1.png,225.1.0.103:1025
央视,CCTV1,CCTV-1 综合,否,http://epg.51zmt.top:8000/tb1/CCTV/CCTV1.png,225.1.8.103:8002
央视,CCTV2,CCTV-2 财经,否,http://epg.51zmt.top:8000/tb1/CCTV/CCTV2.png,225.1.0.104:1025
央视,CCTV2,CCTV-2 财经,否,http://epg.51zmt.top:8000/tb1/CCTV/CCTV2.png,225.1.8.2:8084
北京,北京卫视,,否,http://epg.51zmt.top:8000/tb1/ws/beijing.png,225.1.0.111:1025
北京,北京卫视,,否,http://epg.51zmt.top:8000/tb1/ws/beijing.png,225.1.8.21:8002
```

## 看电视

`MyIPTV` 目前支持两种格式的频道列表，`TEXT` 和 `M3U8`。

如果你电视上安装的 IPTV 应用使用 `TEXT` 格式（比如 DIYP），则对应的频道列表链接为：`http://{serverAddr}/iptv/channels`，例如 `http://192.168.1.2:7709/iptv/channels`。

如果你电视上安装的 IPTV 应用使用 `M3U8` 格式（比如 Kodi），则对应的频道列表链接为：`http://{serverAddr}/iptv/channels?fmt=m3u8`，例如 `http://192.168.1.2:7709/iptv/channels?fmt=m3u8`。

电子节目单目前仅支持 DIYP 使用的 JSON 格式，对应的节目单链接为：`http://{serverAddr}/iptv/epg`，例如 `http://192.168.1.2:7709/iptv/epg`。

## DDNS

`MyIPTV` 内置了一个 Cloudflare 的 DDNS（但这并非必须功能）。所以，如果有公网 IP、域名，且使用 Cloudflare 做解析，就可以把 `MyIPTV` 发布到公网上去了。 当然，后果自负。

如果使用此功能，请参考下面的示例修改 `myiptv.json`：

```json
{
	"ddns": {
		// 需要使用 DDNS 的域名
		"recordName": "myiptv.example.com",
		// Cloudflare zone ID
		"zoneID": "xxxxxxxxxxxxxxxxxxxxxxxx",
		// Cloudflare API key
		"apiKey": "yyyyyyyyyyyyyyyyyyyyyyyy",
		// 用于获取本机公网 IP 的 url，如果这个 url 返回的是 IPv4 地址，将更新对应域名的 A 记录，
		// 如果返回的是 IPv6 地址，将更新对应的 AAAA 记录。
		"wanIPProviders": ["http://ipv4.icanhazip.com", "https://whatismyip.akamai.com"],
		// 用于加速 DNS 解析的服务器地址，可选，必须包含端口（一般是 53），
		// 一般应该设置为指向 Cloudflare 为你的域名指定的 DNS 服务器。
		"dnsServers": ["beth.ns.cloudflare.com:53", "rudy.ns.cloudflare.com:53"]
	}
}
```

## 开发

`MyIPTV` 前端是用 Vue + Ant Design Vue 开发的，我个人不擅长前端，所以只是完成了基本功能，希望有擅长前端的同学帮助改进。

后端使用的 Go，刻意没有使用任何第三方库，后续也将坚持这一原则。

UDP 数据包解析部分是从 udpxy 抄过来的，在此对 udpxy 的作者表示感谢。

后续可能开发的功能包括（但我不做任何保证）：

- 支持更多的频道列表格式。
- 支持更多的电子节目单格式。
- 自动安装为 daemon。
- 自动扫描节目源。
- UI 支持多语言。

最后，欢迎大家提 PR 一起开发。
