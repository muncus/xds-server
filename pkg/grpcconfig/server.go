package grpcconfig

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

//   listeners:
//   - name: "example/whereami-frontend"
//     address:
//       socket_address: { address: 0.0.0.0, port_value: 9091 }
//     filter_chains:
//     - filters:
//       - name: envoy.filters.network.http_connection_manager
//         typed_config:
//           "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
//           codec_type: AUTO
//           route_config:
//             name: local_route
//             virtual_hosts:
//             - name: local_service
//               domains: ["*"]
//               routes:
//               - match: { prefix: "" }
//                 non_forwarding_action:
//                 # route: { cluster: whereami-backend } # what's this for in server side routing?
//           http_filters:
//           - name: envoy.filters.http.router
//             typed_config:
//               "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

type ServerConfigs []ServerConfig
type ServerConfig struct {
	Name    string
	Address corev3.Address_SocketAddress
}

// NewServerConfigs parses GRPC Server configs from a file.
func NewServerConfigs(fname string) (ServerConfigs, error) {

	return nil, nil
}

func GetServerResources(sc ServerConfigs) map[string][]types.Resource {
	return nil

}
