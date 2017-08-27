package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/vladimirvivien/go-grpc/currency_grpc/curproto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// printUSD demonstrates simple binary call from client
func printUSD(client curproto.CurrencyServiceClient) {
	curReq := &curproto.CurrencyRequest{Code: "USD"}
	curList, err := client.GetCurrencyList(context.Background(), curReq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nUSD Countries")
	fmt.Println("-------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.Country, cur.Code)
	}
}

// printEUR demonstrates server stream call from client
func printEUR(client curproto.CurrencyServiceClient) {
	curReq := &curproto.CurrencyRequest{Code: "EUR"}
	stream, err := client.GetCurrencyStream(context.Background(), curReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nEUR Countries")
	fmt.Println("-------------")
	for {
		cur, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			continue
		}
		fmt.Printf("%-50s%-10s\n", cur.Country, cur.Code)
	}
}

// addCurrencies demonstrates client to server stream
func addCurrencies(client curproto.CurrencyServiceClient) {
	currencies := []*curproto.Currency{
		&curproto.Currency{Country: "HAITI", Name: "Gourde", Code: "HTG", Number: 332},
		&curproto.Currency{Country: "MARTINIQUE", Name: "Euro", Code: "EUR", Number: 978},
		&curproto.Currency{Country: "CUBA", Name: "Cuban Peso", Code: "CUP", Number: 192},
		&curproto.Currency{Country: "JAMAICA", Name: "Jamaican Dollar", Code: "JMD", Number: 388},
	}

	// get client stream
	stream, err := client.SaveCurrencyStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// stream Currency to save to server
	for _, cur := range currencies {
		if err := stream.Send(cur); err != nil {
			log.Fatal(err)
		}
	}

	// close stream and get saved currencies as CurrencyList as reply
	curList, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSaved currencies")
	fmt.Println("-----------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.Country, cur.Code)
	}
}

// findCurrencies demonstrates bi-directional stream: one direction streams
// requests to the server while receiving replies from the server.
func findCurrencies(client curproto.CurrencyServiceClient) {
	reqs := []*curproto.CurrencyRequest{
		&curproto.CurrencyRequest{Code: "CDF"},
		&curproto.CurrencyRequest{Code: "AZN"},
		&curproto.CurrencyRequest{Number: 392},
		&curproto.CurrencyRequest{Code: "QAR"},
		&curproto.CurrencyRequest{Number: 949},
	}

	stream, err := client.FindCurrencyStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// goroutine to stream outbound requests to server
	go func() {
		for _, req := range reqs {
			if err := stream.Send(req); err != nil {
				log.Fatal(err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatal(err)
		}
	}()

	// handle incoming Currency reponses from stream
	fmt.Println("\nFound Currencies")
	fmt.Println("-----------------")
	for {
		cur, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			continue
		}
		fmt.Printf("%-50s%-10s\n", cur.Country, cur.Code)
	}
}

func main() {
	server := "127.0.0.1:50051"

	// setup tls options
	tlsCreds, err := credentials.NewClientTLSFromFile("./../certs/ca.crt", "localhost")
	if err != nil {
		log.Fatal(err)
	}

	// setup insecure connection
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := curproto.NewCurrencyServiceClient(conn)

	printUSD(client)

	printEUR(client)

	addCurrencies(client)

	findCurrencies(client)

}
