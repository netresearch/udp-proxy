# udp-proxy

A small proxy that proxies UDP packets from one port to another depending on the source IP.

## Installation

You can download the binary for your platform from the [releases page](https://github.com/netresearch/udp-proxy/releases) or alternatively use the [Docker image](https://github.com/netresearch/udp-proxy/pkgs/container/udp-proxy).

## Usage

With the binary:

```
--port <port>     The port to listen on
--forward <source-ip>:<remote-ip>:<remote-port>
                  Forward packets from <source-ip> to <remote-ip>:<remote-port>, you can specify --forward multiple times
```

```sh
./udp-proxy --port 5000 --forward <source-ip>:<remote-ip>:<remote-port>
```

Or with Docker:

```sh
docker run --rm -p <listen-port>:<listen-port>/udp ghcr.io/netresearch/udp-proxy --port <listen-port> --forward <source-ip>:<remote-ip>:<remote-port>
```

You can also use the following SystemD service file to run the proxy as a service:

```ini
[Unit]
Description=UDP Proxy
Documentation=https://github.com/netresearch/udp-proxy
After=network.target
Requires=network.target

[Service]
Type=simple
ExecStart=/usr/bin/udp-proxy --port 5000 --forward <source-ip>:<remote-ip>:<remote-port>
Restart=always

[Install]
WantedBy=multi-user.target
```

## License

`udp-proxy` is licensed under the MIT license. See the included [LICENSE file](./LICENSE) for more information.
