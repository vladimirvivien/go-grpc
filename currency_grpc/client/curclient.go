package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vladimirvivien/go-grpc/currency_grpc/curproto"

	"google.golang.org/grpc"
)

func main() {
	server := "127.0.0.1:50051"

	// setup insecure connection
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := curproto.NewCurrencyServiceClient(conn)
	curr, err := client.GetCurrency(context.Background(), &curproto.CurrencyRequest{Code: "USD"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(curr)

}
