package csv

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func AppendToFile(filePath string, lineData []string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatal(err)
	}

	writer := csv.NewWriter(file)
	writer.Write(lineData)
	writer.Flush()
	err = writer.Error()
	if err != nil {
		fmt.Println("Error while writing to the file ::", err)
		return
	}
}
