# GateH8: A Configurable API Gateway

GateH8 is a flexible and customizable API Gateway designed to proxy requests to different backends based on a JSON configuration. With the ability to utilize path variables, path wildcards, and domain wildcards, the gateway offers fine-grained control over routing behaviors, embodying simplicity and flexibility.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration Guide](#configuration-guide)
    - [General Settings](#general-settings)
    - [Path Variables and Wildcards](#path-variables-and-wildcards)
    - [Virtual Hosts and Routes](#virtual-hosts-and-routes)
    - [Wildcard Domain Routing](#wildcard-domain-routing)
    - [CORS Settings](#cors-settings)
    - [TLS Configuration for Secure Connections](#tls-configuration-for-secure-connections)
- [Running the Service](#running-the-service)
- [Contributing](#contributing)
- [License](#license)

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/yarlson/GateH8.git
    cd GateH8
    ```

2. Build the binary:

    ```bash
    go build -o gateh8 cmd/main.go 
    ```

## Quick Start

1. Create your `config.json` in the root directory.

2. Define your routes and backends as detailed in the [Configuration Guide](#configuration-guide).

3. Run the Gateway:

    ```bash
    ./gateh8 -a [address:port] # Optional: Use the -a or --addr flags to specify the server address and port.
    ```

4. Your API Gateway is up and listening on port 1973. Direct your requests accordingly.

## Configuration Guide

### General Settings

Your `config.json` is the central configuration for GateH8. This is where you define all routing behaviors.

```json
{
  "apiGateway": {
    "name": "MyAPIGateway",
    "version": "1.0.0"
  },
  ...
}
```

### Path Variables and Wildcards

GateH8 allows you to dynamically inject the requested path into your backend route using the `${path}` variable. This helps in scenarios where you want to forward the incoming request's path to the backend service without redefining it.

```json
{
  ...
  "endpoints": [
    {
      "path": "/endpoint2",
      "methods": ["POST"],
      "backend": {
        "url": "http://backend-service-2.com${path}",
        "timeout": 10000
      }
    }
  ]
}
```

Additionally, GateH8 supports wildcards within path configurations. By using an asterisk (`*`) in your path, you can match a variety of incoming request paths. For example:

```json
{
  ...
  "endpoints": [
    {
      "path": "/products/*",
      ...
    }
  ]
}
```

The above configuration will match and route requests like `/products/1`, `/products/soap`, and so on.

### Virtual Hosts and Routes

Virtual hosts enable you to route traffic differently based on the domain of the incoming request.

```json
{
  ...
  "vhosts": {
    "api.domain.com": {
      ...
      "endpoints": [
        ...
      ]
    },
    ...
  }
}
```

### Wildcard Domain Routing

GateH8 provides support for wildcard domain routing, allowing you to capture all requests to undefined hosts or to host patterns.

For instance:

- `"*.domain.com"` will capture all subdomains of `domain.com`.
- `"*"` will capture all hosts that aren't defined explicitly in the configuration.

### CORS Settings

To configure Cross-Origin Resource Sharing (CORS) for either the entire virtual host or specific endpoints:

```json
{
  ...
  "vhosts": {
    "api.domain.com": {
      "cors": {
        ...
      },
      ...
    },
    ...
  }
}
```

_Note_: CORS settings for an endpoint will override CORS settings for its parent virtual host.

### TLS Configuration for Secure Connections

GateH8 supports secure HTTPS connections via TLS. This can be configured on a per-virtual host basis using the SSL settings.

In your `config.json`, specify the paths to the SSL certificate and its corresponding key for each virtual host that needs to be served over HTTPS.

```json
{
  ...
  "vhosts": {
    "secure.domain.com": {
      "tls": {
        "cert": "path/to/cert.pem",
        "key": "path/to/private-key.pem"
      },
      ...
    },
    ...
  }
}
```

#### Important Points:

- If **all** vhosts have SSL configurations, the gateway will exclusively use HTTPS.
- If **none** of the vhosts have SSL configurations, the gateway will use plain HTTP.
- If **some** vhosts are configured with SSL and some are not, the gateway will panic during startup, as this is considered an inconsistent configuration. Ensure that either all vhosts are secured, or none of them are.

For added security, it's recommended to secure all vhosts. This not only ensures data encryption during transit but also provides trust and confidence to your API users.

## Running the Service

Once you've set up your `config.json`, simply execute the built binary:

```bash
./gateh8 -a [address:port] # Optional: Use the -a or --addr flags to specify the server address and port.
```

To get help regarding available flags:
```bash
./gateh8 -h
```

Your API Gateway will start listening on port 1973.

## Contributing

Your contributions are always welcome! Please fork the repository and create a pull request with your changes.

## License

GateH8 is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
