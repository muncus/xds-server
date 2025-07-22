# GRPC server integration with xDS

For full details of GRPC server integration, consult [gRFC A36](https://github.com/grpc/proposal/blob/master/A36-xds-for-servers.md).

Requirements:
- bootstrap file with `server_listener_resource_name_template`

- `envoy.config.listener.v3.Listener`
    - address must match the SocketAddress of the server listener.
    - forbidden settings:
        - `Listener.listener_filters`
        - `Listener.use_original_dst`
    - FilterChains - implementations may vary, will reject if there are unimplemented filters.
        - must have only one `envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager` and it must be last.
    - `envoy.extensions.filters.http.router.v3.Router` *must* be present in http_filters of HCM.

- HTTPFault filter is not implemented for servers.
    - TODO: followup on why