# gRPC Rate Limit Example
 Another protective measure that can be applied to services
 is to setup limits on the rate at which a service can be called 
 for a given period of times. Once the rate is exceeded, the server 
 forces the client to wait.

 This is done using an InTapHandle on the server to implement
 a rate limiter. The limiter can be associated with a particular
 user or it can be global.  The example in this package shows a
 global limiter that allowcates a large limit for all incoming 
 requests.

 Limits can also be applied at the client where the client
 enforces a limit that matches that of the clientn or even
 more restrictive.  When the client rate limit is reached,
 the code will cause all requests to wait until.


#### Run Example
```sh
// start auth server
$> cd authsvc
$> go run *.go

// start currency server
$> cd grpc_rate
$> go run serv_rate.go

// run client
$> go run client_rate.go

```