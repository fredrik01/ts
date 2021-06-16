package storage

import (
	"log"
	"os"
	"strings"
)

const (
	storageFolder = ".config/ts"
)

func SetupStorage() {
	saveFilePath := GetStoragePath()
	if !FileExist(saveFilePath) {
		os.Mkdir(saveFilePath, 0700)
	}
}

func FileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else {
		return false
	}
}

func GetSaveFilePath() string {
	var sb strings.Builder
	sb.WriteString(GetStoragePath())
	sb.WriteString("/")
	sb.WriteString("timestamps.csv")
	return sb.String()
}

func GetStoragePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	var storage strings.Builder
	storage.WriteString(home)
	storage.WriteString("/")
	storage.WriteString(storageFolder)
	return storage.String()
}

func Delete(path string) {
	e := os.Remove(path)
	if e != nil {
		log.Fatal(e)
	}
}

func GetStoragePathForFile(filename string) string {
	var sb strings.Builder
	sb.WriteString(GetStoragePath())
	sb.WriteString("/")
	sb.WriteString(filename)
	return sb.String()
}
