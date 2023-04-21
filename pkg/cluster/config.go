// Copyright 2023 Nautes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "sigs.k8s.io/yaml"
)

const (
	ClusterfilterFileName = "clusterignorerule"
)

type Config struct {
	Save   Save         `yaml:"save"`
	Remove Remove       `yaml:"remove"`
	Common CommonConfig `yaml:"common"`
}

type Save struct {
	HostCluster     HostClusterConfig `yaml:"hostCluster"`
	PhysicalRuntime RuntimeConfig     `yaml:"physicalRuntime"`
	VirtualRuntime  RuntimeConfig     `yaml:"virtualRuntime"`
}

type Remove struct {
	HostCluster     HostClusterConfig `yaml:"hostCluster"`
	PhysicalRuntime RuntimeConfig     `yaml:"physicalRuntime"`
	VirtualRuntime  RuntimeConfig     `yaml:"virtualRuntime"`
}

type HostClusterConfig struct {
	IgnorePath []string `yaml:"ignorePath"`
	IgnoreFile []string `yaml:"ignoreFile"`
}

type RuntimeConfig struct {
	IgnorePath []string `yaml:"ignorePath"`
	IgnoreFile []string `yaml:"ignoreFile"`
}

type CommonConfig struct {
	IgnorePath []string `yaml:"ignorePath"`
	IgnoreFile []string `yaml:"ignoreFile"`
}

func NewClusterFileConfig(dir string) (*Config, error) {
	file := fmt.Sprintf("%s/%s.yaml", dir, ClusterfilterFileName)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config *Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetSaveHostClusterConfig() (ignorePath, ignoreFile []string) {
	ignorePath = append(ignorePath, c.Save.HostCluster.IgnorePath...)
	ignorePath = append(ignorePath, c.Common.IgnorePath...)
	ignoreFile = append(ignoreFile, c.Save.HostCluster.IgnoreFile...)
	ignoreFile = append(ignoreFile, c.Common.IgnoreFile...)

	return
}

func (c *Config) GetSavePhysicalRuntimeConfig() (ignorePath, ignoreFile []string) {
	ignorePath = append(ignorePath, c.Save.PhysicalRuntime.IgnorePath...)
	ignorePath = append(ignorePath, c.Common.IgnorePath...)
	ignoreFile = append(ignoreFile, c.Save.PhysicalRuntime.IgnoreFile...)
	ignoreFile = append(ignoreFile, c.Common.IgnoreFile...)

	return
}

func (c *Config) GetSaveVirtualRuntimeConfig() (ignorePath, ignoreFile []string) {
	ignorePath = append(ignorePath, c.Save.VirtualRuntime.IgnorePath...)
	ignorePath = append(ignorePath, c.Common.IgnorePath...)
	ignoreFile = append(ignoreFile, c.Save.VirtualRuntime.IgnoreFile...)
	ignoreFile = append(ignoreFile, c.Common.IgnoreFile...)

	return
}

func (c *Config) GetRemoveHostClusterConfig() (ignorePath, ignoreFile []string) {
	ignorePath = append(ignorePath, c.Remove.HostCluster.IgnorePath...)
	ignorePath = append(ignorePath, c.Common.IgnorePath...)
	ignoreFile = append(ignoreFile, c.Remove.HostCluster.IgnoreFile...)
	ignoreFile = append(ignoreFile, c.Common.IgnoreFile...)

	return
}

func (c *Config) GetRemovePhysicalRuntimeConfig() (ignorePath, ignoreFile []string) {
	ignorePath = append(ignorePath, c.Remove.PhysicalRuntime.IgnorePath...)
	ignorePath = append(ignorePath, c.Common.IgnorePath...)
	ignoreFile = append(ignoreFile, c.Remove.PhysicalRuntime.IgnoreFile...)
	ignoreFile = append(ignoreFile, c.Common.IgnoreFile...)

	return
}

func (c *Config) GetRemoveVirtualRuntimeConfig() (ignorePath, ignoreFile []string) {
	ignorePath = append(ignorePath, c.Remove.VirtualRuntime.IgnorePath...)
	ignorePath = append(ignorePath, c.Common.IgnorePath...)
	ignoreFile = append(ignoreFile, c.Remove.VirtualRuntime.IgnoreFile...)
	ignoreFile = append(ignoreFile, c.Common.IgnoreFile...)

	return
}
