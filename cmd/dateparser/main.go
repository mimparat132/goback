package main

import (
	"encoding/json"
	"fmt"
	"os"
	_ "time"

	"github.com/mimparat132/goback/pkg/backup"
)

func updateFilename() {

	return
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

	for _, backupConf := range gobackConf.ServerBackupConfs {
		fmt.Println(backupConf.ServerName)
		err := backupConf.PrintFileNamesInRFC3339()
		if err != nil {
			fmt.Println("could not print backup file in RFC3339 time:",backupConf.ServerName)
		}
		fmt.Println("end backup files:",backupConf.ServerName)
	}

	return
}
