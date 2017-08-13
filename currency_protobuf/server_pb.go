package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/vladimirvivien/go-grpc/currency_protobuf/curproto"
)

var curList *curproto.CurrencyList

func currencies(resp http.ResponseWriter, req *http.Request) {

	pbData, err := proto.Marshal(curList)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		os.Exit(1)
	}
	resp.WriteHeader(http.StatusOK)
	resp.Write(pbData)
}

func main() {
	// load currencies from csv
	list, err := createPbFromCsv("./curdata.csv")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	curList = list

	// start server
	http.HandleFunc("/", currencies)
	fmt.Println("Starting server on port 4040")
	if err := http.ListenAndServe(":4040", nil); err != nil {
		fmt.Println(err)
	}
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
