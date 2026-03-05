package config

type Config struct {
	Editor string
	Backup BackupConfig
}

type BackupConfig struct {
	Provider string
	Git      GitConfig
}

type GitConfig struct {
	Repo   string
	Remote string
	Branch string
}
