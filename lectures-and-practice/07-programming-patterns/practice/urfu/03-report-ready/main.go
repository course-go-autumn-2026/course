package main

import (
	"errors"
	"fmt"
	"slices"
)

type Report struct {
	ID    string
	Title string
	Rows  int
}

type Storage interface {
	Save(reportID string, data []byte) error
}

type MemoryStorage struct {
	files map[string][]byte
}

func (s *MemoryStorage) Save(reportID string, data []byte) error {
	if s.files == nil {
		s.files = make(map[string][]byte)
	}
	s.files[reportID] = slices.Clone(data)
	return nil
}

type EmailSender struct {
	sent []string
}

func (s *EmailSender) Send(reportID string) {
	s.sent = append(s.sent, reportID)
	fmt.Printf("Email: report=%s\n", reportID)
}

type Dashboard struct {
	ready []string
}

func (d *Dashboard) MarkReady(reportID string) {
	d.ready = append(d.ready, reportID)
	fmt.Printf("Dashboard: report=%s\n", reportID)
}

type ReportService struct {
	storage   Storage
	email     *EmailSender
	dashboard *Dashboard
}

func (s ReportService) Generate(report Report) error {
	if report.ID == "" || report.Title == "" || report.Rows <= 0 {
		return errors.New("report ID, title and positive row count are required")
	}
	data := []byte(fmt.Sprintf("%s\nrows=%d\n", report.Title, report.Rows))
	if err := s.storage.Save(report.ID, data); err != nil {
		return err
	}
	s.email.Send(report.ID)
	s.dashboard.MarkReady(report.ID)
	return nil
}

func main() {
	service := ReportService{
		storage:   &MemoryStorage{},
		email:     &EmailSender{},
		dashboard: &Dashboard{},
	}
	if err := service.Generate(Report{ID: "weekly-hours", Title: "Рабочие часы", Rows: 2}); err != nil {
		fmt.Println("Ошибка:", err)
	}
}
