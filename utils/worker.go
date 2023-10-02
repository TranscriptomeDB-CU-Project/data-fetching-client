package utils

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

func WorkerFetchAccession(wg *sync.WaitGroup, queue chan string, metadata *sync.Map, time_stamp *sync.Map) {
	defer wg.Done()

	for accession := range queue {
		time.Sleep(200 * time.Millisecond)

		fmt.Println("Start Accession: ", accession)

		var target dtos.StudyInfo

		err := FetchAccessionInfo(accession, &target)

		if err != nil {
			fmt.Println("Failed to fetch", err.Code, err.Message)
			continue
		}

		current_time := time.Unix(0, int64(target.Modified)*int64(time.Millisecond))

		if _metadata, ok := time_stamp.Load(accession); ok && current_time.Before(time.Unix(0, _metadata.(int64)*int64(time.Millisecond)+int64(7*24*time.Hour))) {
			metadata.Store(accession, "Up to date")
			fmt.Println("Uptodate: ", accession)
			continue
		}

		file_list, err := FetchSDRFFileList(accession)

		if err != nil {
			metadata.Store(accession, "Failed")
			fmt.Println(err)
			continue
		}

		if len(file_list) == 0 {
			metadata.Store(accession, "No SDRF")

			time_stamp.Store(accession, int64(target.Modified))
			continue
		}

		isFailed := false

		for _, file := range file_list {
			time.Sleep(200 * time.Millisecond)

			new_file_name, _ := strings.CutSuffix(file, ".sdrf.txt")

			data_byte, err := FetchAccessionSDRFFile(accession, file)

			if err != nil {
				fmt.Println(err)

				if err.Code != 404 && err.Code != 403 {
					isFailed = true
					metadata.Store(accession, "Failed")
				}

				continue
			}

			fp, _err := os.OpenFile(fmt.Sprintf("%ssdrf/%s.sdrf.csv", constants.FILE_BASE_PATH, new_file_name), os.O_RDWR|os.O_CREATE, 0755)

			if _err != nil {
				fmt.Println("Read Failed: ", _err)
				isFailed = true

				metadata.Store(accession, "Failed")
				continue
			}

			fp.Write(data_byte)
			fp.Close()

			fmt.Println("Done Accession: ", file)

			metadata.Store(accession, "Success")
		}

		if !isFailed {
			time_stamp.Store(accession, int64(target.Modified))
		}
	}

	fmt.Println("Shutdown Accession worker")
}

func WorkerFetchSearch(species string, wg *sync.WaitGroup, queue chan int, result_queue chan string) {
	defer wg.Done()

	for accession := range queue {
		time.Sleep(200 * time.Millisecond)

		var body dtos.SearchResult

		fmt.Println("Start Page: ", accession)

		err := FetchSearch(species, accession, &body)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Done Page: ", accession)

		for _, study := range body.Hits {
			result_queue <- study.Accession
		}
	}

	fmt.Println("Shutdown Search worker")
}
