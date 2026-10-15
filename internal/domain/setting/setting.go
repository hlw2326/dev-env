package setting

type Settings struct {
	Theme             string `json:"theme"`
	Root              string `json:"root"`
	Port              int    `json:"port"`
	Hosts             bool   `json:"hosts"`
	DeveloperToolsCLI string `json:"developerToolsCLI"`
}
