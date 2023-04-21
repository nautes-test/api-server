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

package data

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/nautes-labs/api-server/internal/biz"
	gitlabclient "github.com/nautes-labs/api-server/pkg/gitlab"
	nautesconfigs "github.com/nautes-labs/pkg/pkg/nautesconfigs"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCodeRepo, NewSecretRepo, NewGitRepo, NewDexRepo)

func NewData(logger log.Logger, configs *nautesconfigs.Config) (func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return cleanup, nil
}

func NewGitRepo(config *nautesconfigs.Config) (biz.GitRepo, error) {
	return &gitRepo{config: config}, nil
}

func NewSecretRepo(config *nautesconfigs.Config) (biz.Secretrepo, error) {
	// Get secret platform type according to configuration information
	if config.Secret.RepoType == "vault" {
		return NewVaultClient(config)
	}

	return nil, fmt.Errorf("failed to generate secret repo")
}

func NewCodeRepo(config *nautesconfigs.Config) (biz.CodeRepo, error) {
	// Get secret platform type according to configuration information
	if config.Git.GitType == "gitlab" {
		operator := gitlabclient.NewGitlabOperator()
		return NewGitlabRepo(config.Git.Addr, operator)
	}

	return nil, nil
}

func NewDexRepo(k8sClient client.Client) biz.DexRepo {
	return &Dex{k8sClient: k8sClient}
}
