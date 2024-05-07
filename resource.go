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

package example

import (
	"log"
	"maps"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

const (
	ClusterName  = "example_proxy_cluster"
	RouteName    = "local_route"
	ListenerName = "listener_0"
	ListenerPort = 10000
	UpstreamHost = "www.envoyproxy.io"
	UpstreamPort = 80
)

func MergeResourceMaps(a, b map[string][]types.Resource) map[string][]types.Resource {
	ret := maps.Clone(b)

	for k, v := range a {
		ret[k] = append(ret[k], v...)
	}

	return ret
}

func GenerateSnapshot(rm map[string][]types.Resource) *cache.Snapshot {
	snap, _ := cache.NewSnapshot("1", rm)

	refs := cache.GetAllResourceReferences(snap.Resources)
	log.Printf("References: %#v\n", refs[resource.EndpointType])
	res := snap.GetResources(resource.EndpointType)
	log.Printf("Resources: %#v\n", res)
	return snap
}
