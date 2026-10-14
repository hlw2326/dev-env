package settings

type AppSettings struct {
	Language           string `json:"language"`
	Theme              string `json:"theme"`
	RuntimeRootPath    string `json:"runtimeRootPath"`
	DefaultProjectPath string `json:"defaultProjectPath"`
	AutoStartServices  bool   `json:"autoStartServices"`
}

func DefaultSettings() AppSettings {
	return AppSettings{
		Language:           "zh-CN",
		Theme:              "system",
		RuntimeRootPath:    "",
		DefaultProjectPath: "",
		AutoStartServices:  false,
	}
}
