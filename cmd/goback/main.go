package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mimparat132/goback/pkg/backup"
	"github.com/mimparat132/goback/pkg/discordwebhook"
)

func main() {

	data, err := os.ReadFile("/etc/goback/goback_conf.json")
	if err != nil {
		panic(err)
	}

	gobackConf := backup.GobackConf{}

	err = json.Unmarshal(data, &gobackConf)
	if err != nil {
		fmt.Printf("[goback] - error: could not get goback config: %v", err)
		os.Exit(1)
	}

	backupStatusArr := []string{}

	for _, backupConf := range gobackConf.ServerBackupConfs {
		backupValid, fileInfo,unixTimeStampWithTimeZone, err := backupConf.ValidBackup()
		if err != nil {
			errBackupStatus := fmt.Sprintf("FAIL: could not check backup for server: %s: %v\n", backupConf.ServerName, err)
			backupStatusArr = append(backupStatusArr, errBackupStatus)
			continue
		}

		if backupValid {
			validBackupStatus := fmt.Sprintf("\n%s:\n\tstate: VALID\n\tfile_name: %s\n\tbackup_time: %s\n\tfile_size: %d KB",
				backupConf.ServerName,
				fileInfo.Name(),
				unixTimeStampWithTimeZone,
				fileInfo.Size()/1024,
			)

			backupStatusArr = append(backupStatusArr, validBackupStatus)

		} else {
			failedBackupStatus := fmt.Sprintf("%s:\n\tstate: FAIL",
				backupConf.ServerName)
			backupStatusArr = append(backupStatusArr, failedBackupStatus)
		}
	}

	username := "goback"

	runTime := time.Now()

	content := fmt.Sprintf("goback run: %s\n%s",
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
