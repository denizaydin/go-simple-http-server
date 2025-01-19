# go-simple-http-server

Simple Go HTTP server that responds with IP and HTTP credentials. It can also be used in Kubernetes environments to get node and pod details in the response. The server tries to add the hostname of the system it retrieves from the OS.

## Environment Variables

- `PORT`: The port on which the server listens. Defaults to `8080` if not set.
- `DESTINATION_PORT`: The destination port used in the response. Defaults to `8080` if not set.
- `NODE_NAME`: The name of the node, typically set in a Kubernetes environment.
- `POD_NAME`: The name of the pod, typically set in a Kubernetes environment.
- `IP_MODE`: The IP mode for the server. Valid values are `ipv4` and `ipv6``dualstack`. Defaults to all addresses if not set.

## How the Code Works

1. **Environment Variables**: The server retrieves configuration from environment variables. If not set, it uses default values.
2. **Hostname Retrieval**: The server attempts to retrieve the system's hostname.
3. **Destination Port Validation**: The server checks if the destination port is valid. If not, it falls back to the default port.
4. **Request Handling**: The server handles incoming HTTP requests and constructs a response that includes:
   - Node name
   - Pod name
   - Hostname
   - Destination address
   - Full requested URL
   - Incoming HTTP headers
5. **Response Size Handling**: If a specific response size is requested via the URL path, the server fills the response to meet the requested size, adding filler content if necessary.
6. **IP Mode Configuration**: The server listens on IPv4, IPv6, or both based on the `IP_MODE` environment variable.

## Example Usage

To run the server with default settings:

```sh
PORT=8080 DESTINATION_PORT=8080 NODE_NAME=my-node POD_NAME=my-pod IP_MODE=dualstack go run main.go
```
