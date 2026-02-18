package types

type ChampionPost struct {
	Content Content `json:"content"`
	Title   Title   `json:"title"`
}

type Content struct {
	Rendered string `json:"rendered"`
}

type Title struct {
	Rendered string `json:"rendered"`
}

type Champion struct {
	Name           string
	Class          string
	BasicAbilities []string
	Strengths      []string
	Weaknesses     []string
	Abilities      []Ability
}

type Ability struct {
	Name     string
	Effects  []string
	DevNotes []string
}
