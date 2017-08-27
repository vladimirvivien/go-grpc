package util

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"

	pb "github.com/vladimirvivien/go-grpc/protobuf"
)

// LoadPbFromCsv loads the currency data from csv into protobuf values
func LoadPbFromCsv(path string) ([]*pb.Currency, error) {
	items := make([]*pb.Currency, 0)
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
		c := &pb.Currency{
			Country: row[0],
			Name:    row[1],
			Code:    row[2],
			Number:  num,
		}
		items = append(items, c)
	}
	return items, err
}
