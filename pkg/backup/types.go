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
	// Secondary Backup server target
	SecondaryBaseDir string `json:"secondary_base_dir"`
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

type BackupCheckResponse struct {
	// Does the primary backup file exist
	PrimaryBackupValid bool
	// Primary Backup file info ( name, size, etc...)
	PrimaryBackupFileInfo fs.FileInfo
	// DateTime string of the valid primary backup file
	PrimaryBackupTimeString string
	// Does the secondary backup file exist
	SecondaryBackupValid bool
	// Secondary Backup file info ( name, size, etc...)
	SecondaryBackupFileInfo fs.FileInfo
	// DateTime string of the valid secondary backup file
	SecondaryBackupTimeString string
}

type GobackConf struct {
	ServerBackupConfs []ServerBackupConf `json:"server_backup_confs"`
	DiscordWebHookUrl string             `json:"discord_webhook_url,omitempty"`
}

type BackupFileRes struct {
	fileInfo fs.FileInfo
	valid    bool
}
