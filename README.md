# bfd-gobgp-connector

Connects the bfd application with the gobgpd application. Enables/disables the specific peers on gobgp-side on session state changes by bfd.

### Dependency managenment

`go-dep`

To get started, simply type `dep ensure`

### Configuration

Sample config:

```
logging:
  logfile: interconnector.log
  log-also-to-stdout: False

bfd-host: 172.17.0.2:54211
gobgp-host: 172.17.0.3:50051

peers:
  # bfd-peer: bgp-peer
  - peer1: 10.0.255.1   # Name of peer as specified in bfdd-config
  - peer2: 10.0.255.2
```

### License

[MIT](LICENSE)
