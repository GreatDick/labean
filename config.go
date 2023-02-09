// Copyright (c) 2018, Kirill Ovchinnikov
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"strings"
)

type appConfig struct {
	Listen       string `json:"listen"`
	RealIPHeader string `json:"real_ip_header"`
	ServerIp     net.IP `json:"external_ip"`
	ServerIpv6   net.IP `json:"external_ipv6"`
	TasksRaw     []task `json:"tasks"`
	Tasks        map[string]*task
	UrlPrefix    string `json:"url_prefix"`
	ExplicitIP   bool   `json:"allow_explicit_ips"`
}

func loadConfig(filename string) (newConfig *appConfig, err error) {
	rawConfig, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	var config appConfig
	err = json.Unmarshal([]byte(rawConfig), &config)
	if err != nil {
		return
	}

	if len(config.TasksRaw) == 0 {
		err = errors.New("No tasks set in config, I have nothing to do")
		return
	}

	config.Tasks = make(map[string]*task)
	for i, cmd := range config.TasksRaw {
		if cmd.Name == "" {
			err = errors.New("Some tasks' names are missing")
			return
		}
		if strings.Index(cmd.Name, "/") > -1 {
			err = errors.New("Task name cannot contain '/' symbol")
			return
		}
		if strings.Contains(cmd.TurnOn, "{serverIP}") || strings.Contains(cmd.TurnOff, "{serverIP}") {
			if config.ServerIp == nil {
				err = errors.New("You have {serverIP}, but didn't set 'external_ip' in config")
				return
			}
		}
		if strings.Contains(cmd.TurnOnIpV6, "{serverIP}") || strings.Contains(cmd.TurnOffIpV6, "{serverIP}") {
			if config.ServerIpv6 == nil {
				err = errors.New("You have {serverIP}, but didn't set 'external_ipv6' in config")
				return
			}
		}

		config.Tasks[strings.ToLower(cmd.Name)] = &config.TasksRaw[i]
	}
	config.TasksRaw = nil // we don't need it anymore

	if config.UrlPrefix != "" && config.UrlPrefix[0] != '/' {
		config.UrlPrefix = "/" + config.UrlPrefix
	}
	return &config, nil
}
