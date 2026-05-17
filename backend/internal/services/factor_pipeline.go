package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FactorPipelineStats struct {
	RowCount      int     `json:"row_count"`
	Min           float64 `json:"min"`
	Max           float64 `json:"max"`
	Mean          float64 `json:"mean"`
	Std           float64 `json:"std"`
	WarehousePath string  `json:"warehouse_path"`
	PublishedPath string  `json:"published_path"`
}

type FactorPipelineService struct {
	baseDir       string
	warehousePath string
	publishedPath string
}

func NewFactorPipelineService(baseDir string) *FactorPipelineService {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = filepath.Join("..", "data")
	}
	return &FactorPipelineService{
		baseDir:       baseDir,
		warehousePath: filepath.Join(baseDir, "warehouse.csv"),
		publishedPath: filepath.Join(baseDir, "published.csv"),
	}
}

func (s *FactorPipelineService) SaveAndPublishFactor(factorName string) (*FactorPipelineStats, string, error) {
	if strings.TrimSpace(factorName) == "" {
		return nil, "", errors.New("factor name is required")
	}
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return nil, "", err
	}

	rows, headers, err := s.loadWarehouse()
	if err != nil {
		return nil, "", err
	}
	if len(rows) == 0 {
		rows = s.bootstrapRows()
		headers = []string{"date"}
	}

	if contains(headers, factorName) {
		return nil, "", fmt.Errorf("factor %q already exists in warehouse", factorName)
	}
	headers = append(headers, factorName)

	sum := 0.0
	sumSq := 0.0
	minVal := math.MaxFloat64
	maxVal := -math.MaxFloat64

	for i := range rows {
		v := computeFactorValue(i, factorName)
		rows[i][factorName] = v
		sum += v
		sumSq += v * v
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	if err := s.writeCSV(s.warehousePath, headers, rows); err != nil {
		return nil, "", err
	}
	if err := s.writeCSV(s.publishedPath, headers, rows); err != nil {
		return nil, "", err
	}

	n := float64(len(rows))
	mean := 0.0
	std := 0.0
	if n > 0 {
		mean = sum / n
		variance := (sumSq / n) - mean*mean
		if variance < 0 {
			variance = 0
		}
		std = math.Sqrt(variance)
	}

	stats := &FactorPipelineStats{
		RowCount:      len(rows),
		Min:           round6(minVal),
		Max:           round6(maxVal),
		Mean:          round6(mean),
		Std:           round6(std),
		WarehousePath: s.warehousePath,
		PublishedPath: s.publishedPath,
	}
	return stats, s.publishedPath, nil
}

func (s *FactorPipelineService) LoadPreview(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, headers, err := s.loadCSV(s.publishedPath)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []map[string]interface{}{}, nil
	}
	preview := make([]map[string]interface{}, 0, minInt(limit, len(rows)))
	for i := 0; i < len(rows) && i < limit; i++ {
		out := map[string]interface{}{}
		for _, h := range headers {
			if h == "date" {
				sec := int64(rows[i][h])
				t := time.Unix(sec, 0).UTC()
				out[h] = t.Format("2006-01-02")
			} else {
				out[h] = rows[i][h]
			}
		}
		preview = append(preview, out)
	}
	return preview, nil
}

type FactorStats struct {
	RowCount int     `json:"row_count"`
	Mean     float64 `json:"mean"`
	Std      float64 `json:"std"`
}

func (s *FactorPipelineService) GetFactorStats(factorName string) (*FactorStats, error) {
	rows, headers, err := s.loadCSV(s.publishedPath)
	if err != nil {
		return nil, err
	}
	if !contains(headers, factorName) {
		return nil, fmt.Errorf("factor %q not found in published data", factorName)
	}

	sum := 0.0
	sumSq := 0.0
	count := 0

	for _, row := range rows {
		if val, ok := row[factorName]; ok {
			sum += val
			sumSq += val * val
			count++
		}
	}

	mean := 0.0
	std := 0.0
	if count > 0 {
		mean = sum / float64(count)
		variance := (sumSq / float64(count)) - mean*mean
		if variance < 0 {
			variance = 0
		}
		std = math.Sqrt(variance)
	}

	return &FactorStats{
		RowCount: count,
		Mean:     round6(mean),
		Std:      round6(std),
	}, nil
}

func (s *FactorPipelineService) CountPublishedRows() (int, error) {
	rows, _, err := s.loadCSV(s.publishedPath)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (s *FactorPipelineService) loadWarehouse() ([]map[string]float64, []string, error) {
	rows, headers, err := s.loadCSV(s.warehousePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, nil, err
	}
	return rows, headers, nil
}

func (s *FactorPipelineService) loadCSV(path string) ([]map[string]float64, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) < 1 {
		return []map[string]float64{}, []string{}, nil
	}
	headers := records[0]
	rows := make([]map[string]float64, 0, len(records)-1)
	for _, rec := range records[1:] {
		row := map[string]float64{}
		for i, h := range headers {
			if i >= len(rec) {
				continue
			}
			v, _ := strconv.ParseFloat(rec[i], 64)
			row[h] = v
		}
		rows = append(rows, row)
	}
	return rows, headers, nil
}

func (s *FactorPipelineService) writeCSV(path string, headers []string, rows []map[string]float64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		rec := make([]string, len(headers))
		for i, h := range headers {
			rec[i] = strconv.FormatFloat(row[h], 'f', 6, 64)
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	return nil
}

func (s *FactorPipelineService) bootstrapRows() []map[string]float64 {
	base := time.Date(2021, 4, 19, 0, 0, 0, 0, time.UTC)
	n := 500
	rows := make([]map[string]float64, 0, n)
	for i := 0; i < n; i++ {
		rows = append(rows, map[string]float64{
			"date": float64(base.AddDate(0, 0, i).Unix()),
		})
	}
	return rows
}

func contains(items []string, target string) bool {
	for _, x := range items {
		if x == target {
			return true
		}
	}
	return false
}

func computeFactorValue(idx int, name string) float64 {
	seed := float64(len(name)%7 + 3)
	x := float64(idx) / 21.0
	return round6(math.Sin(x*seed)*0.4 + math.Cos(x/seed)*0.2)
}

func round6(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
