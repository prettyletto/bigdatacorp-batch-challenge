package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCreatesRequestedRecordsAndParentDirectory(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "generated", "clubs.jsonl")

	if err := generate(outputPath, 3, 2); err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("opening generated file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var clubs []Club
	for {
		var club Club
		err := decoder.Decode(&club)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("decoding generated club: %v", err)
		}
		clubs = append(clubs, club)
	}

	if len(clubs) != 3 {
		t.Fatalf("generated clubs = %d, want 3", len(clubs))
	}
	for index, wantChampionship := range []string{"SERIE A", "SERIE B", "SERIE C"} {
		t.Run(wantChampionship, func(t *testing.T) {
			club := clubs[index]
			wantID := []string{"CLUB-00000000", "CLUB-00000001", "CLUB-00000002"}[index]
			if club.ClubID != wantID {
				t.Fatalf("club ID = %q, want %q", club.ClubID, wantID)
			}
			if club.Championship != wantChampionship {
				t.Fatalf("championship = %q, want %q", club.Championship, wantChampionship)
			}
			if len(club.Players) != 2 {
				t.Fatalf("players = %d, want 2", len(club.Players))
			}
		})
	}
}

func TestMakeClubGeneratesRequestedPlayerCount(t *testing.T) {
	club := makeClub(4, 0)

	if club.ClubID != "CLUB-00000004" {
		t.Fatalf("club ID = %q, want %q", club.ClubID, "CLUB-00000004")
	}
	if club.Championship != "SERIE B" {
		t.Fatalf("championship = %q, want %q", club.Championship, "SERIE B")
	}
	if len(club.Players) != 0 {
		t.Fatalf("players = %d, want 0", len(club.Players))
	}
}
