package main

var (
	DefaultConfig = map[string]interface{}{
		"database_config": DatabaseConfig{
			HostName:     "mysql",
			Port:         3306,
			DatabaseName: "ditto",
			UserName:     "root",
			Password:     "root",
			Type:         "mysql",
		},
		"logging_config": LoggingConfig{
			LogLevel: "debug",
		},
		"heart_beat_config": HeartBeatConfig{
			KeepAliveTime:    10,
			KeepAliveTimeOut: 20,
		},
		"server_config": ServerConfig{
			Address:           "0.0.0.0",
			Port:              "7100",
			GatewayEnable:     true,
			GatewayAddress:    "0.0.0.0",
			GatewayPort:       "7101",
			GatewayURL:        "/ditto",
			InternalEnable:    true,
			InternalAddress:   "0.0.0.0",
			InternalPort:      "7102",
			InternalHealth:    "/health",
			InternalReadiness: "/readiness",
		},
	}
)

type DatabaseConfig struct {
	HostName     string
	Port         uint64
	DatabaseName string
	UserName     string
	Password     string
	Dsn          string
	Type         string
}

type LoggingConfig struct {
	LogLevel string
}

type HeartBeatConfig struct {
	KeepAliveTime    uint64
	KeepAliveTimeOut uint64
}

type ServerConfig struct {
	Address           string
	Port              string
	GatewayEnable     bool
	GatewayAddress    string
	GatewayURL        string
	GatewayPort       string
	InternalEnable    bool
	InternalAddress   string
	InternalPort      string
	InternalHealth    string
	InternalReadiness string
}

type PikachuConfig struct {
	DatabaseConfig  DatabaseConfig
	LoggingConfig   LoggingConfig
	HeartBeatConfig HeartBeatConfig
	ServerConfig    ServerConfig
}

