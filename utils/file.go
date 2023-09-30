package utils

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func WriteMetadata(metadata map[string][]dtos.ResultMetadata) {
	fp, err := os.OpenFile(fmt.Sprintf("%smetadata.txt", constants.FILE_BASE_PATH), os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		fmt.Println("Read Failed: ", err)
		return
	}

	defer fp.Close()

	for key, value := range metadata {
		fp.WriteString(fmt.Sprintf("%s: ", key))

		for _, accession := range value {
			fp.WriteString(fmt.Sprintf("%s,", accession.Name))
		}

		fp.WriteString("\n")
	}
}

func WriteTimestamp(timestamp map[string]int64) {
	fp_time, err := os.OpenFile(fmt.Sprintf("%stimestamp.txt", constants.FILE_BASE_PATH), os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		fmt.Println("Read Failed: ", err)
		return
	}

	defer fp_time.Close()

	fp_time.WriteString("accession,timestamp\n")

	for key, value := range timestamp {
		fp_time.WriteString(fmt.Sprintf("%s,%d\n", key, value))
	}
}

func ReadTimestamp() map[string]int64 {
	timestamp := make(map[string]int64)

	fp, err := os.Open(fmt.Sprintf("%stimestamp.txt", constants.FILE_BASE_PATH))

	if err != nil {
		fmt.Println("Read Failed: ", err)
		return timestamp
	}

	defer fp.Close()

	// Read CSV of fp
	fileReader := csv.NewReader(fp)

	records, err := fileReader.ReadAll()
	if err != nil {
		fmt.Println("Read Failed: ", err)
		return timestamp
	}

	for _, record := range records {
		if len(record) == 2 {
			t, err := strconv.ParseInt(record[1], 10, 64)

			if err != nil {
				fmt.Println("Read Failed: ", err)
				continue
			}

			timestamp[record[0]] = t
		}
	}

	return timestamp
}
