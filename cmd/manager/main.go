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

package main

import (
	"flag"
	"os"

	"github.com/nautes-labs/api-server/internal/conf"
	"github.com/nautes-labs/api-server/pkg/kubernetes"
	"github.com/nautes-labs/api-server/pkg/nodestree"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	cluster "github.com/nautes-labs/api-server/pkg/cluster"
	"github.com/nautes-labs/pkg/pkg/log/zap"
	nautesconfigs "github.com/nautes-labs/pkg/pkg/nautesconfigs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(zap.NewLogger(),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	client, err := kubernetes.NewClient()
	if err != nil {
		panic(err)
	}

	resources_layout, err := nodestree.NewConfig()
	if err != nil {
		panic(err)
	}

	fileOptions := &nodestree.FileOptions{
		IgnorePath:       []string{".git", "production"},
		ExclusionsSuffix: []string{".txt", ".md"},
	}

	nodesTree := nodestree.NewNodestree(fileOptions, resources_layout, client)

	globalconfigs, err := nautesconfigs.NewConfigInstanceForK8s("nautes", "nautes-configs", "")
	if err != nil {
		panic(err)
	}

	clusteroperator := cluster.NewClusterRegistration()

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger, nodesTree, globalconfigs, client, clusteroperator)
	if err != nil {
		panic(err)
	}

	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
