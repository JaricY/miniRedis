package config

import (
	"miniRedis/lib/utils"
	"time"
)

var (
	ClusterMode    = "cluster"
	StandaloneMode = "standalone"
)

// ServerProperties 定义了Redis服务器全局的配置
type ServerProperties struct {
	RunID             string `cfg:"runid"`               // 每次启动 Redis 服务器时，都会生成一个唯一的 RunID。
	Bind              string `cfg:"bind"`                // 服务器绑定的 IP 地址。
	Port              int    `cfg:"port"`                // 服务器绑定的端口号。
	AppendOnly        bool   `cfg:"appendonly"`          // 是否开启 AOF 持久化。
	AppendFilename    string `cfg:"appendfilename"`      // AOF 持久化日志的文件名。
	AppendFsync       string `cfg:"appendfsync"`         // AOF 持久化的同步策略。
	MaxClients        int    `cfg:"maxclients"`          // 服务器能够处理的最大客户端连接数。
	RequirePass       string `cfg:"requirepass"`         // 连接 Redis 服务器所需的密码。
	Databases         int    `cfg:"databases"`           // Redis 服务器支持的数据库数。
	RDBFilename       string `cfg:"dbfilename"`          // RDB 持久化的文件名。
	MasterAuth        string `cfg:"masterauth"`          // 主从复制模式下从服务器连接主服务器的密码。
	SlaveAnnouncePort int    `cfg:"slave-announce-port"` // 从服务器向主服务器宣告自己的端口号。
	SlaveAnnounceIP   string `cfg:"slave-announce-ip"`   // 从服务器向主服务器宣告自己的 IP 地址。
	ReplTimeout       int    `cfg:"repl-timeout"`        //主从复制模式下复制超时时间。

	// 集群模式下的配置属性
	ClusterEnabled string   `cfg:"cluster-enabled"` // 是否开启集群模式。
	Peers          []string `cfg:"peers"`           // Redis 集群中所有节点的 IP 地址和端口号。
	Self           string   `cfg:"self"`            // 当前节点的 IP 地址和端口号

	// 配置文件的路径。
	CfPath string `cfg:"cf,omitempty"`
}

type ServerInfo struct {
	StartUpTime time.Time
}

var Properties *ServerProperties

var EachTimeServerInfo *ServerInfo

func init() {
	EachTimeServerInfo = &ServerInfo{
		StartUpTime: time.Now(),
	}

	Properties = &ServerProperties{
		Bind:       "127.0.0.1",
		Port:       6379,
		AppendOnly: false,
		RunID:      utils.RandString(40),
	}
}
