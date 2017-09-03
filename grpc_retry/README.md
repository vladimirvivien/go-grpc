# gRPC Retry Example
This pacakge shows how to use intecerptors to implement 
a simple retry logic.  The interceptor attemps a connection
to the server, if there's a failure condition, it looks at
the type of error to determine if it should retry the connection.

The retry is kept simple to illustrate how it works. More complex
retry implementations are available here https://github.com/grpc-ecosystem/go-grpc-middleware/tree/master/retry.

#### Run Example
```sh
// start auth server
$> cd authsvc
$> go run *.go

// start currency server
$> cd grpc_auth
$> go run serv_auth.go

// run client
$> go run client_auth.go

```