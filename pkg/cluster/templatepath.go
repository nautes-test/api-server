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

import "fmt"

func GetHostClusterAppsetFilePath(dir string) string {
	return fmt.Sprintf("%s/tenant/production/host-cluster-appset.yaml", dir)
}

func GetHostClustesrDir(dir string) string {
	return fmt.Sprintf("%s/host-clusters", dir)
}

func GetRuntimesDir(dir string) string {
	return fmt.Sprintf("%s/runtimes", dir)
}

func GetClustersDir(dir string) string {
	return fmt.Sprintf("%s/nautes/overlays/production/clusters", dir)
}

func GetTenantProductionDir(dir string) string {
	return fmt.Sprintf("%s/tenant/production", dir)
}

func GetVclustersDir(dir, subDir string) string {
	return fmt.Sprintf("%s/host-clusters/%s/vclusters", dir, subDir)
}
