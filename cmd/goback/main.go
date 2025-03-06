package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mimparat132/goback/pkg/backup"
	"github.com/mimparat132/goback/pkg/discordwebhook"
)

func returnStateString(valid bool) string {
	if valid {
		return "VALID"
	} else {
		return "FAIL"
	}
}

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
		backupCheckRes, err := backupConf.ValidBackup()
		if err != nil {
			errBackupStatus := fmt.Sprintf("FAIL: could not check backup for server: %s: %v\n", backupConf.ServerName, err)
			backupStatusArr = append(backupStatusArr, errBackupStatus)
			continue
		}

		if !backupCheckRes.PrimaryBackupValid {
			backupStatus := fmt.Sprintf("\n%s:\n\tprimary_backup_state: %s\n\tsecondary_backup_state: %s",
				backupConf.ServerName,
				returnStateString(backupCheckRes.PrimaryBackupValid),
				returnStateString(backupCheckRes.SecondaryBackupValid),
			)
			backupStatusArr = append(backupStatusArr, backupStatus)
			continue
		}

		backupStatus := fmt.Sprintf("\n%s:\n\tprimary_backup_state: %s\n\tprimary_backup_filename: %s\n\tprimary_backup_time: %s\n\tprimary_backup_size: %d KB\n\tsecondary_backup_state: %s\n\tsecondary_backup_filename: %s\n\tsecondary_backup_time: %s\n\tsecondary_backup_size: %d KB",
			backupConf.ServerName,
			returnStateString(backupCheckRes.PrimaryBackupValid),
			backupCheckRes.PrimaryBackupFileInfo.Name(),
			backupCheckRes.PrimaryBackupTimeString,
			backupCheckRes.PrimaryBackupFileInfo.Size()/1024,
			returnStateString(backupCheckRes.SecondaryBackupValid),
			backupCheckRes.SecondaryBackupFileInfo.Name(),
			backupCheckRes.SecondaryBackupTimeString,
			backupCheckRes.SecondaryBackupFileInfo.Size()/1024,
		)

		backupStatusArr = append(backupStatusArr, backupStatus)
	}

	username := "goback"

	runTime := time.Now()

	content := fmt.Sprintf("goback run: %s\n", runTime.Format("01-02-2006 15:04:05"))

	messageSlice := discordwebhook.Message{
		Username: &username,
		Content:  &content,
	}

	err = discordwebhook.SendMessage(gobackConf.DiscordWebHookUrl, messageSlice)
	if err != nil {
		fmt.Printf("[goback] - error: could not send message to discord notification channel: %v\n", err)
		os.Exit(1)
	}

	// batching the backup status report being sent to discord
	for _, backupStatus := range backupStatusArr {

		// Add a newline to make the output more readable by separating
		// each backup status block
		backupStatus = backupStatus + "\n\n"

		messageSlice := discordwebhook.Message{
			Username: &username,
			Content:  &backupStatus,
		}

		err = discordwebhook.SendMessage(gobackConf.DiscordWebHookUrl, messageSlice)
		if err != nil {
			fmt.Printf("[goback] - error: could not send message to discord notification channel: %v\n", err)
			os.Exit(1)
		}
	}

	return
}
