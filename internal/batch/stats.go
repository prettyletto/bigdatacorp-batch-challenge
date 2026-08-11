package batch

type Stats struct {
	RecordsRead          int64
	MalformedRecords     int64
	FilteredChampionship int64
	SerieAClubs          int64
	SerieBClubs          int64
	ClubRowsWritten      int64
	PlayerRowsWritten    int64
}
