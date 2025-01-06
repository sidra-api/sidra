package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func (j *Job) setupPlugin() {
	fmt.Println("[TASK]Setting up plugin..")
	if _, err := os.Stat("/tmp/privatekey"); err != nil {		
		return
	}
	fmt.Println("Installed Plugins number : ", len(j.dataSet.Plugins))
	for _, plugin := range j.dataSet.Plugins {
		fmt.Println("- checking plugin", plugin)
		basePath := ""
		if _, err := os.Stat("/usr/local/bin/plugin_" + plugin.TypePlugin); err != nil {
			if _, err := os.Stat("./plugins/plugin-" + plugin.TypePlugin); err != nil {
				fmt.Println("Plugin does not exist")
				continue
			} else {
				basePath = "./plugins/plugin-" + plugin.TypePlugin + "/"
			}
		}
		env := make(map[string]string)
		err := json.Unmarshal([]byte(plugin.Config), &env)
		if err != nil {
			fmt.Println("- Failed to parse plugin config", plugin.TypePlugin, plugin.Config, err)
			continue
		}
		env["PLUGIN_NAME"] = plugin.Name
		cmd := exec.Command(basePath+"plugin_plugin-"+plugin.TypePlugin, "")
		for key, value := range env {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
		fmt.Println("- Starting plugin", cmd.Env)
		err = cmd.Start()
		if err != nil {
			fmt.Println("- Failed to start plugin", plugin.TypePlugin, err)
			continue
		}
	}
}
