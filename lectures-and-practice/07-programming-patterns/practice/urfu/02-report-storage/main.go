package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Storage interface {
	Save(name string, data []byte) error
}

type FileStorage struct {
	Dir string
}

func (s FileStorage) Save(name string, data []byte) error {
	return os.WriteFile(filepath.Join(s.Dir, name), data, 0o600)
}

type ArchiveService struct {
	storage Storage
}

func (s ArchiveService) Save(reportID string, data []byte) error {
	if reportID == "" {
		return errors.New("report ID is required")
	}
	return s.storage.Save(reportID+".txt", data)
}

func run() error {
	dir, err := os.MkdirTemp("", "report-storage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	service := ArchiveService{storage: FileStorage{Dir: dir}}
	if err := service.Save("weekly-hours", []byte("Platform: 32 hours\nSupport: 18 hours\n")); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(dir, "weekly-hours.txt"))
	if err != nil {
		return err
	}
	fmt.Printf("weekly-hours.txt\n%s", data)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
