// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	example "github.com/muncus/xds-server"
	bsc "github.com/muncus/xds-server/pkg/bootstrap"
	"github.com/muncus/xds-server/pkg/grpcconfig"
)

var (
	l          example.Logger
	port       uint
	nodeID     string
	config     = flag.String("config", "", "bootstrap config to read and serve.")
	logLevel   = new(slog.LevelVar)
	clientFile = flag.String("grpcclients", "", "GRPC Client config to read and serve.")
)

func init() {
	l = example.Logger{}
	if l.Debug {
		logLevel.Set(slog.LevelDebug)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	_ = logger
	// slog.SetDefault(logger)
	flag.BoolVar(&l.Debug, "debug", false, "Enable xDS server debug logging")

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 18000, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")
}

func main() {
	flag.Parse()

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	var err error
	var bootstrapResources map[string][]types.Resource
	if len(*config) > 0 {
		bc, err := bsc.NewFromFile(*config)
		if err != nil {
			log.Fatal(err)
		}
		bootstrapResources, err = bsc.GetResources(bc)
		if err != nil {
			log.Fatal(err)
		}
	}

	// load grpc client info
	var clients grpcconfig.ClientConfigs
	if len(*clientFile) > 0 {
		clients, err = grpcconfig.NewClientsFromFile(*clientFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	clientResources := grpcconfig.GetResources(clients)
	mergedRes := example.MergeResourceMaps(bootstrapResources, clientResources)
	for _, t := range []string{resource.ListenerType, resource.ClusterType} {
		l.Infof("Dump of %s: %+v\n", t, mergedRes[t])
	}
	// Create the snapshot that we'll serve to Envoy
	snapshot := example.GenerateSnapshot(mergedRes)
	if err := snapshot.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}
	l.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := cache.SetSnapshot(context.Background(), nodeID, snapshot); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}

	// Run the xDS server
	ctx := context.Background()
	cb := &test.Callbacks{Debug: l.Debug}
	srv := server.NewServer(ctx, cache, cb)
	example.RunServer(srv, port)
}
