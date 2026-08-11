package batch

import (
	"encoding/csv"
	"fmt"
	"io"
)

type Options struct {
	Workers       int
	MaxBatchBytes int
	OnProgress    func(Progress)
}

type Progress struct {
	RecordsRead      int64
	BytesRead        int64
	MalformedRecords int64
	SkippedClubs     int64
	MalformedError   string
}

func Process(r io.Reader, clubsOutput, playersOutput io.Writer) (Stats, error) {
	return ProcessWithOptions(r, clubsOutput, playersOutput, Options{Workers: 1})
}

func ProcessWithOptions(r io.Reader, clubsOutput, playersOutput io.Writer, options Options) (Stats, error) {
	workers := options.Workers
	if workers == 0 {
		workers = 1
	}

	if workers < 1 {
		return Stats{}, fmt.Errorf("Workers deve ser maior que zero.")
	}

	maxBatchBytes := options.MaxBatchBytes
	if maxBatchBytes == 0 {
		maxBatchBytes = defaultMaxBatchBytes
	}
	if maxBatchBytes < 1 {
		return Stats{}, fmt.Errorf("MaxBatchBytes deve ser maior que zero.")
	}

	clubsWriter := csv.NewWriter(clubsOutput)
	playersWriter := csv.NewWriter(playersOutput)

	if err := clubsWriter.Write(clubHeader); err != nil {
		return Stats{}, err
	}

	if err := playersWriter.Write(playerHeader); err != nil {
		return Stats{}, err
	}

	var (
		stats Stats
		err   error
	)

	stats, err = processParallel(r, clubsWriter, playersWriter, workers, maxBatchBytes, options.OnProgress)
	if err != nil {
		return stats, err
	}

	clubsWriter.Flush()
	if err := clubsWriter.Error(); err != nil {
		return stats, err
	}

	playersWriter.Flush()
	if err := playersWriter.Error(); err != nil {
		return stats, err
	}

	return stats, nil
}

func reportProgress(onProgress func(Progress), stats Stats, bytesRead int64, malformedError string) {
	if onProgress == nil {
		return
	}

	onProgress(Progress{
		RecordsRead:      stats.RecordsRead,
		BytesRead:        bytesRead,
		MalformedRecords: stats.MalformedRecords,
		SkippedClubs:     stats.FilteredChampionship,
		MalformedError:   malformedError,
	})
}

func writeResult(
	record processedRecord,
	clubsWriter *csv.Writer,
	playersWriter *csv.Writer,
	stats *Stats,
) error {
	stats.RecordsRead++

	if record.malformed {
		stats.MalformedRecords++
		return nil
	}
	if record.filtered {
		stats.FilteredChampionship++
		return nil
	}

	switch record.championship {
	case "SERIE A":
		stats.SerieAClubs++
	case "SERIE B":
		stats.SerieBClubs++
	}

	if err := clubsWriter.Write(record.clubRow); err != nil {
		return err
	}

	stats.ClubRowsWritten++

	for _, row := range record.playersRows {
		if err := playersWriter.Write(row); err != nil {
			return err
		}

		stats.PlayerRowsWritten++
	}
	return nil
}
