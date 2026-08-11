package batch

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"sync"
)

const (
	recordsPerWorkerBatch = 512
	defaultMaxBatchBytes  = 32 << 20
)

func processParallel(
	r io.Reader,
	clubsWriter *csv.Writer,
	playersWriter *csv.Writer,
	workers int,
	maxBatchBytes int,
	onProgress func(Progress),
) (Stats, error) {
	reader := bufio.NewReader(r)

	var stats Stats
	var bytesRead int64

	batchSize := workers * recordsPerWorkerBatch

	for {
		lines, eof, err := readBatch(reader, batchSize, maxBatchBytes)
		if err != nil {
			return stats, err
		}

		if len(lines) > 0 {
			records := processBatch(lines, workers)

			for index, record := range records {
				if err := writeResult(record, clubsWriter, playersWriter, &stats); err != nil {
					return stats, err
				}

				bytesRead += int64(len(lines[index]))
				reportProgress(onProgress, stats, bytesRead)
			}
		}
		if eof {
			break
		}
	}
	return stats, nil
}

func readBatch(reader *bufio.Reader, size, maxBytes int) (lines [][]byte, eof bool, err error) {
	lines = make([][]byte, 0, size)
	batchBytes := 0

	for len(lines) < size {
		line, readErr := reader.ReadBytes('\n')

		if len(line) > 0 {
			lines = append(lines, line)
			batchBytes += len(line)
		}

		if errors.Is(readErr, io.EOF) {
			return lines, true, nil
		}

		if readErr != nil {
			return lines, false, readErr
		}

		if batchBytes >= maxBytes {
			return lines, false, nil
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

		go func(start, end int) {
			defer wg.Done()

			for i := start; i < end; i++ {
				records[i] = processLine(lines[i])
			}
		}(start, end)
	}

	wg.Wait()

	return records
}
