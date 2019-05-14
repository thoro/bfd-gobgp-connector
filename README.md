# bfd-gobgp-connector

Connects the bfd application with the gobgpd application. Enables/disables the specific peers on gobgp-side on session state changes by bfd.

### Dependency managenment

`go-dep`

To get started, simply type `dep ensure`

### Configuration

Sample config:

```
bfd-host: localhost:54211
gobgp-host: localhost:50051

peers:
  - bfd: peer1   # Name of peer as specified in bfdd-config
    bgp: 10.0.255.1
  - bfd: peer2
    bgp: 10.0.255.2
```
