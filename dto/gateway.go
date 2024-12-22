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
	GatewayService GatewayService `json:"GatewayService"`
	Routes          []Route         `json:"Routes"`
	Consumers       []Consumer      `json:"Consumers"`
	Plugins         []Plugin        `json:"Plugins"`
}

type GatewayService struct {
	Host string `json:"host"`
}


type Route struct {
	ID           string `json:"id"`
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

type Consumer struct {
	ID        string `json:"id"`
	GatewayID string `json:"gateway_id"`
	PluginID  string `json:"plugin_id"`
	Username  string `json:"username"`
	CustomID  string `json:"custom_id"`
	Tags      string `json:"tags"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Plugin struct {
	ID         string `json:"id"`
	GatewayID  string `json:"gateway_id"`
	Name       string `json:"name"`
	TypePlugin string `json:"type_plugin"`
	Enabled    int    `json:"enabled"`
	Config     string `json:"config"`
	Protocols  string `json:"protocols"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}