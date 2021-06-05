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
	storagePath := GetStoragePath()
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.Mkdir(storagePath, 0700)
	}
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

func GetStoragePathForFile(filename string) string {
	var sb strings.Builder
	sb.WriteString(GetStoragePath())
	sb.WriteString("/")
	sb.WriteString(filename)
	return sb.String()
}

func GetFilePathForFilename(filename string) string {
	var sb strings.Builder
	sb.WriteString(GetStoragePath())
	sb.WriteString("/")
	sb.WriteString(filename)
	return sb.String()
}

func GetFilePath(name string) string {
	var sb strings.Builder
	sb.WriteString(GetStoragePath())
	sb.WriteString("/")
	sb.WriteString(".timestamps-")
	sb.WriteString(name)
	return sb.String()
}
