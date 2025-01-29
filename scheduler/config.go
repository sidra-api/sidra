package scheduler

import (
	"flag"
	"log"
	"os"

	"github.com/sidra-api/sidra/dto"
	"gopkg.in/yaml.v2"
)

func (j *Job) loadConfig() {
	var gsDetail dto.GatewayServiceDetail

	configPath := flag.String("config", "", "path to the config file")
	flag.Parse()
	if *configPath == "" {
		if _, err := os.Stat("/tmp/config.yaml"); err == nil {
			log.Default().Println("Config file found in /tmp/config.yaml")
			*configPath = "/tmp/config.yaml"
		}
	}
	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Default().Println(err, "Error reading the file")
	}
	// Unmarshal the YAML data into the config struct
	err = yaml.Unmarshal(data, &gsDetail)
	if err != nil {
		log.Default().Println(err, "Error unmarshalling the data")
	}
	for _, gs := range gsDetail.Plugins {
		j.dataSet.Plugins[gs.Name] = dto.Plugin{
			ID:         gs.ID,
			Name:       gs.Name,
			TypePlugin: gs.TypePlugin,
			Config:     gs.Config,
			Enabled:    gs.Enabled,
		}
	}
	for _, route := range gsDetail.Routes {
		key := gsDetail.GatewayService.Host + route.Path
		j.dataSet.SerializeRoute[key] = dto.SerializeRoute{
			ID:           route.ID,
			Host:         gsDetail.GatewayService.Host,
			GatewayID:    route.GatewayID,
			Name:         route.Name,
			Tags:         route.Tags,
			Methods:      route.Methods,
			UpstreamPort: route.UpstreamPort,
			UpstreamHost: route.UpstreamHost,
			Path:         route.Path,
			PathType:     route.PathType,
			Plugins:      route.Plugins,
			Expression:   route.Expression,
			CreatedAt:    route.CreatedAt,
			UpdatedAt:    route.UpdatedAt,
		}

	}
}
