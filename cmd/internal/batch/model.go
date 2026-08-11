package batch

type Club struct {
	ClubID       string   `json:"club_id"`
	Name         string   `json:"name"`
	Championship string   `json:"championship"`
	FoundingDate string   `json:"founding_date"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	Country      string   `json:"country"`
	Stadium      string   `json:"stadium"`
	President    string   `json:"president"`
	Nickname     *string  `json:"nickname"`
	Players      []Player `json:"players"`
	Colors       []string `json:"colors"`
}

type Player struct {
	PlayerID    string `json:"player_id"`
	Name        string `json:"name"`
	Age         *int   `json:"age"`
	Goals       *int   `json:"goals"`
	DebutDate   string `json:"debut_date"`
	Position    string `json:"position"`
	ShirtNumber *int   `json:"shirt_number"`
}
