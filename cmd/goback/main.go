package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/mimparat132/goback/pkg/discordwebhook"
)

// all backup files are in the form of <dateTimePrefix><dateTime><fileExtension>
type serverBackupConf struct {
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

type gobackConf struct {
	ServerBackupConfs []serverBackupConf `json:"server_backup_confs"`
	DiscordWebHookUrl string             `json:"discord_webhook_url,omitempty"`
}

func (sbc serverBackupConf) isValidBackupDir() (bool, error) {

	fileInfo, err := os.Stat(sbc.BaseDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		} else {
			return false, fmt.Errorf("could not stat base directory path: %v", err)
		}
	}

	// not only check if the path exists but also
	// making sure the path is a directory
	return fileInfo.IsDir(), nil

}

func (sbc serverBackupConf) validBackup() (bool, error) {

	files := []string{}

	validbackupDir, err := sbc.isValidBackupDir()
	if err != nil {
		return false, fmt.Errorf("could not check if backup is valid: %v", err)
	}

	if !validbackupDir {
		return false, fmt.Errorf("backup directory does not exist.")
	}

	fsys := os.DirFS(sbc.BaseDir)

	fs.WalkDir(fsys, ".", func(path string, dir fs.DirEntry, err error) error {

		if !dir.IsDir() {
			files = append(files, dir.Name())
		}

		return nil
	})

	for _, file := range files {

		noPrefix := strings.ReplaceAll(file, sbc.FileExtension, "")
		dateTimeString := strings.ReplaceAll(noPrefix, sbc.DateTimePrefix, "")

		// all backup file time stamps are RFC3339 compliant
		backupTime, err := time.Parse(time.RFC3339, dateTimeString)
		if err != nil {
			fmt.Printf("could not parse time string: %v", err)
		}

		// get current time and adjust to CST (-6 hours)
		currentTime := time.Now().UTC().Add(-6 * time.Hour)
		// threshold time is 24 hours before the current time
		thresholdTime := currentTime.Add(-24 * time.Hour)

		if backupTime.After(thresholdTime) {
			return true, nil
		}
	}

	return false, nil
}

func main() {

	data, err := os.ReadFile("/etc/goback/goback_conf.json")
	if err != nil {
		panic(err)
	}

	gobackConf := gobackConf{}

	err = json.Unmarshal(data, &gobackConf)
	if err != nil {
		fmt.Printf("[goback] - error: could not get goback config: %v", err)
		os.Exit(1)
	}

	backupStatusArr := []string{}

	for _, backupConf := range gobackConf.ServerBackupConfs {
		backupValid, err := backupConf.validBackup()
		if err != nil {
			errBackupStatus := fmt.Sprintf("FAIL: could not check backup for server: %s: %v\n", backupConf.ServerName, err)
			backupStatusArr = append(backupStatusArr, errBackupStatus)
			continue
		}

		if backupValid {
			validBackupStatus := fmt.Sprintf("SUCCESS: backup for server: %s, was taken within the last 24 hours!\n", backupConf.ServerName)
			backupStatusArr = append(backupStatusArr, validBackupStatus)
		} else {
			failedBackupStatus := fmt.Sprintf("FAIL: backup for server: %s, was not taken within the last 24 hours!\n", backupConf.ServerName)
			backupStatusArr = append(backupStatusArr, failedBackupStatus)
		}
	}

	username := "goback"

	runTime := time.Now()

	content := fmt.Sprintf("goback run: %s\n\n%s",
		runTime.Format("01-02-2006 15:04:05"),
		strings.Join(backupStatusArr, "\n"))

	message := discordwebhook.Message{
		Username: &username,
		Content:  &content,
	}

	err = discordwebhook.SendMessage(gobackConf.DiscordWebHookUrl, message)
	if err != nil {
		fmt.Printf("[goback] - error: could not send message to discord notification channel: %v\n", err)
		os.Exit(1)
	}

	return
}
