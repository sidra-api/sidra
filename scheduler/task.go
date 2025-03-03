package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sidra-api/sidra/dto"
	"io"
	"log"
	"net/http"
	"os"
)

func (j *Job) register() {
	if os.Getenv("dataplaneid") == "" {
		return
	}
	if _, err := os.Stat("/tmp/privatekey"); err == nil {
		return
	}
	body := map[string]string{
		"id": os.Getenv("dataplaneid"),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Default().Println("err", err)
	}
	// Make the API call
	url := j.controlPlaneHost + "/api/v1/register"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Default().Println("err", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		j.dataSet.ID = os.Getenv("dataplaneid")
		var response struct {
			PrivateKey string `json:"PrivateKey"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Default().Println("err", err)
		}
		err = os.WriteFile("/tmp/privatekey", []byte(response.PrivateKey), 0644)
		if err != nil {
			log.Default().Println("err", err)
		}
	}

}

func (j *Job) storeConfig() {
	if os.Getenv("dataplaneid") == "" {
		return
	}
	resp, err := http.Get(j.controlPlaneHost + "/api/v1/get/gs/" + os.Getenv("dataplaneid"))
	if err != nil {
		log.Default().Println("err", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var response struct {
			GatewayServices []string `json:"gateway_service_id"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Default().Println("err", err)
		}
		for _, gsID := range response.GatewayServices {
			fmt.Println("Gs Url", j.controlPlaneHost+"/api/v1/config/"+gsID)
			gsResp, err := http.Get(j.controlPlaneHost + "/api/v1/config/" + gsID)
			if err != nil {
				log.Default().Println("api err", err)
				continue
			}
			defer gsResp.Body.Close()
			if gsResp.StatusCode == http.StatusOK {
				gsData, err := io.ReadAll(gsResp.Body)
				if err != nil {
					log.Default().Println("err", err)
					continue
				}
				var gsDetail dto.GatewayServiceDetail
				err = json.Unmarshal(gsData, &gsDetail)
				if err != nil {
					log.Default().Println("err", err)
					continue
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
						UpstreamHost: route.UpstreamHost,
						UpstreamPort: route.UpstreamPort,
						Path:         route.Path,
						PathType:     route.PathType,
						Plugins:      route.Plugins,
						Expression:   route.Expression,
						CreatedAt:    route.CreatedAt,
						UpdatedAt:    route.UpdatedAt,
					}

				}
				for _, plugin := range gsDetail.Plugins {
					j.dataSet.Plugins[plugin.Name] = plugin
				}
			}
		}
	}
}
