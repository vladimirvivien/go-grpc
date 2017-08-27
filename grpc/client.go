package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/vladimirvivien/go-grpc/protobuf"

	"google.golang.org/grpc"
)

const (
	server     = "127.0.0.1"
	serverPort = "50051"
)

// printUSD demonstrates simple binary call from client
func printUSD(client pb.CurrencyServiceClient) {
	curReq := &pb.CurrencyRequest{Code: "USD"}
	curList, err := client.GetCurrencyList(context.Background(), curReq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nUSD Countries")
	fmt.Println("-------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// printEUR demonstrates server stream call from client
func printEUR(client pb.CurrencyServiceClient) {
	curReq := &pb.CurrencyRequest{Code: "EUR"}
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
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// addCurrencies demonstrates client to server stream
func addCurrencies(client pb.CurrencyServiceClient) {
	currencies := []*pb.Currency{
		&pb.Currency{Country: "HAITI", Name: "Gourde", Code: "HTG", Number: 332},
		&pb.Currency{Country: "MARTINIQUE", Name: "Euro", Code: "EUR", Number: 978},
		&pb.Currency{Country: "CUBA", Name: "Cuban Peso", Code: "CUP", Number: 192},
		&pb.Currency{Country: "JAMAICA", Name: "Jamaican Dollar", Code: "JMD", Number: 388},
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

	// close stream and get saved currencies as pb.CurrencyList as reply
	curList, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSaved currencies")
	fmt.Println("-----------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// findCurrencies demonstrates bi-directional stream: one direction streams
// requests to the server while receiving replies from the server.
func findCurrencies(client pb.CurrencyServiceClient) {
	reqs := []*pb.CurrencyRequest{
		&pb.CurrencyRequest{Code: "CDF"},
		&pb.CurrencyRequest{Code: "AZN"},
		&pb.CurrencyRequest{Number: 392},
		&pb.CurrencyRequest{Code: "QAR"},
		&pb.CurrencyRequest{Number: 949},
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
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

func main() {
	serverAddr := net.JoinHostPort(server, serverPort)

	// setup insecure connection
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewCurrencyServiceClient(conn)

	printUSD(client)

	printEUR(client)

	addCurrencies(client)

	findCurrencies(client)
}
