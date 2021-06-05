package timezone

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fredrik01/ts/src/storage"
)

const (
	timezoneFilename = "tz"
)

func InTimezone(timestamp time.Time) time.Time {
	if timezoneFileExists() {
		location, _ := time.LoadLocation(readLocation())
		return timestamp.In(location)
	}
	return timestamp.Local()
}

func SetTimezone(location string) {
	// Remove old file
	DeleteTimezoneFileIfExists()

	// Create location file
	file := getTimezoneFilePath()
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(location)); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
}

func DeleteTimezoneFileIfExists() {
	file := getTimezoneFilePath()

	if _, err := os.Stat(file); err == nil {
		e := os.Remove(file)
		if e != nil {
			log.Fatal(e)
		}
	}
}

func GetTimestampFiles() []string {
	files, err := ioutil.ReadDir(storage.GetStoragePath())
	if err != nil {
		log.Fatal(err)
	}

	var timestampFiles []string
	for _, file := range files {
		filename := file.Name()
		if strings.Contains(filename, ".timestamps-") {
			timestampFiles = append(timestampFiles, filename)
		}
	}
	return timestampFiles
}

func getTimezoneFilePath() string {
	return storage.GetStoragePathForFile(timezoneFilename)
}

func readLocation() string {
	file := storage.GetStoragePathForFile(timezoneFilename)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func timezoneFileExists() bool {
	file := getTimezoneFilePath()
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}
