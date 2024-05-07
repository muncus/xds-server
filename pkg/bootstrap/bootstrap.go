// package bootstrap handles the loading of envoy Bootstrap configs
package bootstrap

import (
	"io"
	"os"

	"github.com/bufbuild/protoyaml-go"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

func NewFromFile(fname string) (*bootstrapv3.Bootstrap, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	fb, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var bsc bootstrapv3.Bootstrap
	err = protoyaml.Unmarshal(fb, &bsc)
	if err != nil {
		return nil, err
	}
	return &bsc, nil
}

func GetResources(bs *bootstrapv3.Bootstrap) (map[string][]types.Resource, error) {
	var clusters []types.Resource
	var listeners []types.Resource
	var endpoints []types.Resource
	for _, c := range bs.GetStaticResources().GetClusters() {
		clusters = append(clusters, c)
		endpoints = append(endpoints, c.GetLoadAssignment())
	}
	for _, l := range bs.GetStaticResources().GetListeners() {
		listeners = append(listeners, l)
	}
	ret := make(map[string][]types.Resource)
	ret[resource.ClusterType] = clusters
	ret[resource.ListenerType] = listeners
	ret[resource.EndpointType] = endpoints
	return ret, nil

}
