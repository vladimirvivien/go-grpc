# gRPC Limits Example
This pacakge shows how add preventive measures to guard your
gRPC service and clients from failures by specifying limits.
There are several places where limits can prevent things from
going bad, including:

 * Limit # of TCP connections open for a service
 * Limit size of message server can receive
 * Limit size of message server can send
 * Limit # of concurrent streams multiplexing rcp requests
 * Limit size of message client can receive
 * Limit size of message client can send

#### Run Example
```sh
// start auth server
$> cd authsvc
$> go run *.go

// start currency server
$> cd grpc_limits
$> go run serv_limits.go

// run client
$> go run client_limits.go

```

