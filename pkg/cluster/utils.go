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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"

	argocdapplicationv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argocdapplicationsetv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/applicationset/v1alpha1"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	yaml "sigs.k8s.io/yaml"
)

const (
	minNodePort = 30000
	maxNodePort = 32767
)

func readFile(filePath string) (data []byte, err error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	data, err = ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	return
}

func GetVclusterNames(filePath string) (vclusterNames []string, err error) {
	data, err := readFile(filePath)
	if err != nil {
		return vclusterNames, nil
	}
	data = []byte(ReplacePlaceholders(string(data), "{{vcluster}}", "vcluster"))
	return GetApplicationSetElements(data)
}

func GetHostClusterNames(filePath string) (hostClusterNames []string, err error) {
	data, err := readFile(filePath)
	if err != nil {
		return hostClusterNames, nil
	}

	data = []byte(ReplacePlaceholders(string(data), "{{cluster}}", "cluster"))
	str, err := GetApplicationSetElements(data)
	if err != nil {
		return
	}

	return str, nil
}

func GetApplicationSetElements(data []byte) ([]string, error) {
	var as argocdapplicationsetv1alpha1.ApplicationSet
	var elements []string
	err := yaml.Unmarshal(data, &as)
	if err != nil {
		return nil, err
	}

	for _, element := range as.Spec.Generators[0].List.Elements {
		var m map[string]interface{}
		err := json.Unmarshal(element.Raw, &m)
		if err != nil {
			return nil, err
		}
		clusterName, ok := m["cluster"].(string)
		if !ok {
			return nil, fmt.Errorf("unable to obtain cluster information")
		}
		elements = append(elements, clusterName)
	}

	return elements, nil
}

func AddOrUpdate(list []*RuntimeNameAndApiServer, item *RuntimeNameAndApiServer) []*RuntimeNameAndApiServer {
	for i, v := range list {
		if v.Name == item.Name {
			return append(append(list[:i:i], list[i+1:]...), item)
		}
	}
	return append(list, item)
}

func DeleteRuntimeByName(runtimeList []*RuntimeNameAndApiServer, name string) []*RuntimeNameAndApiServer {
	for i, r := range runtimeList {
		if r.Name == name {
			runtimeList = append(runtimeList[:i], runtimeList[i+1:]...)
			break
		}
	}
	return runtimeList
}

func ReplacePlaceholders(data string, placeholder, value string) string {
	replacements := map[string]string{
		placeholder: value,
	}

	for placeholder, value := range replacements {
		data = strings.ReplaceAll(data, placeholder, value)
	}

	return data
}

func AddIfNotExists(list []string, item string) []string {
	for _, v := range list {
		if v == item {
			return list
		}
	}
	return append(list, item)
}

func GenerateNodePort(usedPorts []int) int {
	for {
		port := rand.Intn(maxNodePort-minNodePort+1) + minNodePort
		if port >= minNodePort && port <= maxNodePort {
			if !contains(usedPorts, port) {
				return port
			}
		}
	}
}

func contains(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseArgocdApplication(fileName string) (*argocdapplicationv1alpha1.Application, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, nil
	}

	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var app argocdapplicationv1alpha1.Application
	if yaml.Unmarshal(bytes, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

func getTraefikHttpsNodePort(app *argocdapplicationv1alpha1.Application) (int, error) {
	var values struct {
		Ports struct {
			WebSecure struct {
				NodePort int `yaml:"nodePort"`
			} `yaml:"websecure"`
		} `yaml:"ports"`
	}

	if err := yaml.Unmarshal([]byte(app.Spec.Source.Helm.Values), &values); err != nil {
		return 0, fmt.Errorf("failed to unmarshal values YAML: %w", err)
	}

	return values.Ports.WebSecure.NodePort, nil
}

func parseCluster(fileName string) (*resourcev1alpha1.Cluster, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var cluster resourcev1alpha1.Cluster
	if err := yaml.Unmarshal(bytes, &cluster); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func getRuntimeType(cluster *resourcev1alpha1.Cluster) string {
	var physical = "physical"
	var virtual = "virtual"
	if ok := IsPhysicalRuntime(cluster); ok {
		return physical
	}

	return virtual
}
