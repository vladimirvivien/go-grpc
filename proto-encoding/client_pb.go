package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/vladimirvivien/go-grpc/proto-encoding/curproto"
)

func main() {
	resp, err := http.Get("http://localhost:4040/")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// read data from body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// unmarshal
	curList := new(curproto.CurrencyList)
	if err := proto.Unmarshal(data, curList); err != nil {
		log.Fatalln(err)
	}

	// display protobuf data
	for i, item := range curList.Items {
		fmt.Printf("%-25s%-20s\n", item.Name, item.Code)
		if i > 10 {
			fmt.Println("...")
			break
		}
	}
}
