package fleet

import (
	"bytes"
	"encoding/csv"
)

type CsvParser struct {
	headers []string
	records [][]string
}

func NewCsvParser(data string) (*CsvParser, error) {
	reader := csv.NewReader(bytes.NewReader([]byte(data)))
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return &CsvParser{
		headers: headers,
		records: records,
	}, nil
}

func (cp *CsvParser) GetRecords() []map[string]string {
	result := []map[string]string{}
	for _, record := range cp.records {
		entry := map[string]string{}
		for i, header := range cp.headers {
			if i < len(record) {
				entry[header] = record[i]
			} else {
				entry[header] = ""
			}
		}
		result = append(result, entry)
	}
	return result
}
