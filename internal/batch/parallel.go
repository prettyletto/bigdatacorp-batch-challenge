package batch

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"sync"
)

const recordsPerWorkerBatch = 512

func processParallel(
	r io.Reader,
	clubsWriter *csv.Writer,
	playersWriter *csv.Writer,
	workers int,
) (Stats, error) {
	reader := bufio.NewReader(r)

	var stats Stats

	batchSize := workers * recordsPerWorkerBatch

	for {
		lines, eof, err := readBatch(reader, batchSize)
		if err != nil {
			return stats, err
		}

		if len(lines) > 0 {
			records := processBatch(lines, workers)

			for _, record := range records {
				if err := writeResult(record, clubsWriter, playersWriter, &stats); err != nil {
					return stats, err
				}
			}
		}
		if eof {
			break
		}
	}
	return stats, nil
}

func readBatch(reader *bufio.Reader, size int) (lines [][]byte, eof bool, err error) {
	lines = make([][]byte, 0, size)

	for len(lines) < size {
		line, readErr := reader.ReadBytes('\n')

		if len(line) > 0 {
			lines = append(lines, line)
		}

		if errors.Is(readErr, io.EOF) {
			return lines, true, nil
		}

		if readErr != nil {
			return lines, false, readErr
		}
	}

	return lines, false, nil
}

func processBatch(lines [][]byte, workers int) []processedRecord {
	records := make([]processedRecord, len(lines))

	if len(lines) == 0 {
		return records
	}

	if workers > len(lines) {
		workers = len(lines)
	}

	chunkSize := (len(lines) + workers - 1) / workers

	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		start := worker * chunkSize

		if start >= len(lines) {
			break
		}

		end := start + chunkSize

		if end > len(lines) {
			end = len(lines)
		}

		wg.Add(1)

		go func(start, ent int) {
			defer wg.Done()

			for i := start; i < end; i++ {
				records[i] = processLine(lines[i])
			}
		}(start, end)
	}

	wg.Wait()

	return records
}
