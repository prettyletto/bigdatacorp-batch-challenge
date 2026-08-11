package batch

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
)

func processSequential(
	r io.Reader,
	clubsWriter *csv.Writer,
	playersWriter *csv.Writer,
	onProgress func(Progress),
) (Stats, error) {
	reader := bufio.NewReader(r)

	var stats Stats
	var bytesRead int64

	for {
		line, readErr := reader.ReadBytes('\n')

		if len(line) > 0 {
			record := processLine(line)

			if err := writeResult(
				record,
				clubsWriter,
				playersWriter,
				&stats,
			); err != nil {
				return stats, err
			}

			bytesRead += int64(len(line))
			reportProgress(onProgress, stats, bytesRead, record.malformedError)
		}

		if errors.Is(readErr, io.EOF) {
			break
		}

		if readErr != nil {
			return stats, readErr
		}
	}

	return stats, nil
}
