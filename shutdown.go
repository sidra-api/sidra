package main

import (
	"fmt"
	"github.com/sidra-api/sidra/dto"
	"log"
	"os"
	"path/filepath"
)

func cleanPluginSocket(plugins map[string]dto.Plugin) {
	for _, plugin := range plugins {
		file := filepath.Join("/tmp", plugin.Name+".sock")
		err := os.Remove(file)
		if err != nil {
			log.Printf("Failed to remove %s: %v\n", file, err)
		} else {
			fmt.Printf("Removed: %s\n", file)
		}
	}
}
