package util

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"sync"

	pb "github.com/vladimirvivien/go-grpc/protobuf"
)

type DataStore struct {
	mtx      sync.Mutex
	dataFile string
	data     []*pb.Currency
}

func NewDataStore(file string) *DataStore {
	return &DataStore{dataFile: file}
}

func (ds *DataStore) Load() error {
	data, err := LoadPbFromCsv(ds.dataFile)
	if err != nil {
		return err
	}
	ds.data = data
	return nil
}

func (ds *DataStore) Search(code string, number int32) []*pb.Currency {
	var items []*pb.Currency
	for _, cur := range ds.data {
		if cur.GetNumber() == number || cur.GetCode() == code {
			items = append(items, cur)
		}
	}
	return items
}

func (ds *DataStore) Add(items []*pb.Currency) {
	ds.mtx.Lock()
	copy(ds.data, items)
	ds.mtx.Unlock()
}

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
