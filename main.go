package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type RecordSlice struct {
	lifter string
	date   time.Time
	weight float64
}

type Filters struct {
	Gender    string `yaml:"Gender"`
	Lift      string `yaml:"Lift"`
	WeightCat string `yaml:"WeightCat"`
	Eq        string `yaml:"Eq"`
}

func writeRecordsToCSV(records []RecordSlice, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"Lifter", "Date", "Weight"})
	if err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	for _, record := range records {
		row := []string{
			record.lifter,
			record.date.Format("2006-01-02"),
			fmt.Sprintf("%.1f", record.weight),
		}
		err = writer.Write(row)
		if err != nil {
			return fmt.Errorf("error writing record: %v", err)
		}
	}
	return nil
}

func main() {
	var squatR, benchR, deadliftR, totalR float64
	var SquatRecordTimeline, BenchRecordTimeline, DeadliftRecordTimeline, TotalRecordTimeline []RecordSlice
	var filters Filters

	options, err := os.ReadFile("options.yaml")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	err = yaml.Unmarshal(options, &filters)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

	file, err := os.Open("./csv/data.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
	}
	if len(records) < 2 {
		return
	}
	data := records[1:]
	dateFormat := "2006-01-02"
	sort.Slice(data, func(i, j int) bool {
		date1, err1 := time.Parse(dateFormat, data[i][36])
		date2, err2 := time.Parse(dateFormat, data[j][36])
		if err1 != nil && err2 != nil {
			return false
		}
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		return date1.Before(date2)
	})

	for _, record := range data {

		if record[3] == filters.Eq && record[2] == "SBD" && record[1] == filters.Gender &&
			(record[9] == filters.WeightCat || (filters.WeightCat == "120+" && (record[9] == "Open" || record[9] == ""))) {
			date_, _ := time.Parse(dateFormat, record[36])
			squat, err := strconv.ParseFloat(record[14], 32)

			if err == nil && squat > squatR {
				SquatRecordTimeline = append(SquatRecordTimeline, RecordSlice{record[0], date_, squat})
				squatR = squat
			}

			bench, err := strconv.ParseFloat(fmt.Sprintf("%v", record[19]), 32)

			if err == nil && bench > benchR {
				BenchRecordTimeline = append(BenchRecordTimeline, RecordSlice{record[0], date_, bench})

				benchR = bench
			}

			deadlift, err := strconv.ParseFloat(record[24], 32)

			if err == nil && deadlift > deadliftR {
				DeadliftRecordTimeline = append(DeadliftRecordTimeline, RecordSlice{record[0], date_, deadlift})
				deadliftR = deadlift
			}

			total, err := strconv.ParseFloat(record[25], 32)
			if err == nil && total > totalR {
				TotalRecordTimeline = append(TotalRecordTimeline, RecordSlice{record[0], date_, total})
				totalR = total
			}

		}
	}
	switch filters.Lift {
	case "S":
		{
			writeRecordsToCSV(SquatRecordTimeline, "./csv/timeline.csv")
		}
	case "B":
		{
			writeRecordsToCSV(BenchRecordTimeline, "./csv/timeline.csv")
		}
	case "D":
		{
			writeRecordsToCSV(DeadliftRecordTimeline, "./csv/timeline.csv")
		}
	case "T":
		{
			writeRecordsToCSV(TotalRecordTimeline, "./csv/timeline.csv")
		}
	}
}
