package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
)

type Row struct {
	Team  string `json:"team"`
	Hours int    `json:"hours"`
}

type Storage interface {
	Save(name string, data []byte) error
}

type MemoryStorage struct {
	files map[string][]byte
}

func (s *MemoryStorage) Save(name string, data []byte) error {
	if s.files == nil {
		s.files = make(map[string][]byte)
	}
	s.files[name] = slices.Clone(data)
	return nil
}

type ExportService struct {
	storage Storage
}

func (s ExportService) Export(reportID string, rows []Row, format string) error {
	if reportID == "" {
		return errors.New("report ID is required")
	}
	if len(rows) == 0 {
		return errors.New("report has no rows")
	}

	var data []byte
	switch format {
	case "csv":
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)
		if err := writer.Write([]string{"team", "hours"}); err != nil {
			return err
		}
		for _, row := range rows {
			if err := writer.Write([]string{row.Team, strconv.Itoa(row.Hours)}); err != nil {
				return err
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return err
		}
		data = buffer.Bytes()
	case "json":
		var err error
		data, err = json.Marshal(rows)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format %q", format)
	}

	return s.storage.Save(reportID+"."+format, data)
}

func main() {
	storage := &MemoryStorage{}
	service := ExportService{storage: storage}
	rows := []Row{{Team: "Платформа", Hours: 32}, {Team: "Поддержка", Hours: 18}}

	for _, format := range []string{"csv", "json"} {
		if err := service.Export("weekly-hours", rows, format); err != nil {
			fmt.Println("Ошибка:", err)
			return
		}
		name := "weekly-hours." + format
		fmt.Printf("%s\n%s\n", name, storage.files[name])
	}
}
