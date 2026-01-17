package config

// GetDefaultConfkeeperConfig 返回默认的ConfkeeperConfig
func GetDefaultConfkeeperConfig() ConfkeeperConfig {
	return ConfkeeperConfig{
		ConfigType: []string{
			"text",
			"json",
			"xml",
			"yaml",
			"html",
			"properties",
			"toml",
			"ini",
		},
		ActionType: []string{
			"r",
			"w",
			"rw",
		},
	}
}
