package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/vladimirvivien/go-grpc/proto-encoding/curproto"
)

const fileName = "data.pb"

func main() {
	currencyItems, err := createPbFromCsv("../curdata.csv")
	if err != nil {
		log.Fatalf("failed to load csv: %v\n", err)
	}
	// print on screen
	for i, item := range currencyItems.Items {
		fmt.Printf("%-25s%-20s\n", item.Name, item.Code)
		if i > 10 {
			fmt.Println("...")
			break
		}
	}
	// encode as protobuf data
	data, err := proto.Marshal(currencyItems)
	if err != nil {
		log.Fatal(err)
	}

	// save to file
	if err := ioutil.WriteFile(fileName, data, 0644); err != nil {
		log.Fatal(err)
	}
	log.Println("data file saved as", fileName)

}

// createPbFromCsv loads the currency data from csv into protobuf values
func createPbFromCsv(path string) (*curproto.CurrencyList, error) {
	items := make([]*curproto.Currency, 0)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create CSV reader from file
	reader := csv.NewReader(file)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		var num int32
		if i, err := strconv.Atoi(row[3]); err == nil {
			num = int32(i)
		}
		// create data row with protobuf
		c := &curproto.Currency{
			Country: row[0],
			Name:    row[1],
			Code:    row[2],
			Number:  num,
		}
		items = append(items, c)
	}
	return &curproto.CurrencyList{Items: items}, err
}
