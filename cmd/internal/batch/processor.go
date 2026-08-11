package batch

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
)

func Process(r io.Reader, clubsOutput, playersOutput io.Writer) (Stats, error) {
	reader := bufio.NewReader(r)

	clubsWriter := csv.NewWriter(clubsOutput)
	playersWriter := csv.NewWriter(playersOutput)

	if err := clubsWriter.Write(clubHeader); err != nil {
		return Stats{}, err
	}

	if err := playersWriter.Write(playerHeader); err != nil {
		return Stats{}, err
	}

	var stats Stats

	for {
		line, err := reader.ReadBytes('\n')

		if len(line) > 0 {
			stats.RecordsRead++

			var club Club

			if err := json.Unmarshal(line, &club); err != nil {
				stats.MalformedRecords++
			} else if err := processClub(
				club,
				clubsWriter,
				playersWriter,
				&stats,
			); err != nil {
				return stats, err
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return stats, err
		}
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

func processClub(club Club, clubsWriter, playersWriter *csv.Writer, stats *Stats) error {
	championship, ok := filterChampionship(club.Championship)
	if !ok {
		stats.FilteredChampionship++
		return nil
	}

	switch championship {
	case "SERIE A":
		stats.SerieAClubs++
	case "SERIE B":
		stats.SerieBClubs++
	}

	if err := clubsWriter.Write(clubRow(club)); err != nil {
		return err
	}
	stats.ClubRowsWritten++

	for _, player := range club.Players {
		if err := playersWriter.Write(playerRow(club.ClubID, player)); err != nil {
			return err
		}
		stats.PlayerRowsWritten++
	}

	return nil
}
