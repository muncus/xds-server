package grpcconfig

// types and methods for creating xDS Resources for GRPC Clients

import (
	"io"
	"log"
	"os"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v2"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
)

type ClientConfigs []ClientConfig
type ClientConfig struct {
	Name     string           `json:"name"`
	Backends []*BackendConfig `json:"backends"`
}

type BackendConfig struct {
	Locality corev3.Locality `json:"locality"`
	Address  string          `json:"address"`
	Port     int             `json:"port"`
	Weight   uint32          `json:"weight"`
}

func NewClientsFromFile(fname string) (ClientConfigs, error) {
	// load grpc client info
	clients := make([]ClientConfig, 0)
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	cdec := yaml.NewDecoder(f)
	for {
		var cc ClientConfig
		err := cdec.Decode(&cc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		clients = append(clients, cc)
	}
	return clients, nil
}

// GetResources creates necessary xds resources for a grpc client
// the returned map is keyed by resource type url.
func GetResources(clientConfigs ClientConfigs) map[string][]types.Resource {
	ret := make(map[string][]types.Resource, 0)

	for _, cc := range clientConfigs {
		cl := &clusterv3.Cluster{
			Name: cc.Name,
			ClusterDiscoveryType: &clusterv3.Cluster_Type{
				Type: clusterv3.Cluster_EDS,
			},
			EdsClusterConfig: &clusterv3.Cluster_EdsClusterConfig{
				EdsConfig: &corev3.ConfigSource{
					ConfigSourceSpecifier: &corev3.ConfigSource_Ads{},
				},
			},
		}
		ret[resource.ClusterType] = append(ret[resource.ClusterType], cl)

		l := makeClientListener(cc)
		ret[resource.ListenerType] = append(ret[resource.ListenerType], l)

		endpoints := make([]types.Resource, 0)
		for _, b := range cc.Backends {
			cla, err := makeEndpoint(cc.Name, b)
			if err != nil {
				log.Printf("failed to create endpoint %q: %v", b.Address, err)
			}
			endpoints = append(endpoints, cla)
		}
		ret[resource.EndpointType] = append(ret[resource.EndpointType], endpoints...)
	}
	return ret
}

func makeClientListener(cc ClientConfig) *listenerv3.Listener {
	// Note: we could un-nest these by using RDS here instead.
	router, _ := anypb.New(&routerv3.Router{})
	cm, _ := anypb.New(&hcm.HttpConnectionManager{
		CodecType: hcm.HttpConnectionManager_AUTO,
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &route.RouteConfiguration{
				Name: "default",
				VirtualHosts: []*route.VirtualHost{
					{
						Name:    "default",
						Domains: []string{"*"},
						Routes: []*route.Route{
							{
								Match: &route.RouteMatch{
									PathSpecifier: &route.RouteMatch_Prefix{},
									Grpc:          &route.RouteMatch_GrpcRouteMatchOptions{},
								},
								Action: &route.Route_Route{
									Route: &route.RouteAction{
										ClusterSpecifier: &route.RouteAction_Cluster{
											Cluster: cc.Name,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: "envoy.filters.http.router",
				ConfigType: &hcm.HttpFilter_TypedConfig{
					TypedConfig: router,
				},
			},
		},
	})
	l := &listenerv3.Listener{
		Name: cc.Name,
		ApiListener: &listenerv3.ApiListener{
			ApiListener: cm,
		},
	}

	return l
}

func makeEndpoint(cluster string, bec *BackendConfig) (*endpointv3.ClusterLoadAssignment, error) {
	// TODO: one CLA with all endpoints, or multiple CLAs?
	// TODO: which LB Weight do we want to use?
	ep := &endpointv3.ClusterLoadAssignment{
		ClusterName: cluster,
		Endpoints: []*endpointv3.LocalityLbEndpoints{
			{
				Locality: &bec.Locality,
				LbEndpoints: []*endpointv3.LbEndpoint{
					{
						HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
							Endpoint: &endpointv3.Endpoint{
								Address: &corev3.Address{
									Address: &corev3.Address_SocketAddress{
										SocketAddress: &corev3.SocketAddress{
											Protocol: corev3.SocketAddress_TCP,
											Address:  bec.Address,
											PortSpecifier: &corev3.SocketAddress_PortValue{
												PortValue: uint32(bec.Port),
											},
											ResolverName: "",
										},
									},
								},
							},
						},
						LoadBalancingWeight: wrapperspb.UInt32(bec.Weight),
					},
				},
				LoadBalancingWeight: wrapperspb.UInt32(uint32(bec.Weight)),
			},
		},
	}
	return ep, nil
}
