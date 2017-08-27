# gRPC Example
This is a very simple gRPC client and server example.  The server exposes several operations
which demonstrates all 4 call types of gRPC unary, client stream, server stream, and 
bi-directional streams as shown in the IDL below:

```
service CurrencyService {
    // GetCurrencyList  returns matching Currency values as list
    // Example of a unary call
    rpc GetCurrencyList(CurrencyRequest) returns (CurrencyList){}

    // GetCurrencyStream returns matching Currencies as a server stream
    // Example of using server to client stream.
    rpc GetCurrencyStream(CurrencyRequest) returns (stream Currency){}

    // SaveCurrencyStream sends multiple currencies to server to be saved
    // returns a list of saved currency.
    // Example of using client stream to server.
    rpc SaveCurrencyStream(stream Currency) returns (CurrencyList){}

    // FindCurrencyStream sends a stream of CurrencyRequest to server and returns
    // a stream of Currency values.
    // Example of bi-directional stream
    rpc FindCurrencyStream(stream CurrencyRequest) returns (stream Currency){}
}
```