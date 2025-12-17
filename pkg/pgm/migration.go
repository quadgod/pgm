package pgm

type Migration struct {
	Name string `json:"name"`
	Up   string `json:"up"`
	Down string `json:"down"`
}
