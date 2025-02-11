package backup

import (
	"io/fs"
)

// all backup files are in the form of <dateTimePrefix><dateTime><fileExtension>
type ServerBackupConf struct {
	// The server backup target name
	ServerName string `json:"server_name"`
	// directory name that holds the backups for the given server
	BaseDir string `json:"base_dir"`
	// does this backup get taken daily
	Daily bool `json:"daily"`
	// the file extension of the backup file
	FileExtension string `json:"file_extension"`
	// the entire string before the date time starts
	DateTimePrefix string `json:"date_time_prefix"`
	// how is the time formatted in the file name
	// a time format compatible with time.Parse()
	TimeFormat string `json:"time_format"`
}

type GobackConf struct {
	ServerBackupConfs []ServerBackupConf `json:"server_backup_confs"`
	DiscordWebHookUrl string             `json:"discord_webhook_url,omitempty"`
}

type BackupFileRes struct {
	fileInfo fs.FileInfo
	valid bool
}
