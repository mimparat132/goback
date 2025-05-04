package backup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"
)

func (sbc ServerBackupConf) IsValidBackupDir(dirPath string) (bool, error) {

	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		} else {
			return false, fmt.Errorf("could not stat path '%s': %v", dirPath, err)
		}
	}

	// not only check if the path exists but also
	// making sure the path is a directory
	return fileInfo.IsDir(), nil

}

func (sbc ServerBackupConf) ValidBackup() (BackupCheckResponse, error) {

	backupCheckRes := BackupCheckResponse{}

	fileInfoArr := []fs.FileInfo{}

	validBaseDir, err := sbc.IsValidBackupDir(sbc.BaseDir)
	if err != nil {
		return backupCheckRes, fmt.Errorf("could not check if backup is valid: %v", err)
	}

	if !validBaseDir {
		return backupCheckRes, fmt.Errorf("primary backup base directory does not exist.")
	}

	validSecondaryBaseDir, err := sbc.IsValidBackupDir(sbc.SecondaryBaseDir)
	if err != nil {
		return backupCheckRes, fmt.Errorf("could not check if backup is valid: %v", err)
	}

	if !validSecondaryBaseDir {
		return backupCheckRes, fmt.Errorf("secondary backup base directory does not exist.")
	}

	// search the primary backup target
	fsys := os.DirFS(sbc.BaseDir)

	err = fs.WalkDir(fsys, ".", func(path string, dir fs.DirEntry, err error) error {

		if !dir.IsDir() {
			fileInfo, err := dir.Info()

			if err != nil {
				return err
			}

			fileInfoArr = append(fileInfoArr, fileInfo)
		}

		return nil
	})

	if err != nil {
		return backupCheckRes, fmt.Errorf("could not walk directory: %s: %v", sbc.BaseDir, err)
	}

	for _, fileInfo := range fileInfoArr {

		noPrefix := strings.ReplaceAll(fileInfo.Name(), sbc.FileExtension, "")
		dateTimeString := strings.ReplaceAll(noPrefix, sbc.DateTimePrefix, "")

		i, err := strconv.ParseInt(dateTimeString, 10, 64)
		if err != nil {
			return backupCheckRes, fmt.Errorf("could not parse unix time int: %v", err)
		}
		unixTimeStringWithTimeZone := time.Unix(i, 0).String()
		unixTimeString := time.Unix(i, 0).UTC().String()

		// 2025-02-08 19:15:36 -0600 CST is format of unixTimeString

		// 2006-01-02 15:04:05 -0700 MST is the time package format for the time
		// returned by time.Unix(i, 0).UTC().String()
		// we have to call .UTC() since time.Unix() always returns the time
		// with the timezone of the machine calling the time.Unix() method
		// and we really want the time in -0000 UTC

		backupTime, err := time.Parse("2006-01-02 15:04:05 -0700 MST", unixTimeString)
		if err != nil {
			fmt.Printf("could not parse time string: %v", err)
		}

		currentTime := time.Now().UTC()
		// threshold time is 24 hours before the current time
		thresholdTime := currentTime.Add(-24 * time.Hour)

		if backupTime.After(thresholdTime) {
			backupCheckRes.PrimaryBackupValid = true
			backupCheckRes.PrimaryBackupFileInfo = fileInfo
			backupCheckRes.PrimaryBackupTimeString = unixTimeStringWithTimeZone
			break
		}
	}

	// If the primary backup isn't in place then the secondary backup wont
	// be present. We can return false for everything here
	if !backupCheckRes.PrimaryBackupValid {
		return backupCheckRes, nil
	}

	// search the primary backup target
	fsysSecondary := os.DirFS(sbc.SecondaryBaseDir)

	fs.WalkDir(fsysSecondary, ".", func(path string, dir fs.DirEntry, err error) error {

		if !dir.IsDir() {
			fileInfo, err := dir.Info()

			if err != nil {
				return err
			}

			fileInfoArr = append(fileInfoArr, fileInfo)
		}

		return nil
	})

	for _, fileInfo := range fileInfoArr {
		// If true then the backup file exists on the primary and secondary
		// backup targets
		if fileInfo.Name() == backupCheckRes.PrimaryBackupFileInfo.Name() {
			backupCheckRes.SecondaryBackupValid = true
			backupCheckRes.SecondaryBackupFileInfo = fileInfo
			backupCheckRes.SecondaryBackupTimeString = fileInfo.ModTime().String()
			break
		}
	}

	return backupCheckRes, nil
}

func (sbc ServerBackupConf) PrintFileNamesInEpoch() error {

	files := []string{}

	validbackupDir, err := sbc.IsValidBackupDir()
	if err != nil {
		return fmt.Errorf("could not check if backup is valid: %v", err)
	}

	if !validbackupDir {
		return fmt.Errorf("backup directory does not exist.")
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

		newBackupFileName := fmt.Sprintf("%s%d%s",
			sbc.DateTimePrefix,
			backupTime.Unix(),
			sbc.FileExtension)
		fmt.Println(newBackupFileName)

	}

	return nil
}

func (sbc ServerBackupConf) PrintFileNamesInRFC3339() error {

	files := []string{}

	validbackupDir, err := sbc.IsValidBackupDir()
	if err != nil {
		return fmt.Errorf("could not check if backup is valid: %v", err)
	}

	if !validbackupDir {
		return fmt.Errorf("backup directory does not exist.")
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
		backupTime, err := time.Parse(time.UnixDate, dateTimeString)
		if err != nil {
			fmt.Printf("could not parse time string: %v", err)
		}

		newBackupFileName := fmt.Sprintf("%s%s%s",
			sbc.DateTimePrefix,
			backupTime.Format(time.RFC3339),
			sbc.FileExtension)
		fmt.Println(newBackupFileName)

	}

	return nil
}
