# tcp-bridge

Simple TCP forwarding utility, written in Go.

## Usage

```
tcp-bridge: [-h|--help] HOST PORTS...

Forwards all TCP connections to HOST, from a list of PORTS.

Arguments:
    HOST        destination host (e.g. "example.com", "192.168.1.42")
    PORTS       space-delimited list of ports to be forwarded (e.g. "8080")
                or pairs of port mappings (e.g. \"80->8080 443->8443\")
```

### Examples

- `tcp-bridge 192.168.1.150 8080 8081 8082` - forwards all local connections
  from ports `8080`, `8081` and `8082` to to remote host `192.168.1.150`
    
- `tcp-bridge example.com 8080-80 8443-443` - forwards all local connections
  from port `8080` to `example.port:80`, and from port `8443` to
  `example.com:443`

## Installation

Prerequisites:
- [Go compiler](https://go.dev/doc/install)

### go run

The simplest way to run `tcp-bridge` is using `go run`, e.g.:

```
go run github.com/paskozdilar/tcp-bridge@latest 192.168.1.150 8080 8081-8082
```

### GOPATH

If you've [set up GOPATH](https://go.dev/wiki/SettingGOPATH), you can install
the binary locally:

```
go install github.com/paskozdilar/tcp-bridge@latest
```

### Compile from source

To compile tcp-bridge from source:

```bash
    git clone https://github.com/paskozdilar/tcp-bridge.git
    cd tcp-bridge
    go build
```
