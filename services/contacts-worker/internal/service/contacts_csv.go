package service

import (
	"encoding/csv"
	"io"
)

func (cs *ContactsService) createCsvRowProvider(file io.Reader) (rowProvider, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	provider := func(jobs chan<- []string) error {
		for {
			record, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				continue
			}
			jobs <- record
		}
	}

	return provider, nil
}
