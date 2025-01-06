package dto

import "os"

type DataPlane struct {
	ID string `json:"id"`
	SerializeRoute map[string]SerializeRoute `json:"SerializeRoute"`
	Plugins map[string]Plugin `json:"Plugins"`
}

func NewDataPlane() *DataPlane {
	return &DataPlane{
		ID: os.Getenv("dataplaneid"),
		SerializeRoute: make(map[string]SerializeRoute),
		Plugins: make(map[string]Plugin),
	}
}

type SerializeRoute struct {
	ID           string `json:"id"`
	Host		 string `json:"host"`
	GatewayID    string `json:"gateway_id"`
	Name         string `json:"name"`
	Tags         string `json:"tags"`
	Methods      string `json:"methods"`
	UpstreamHost string `json:"upstream_host"`
	UpstreamPort string `json:"upstream_port"`
	Path         string `json:"path"`
	PathType     string `json:"path_type"`
	Plugins      string `json:"plugins"`
	Expression   string `json:"expression"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}


type GatewayServiceDetail struct {
	GatewayService  GatewayService  `json:"GatewayService" yaml:"GatewayService"`
	Routes          []Route         `json:"Routes" yaml:"Routes"`
	Consumers       []Consumer      `json:"Consumers" yaml:"Consumers"`
	Plugins         []Plugin        `json:"Plugins" yaml:"Plugins"`
}

type GatewayService struct {
	Host string `json:"host" yaml:"host"`
}


type Route struct {
	ID           string `json:"id" yaml:"id"`
	GatewayID    string `json:"gateway_id" yaml:"gateway_id"`
	Name         string `json:"name" yaml:"name"`
	Tags         string `json:"tags" yaml:"tags"`
	Methods      string `json:"methods" yaml:"methods"`
	UpstreamHost string `json:"upstream_host" yaml:"upstream_host"`
	UpstreamPort int `json:"upstream_port" yaml:"upstream_port"`
	Path         string `json:"path" yaml:"path"`
	PathType     string `json:"pathType" yaml:"path_type"`
	Plugins      []string `json:"plugins" yaml:"plugins"`
	Expression   string `json:"expression" yaml:"expression"`
	CreatedAt    string `json:"created_at" yaml:"created_at"`
	UpdatedAt    string `json:"updated_at" yaml:"updated_at"`
}

type Consumer struct {
	ID        string `json:"id" yaml:"id"`
	GatewayID string `json:"gateway_id" yaml:"gateway_id"`
	PluginID  string `json:"plugin_id" yaml:"plugin_id"`
	Username  string `json:"username" yaml:"username"`
	CustomID  string `json:"custom_id" yaml:"custom_id"`
	Tags      string `json:"tags" yaml:"tags"`
	CreatedAt string `json:"created_at" yaml:"created_at"`
	UpdatedAt string `json:"updated_at" yaml:"updated_at"`
}

type Plugin struct {
	ID         string `json:"id" yaml:"id"`
	GatewayID  string `json:"gateway_id" yaml:"gateway_id"`
	Name       string `json:"name_plugin" yaml:"name"`
	TypePlugin string `json:"type_plugin" yaml:"type_plugin"`
	Enabled    int    `json:"enabled" yaml:"enabled"`
	Config     string `json:"config" yaml:"config"`
	Protocols  string `json:"protocols" yaml:"protocols"`
	CreatedAt  string `json:"created_at" yaml:"created_at"`
	UpdatedAt  string `json:"updated_at" yaml:"updated_at"`
}