package config

// Config represents the application configuration loaded from TOML file.
type Config struct {
	UseDirectory string `toml:"use_directory"`
	Port         int    `toml:"port"`
	BaseUrl      string `toml:"base_url"`
	Users        []User `toml:"users"`
}

// User represents a user with authorization code for uploading archives.
type User struct {
	Name string `toml:"name"`
	Auth string `toml:"auth"`
}
