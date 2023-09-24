package utils

import (
	"arrayexpress-fetch/dtos"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

func WorkerFetchAccession(wg *sync.WaitGroup, queue chan string, metadata map[string][]dtos.ResultMetadata, time_stamp map[string]int64) {
	defer wg.Done()

	for accession := range queue {
		time.Sleep(200 * time.Millisecond)

		fmt.Println("Start Accession: ", accession)

		var target dtos.StudyInfo

		err := FetchAccessionInfo(accession, &target)

		if err != nil {
			fmt.Println("Failed to fetch", err)
			continue
		}

		current_time := time.Unix(0, int64(target.Modified)*int64(time.Millisecond))

		if _metadata, ok := time_stamp[accession]; ok && current_time.Before(time.Unix(0, int64(_metadata)*int64(time.Millisecond)+int64(7*24*time.Hour))) {
			metadata["Uptodate"] = append(metadata["Uptodate"], dtos.ResultMetadata{
				Name: accession,
			})
			fmt.Println("Uptodate: ", accession)
			continue
		}

		time_stamp[accession] = int64(target.Modified)

		file_list, err := FetchSDRFFileList(accession)

		if err != nil {
			fmt.Println(err)
			continue
		}

		if len(file_list) == 0 {
			metadata["NoSDRF"] = append(metadata["NoSDRF"], dtos.ResultMetadata{
				Name: accession,
			})
			continue
		}

		for _, file := range file_list {
			time.Sleep(200 * time.Millisecond)

			new_file_name, _ := strings.CutSuffix(file, ".sdrf.txt")

			data_byte, err := FetchAccessionSDRFFile(accession, file)

			if err != nil {
				fmt.Println(err)
				metadata["Failed"] = append(metadata["Failed"], dtos.ResultMetadata{
					Name: accession,
				})
				continue
			}

			fp, err := os.OpenFile(fmt.Sprintf("sdrf/%s.sdrf.csv", new_file_name), os.O_RDWR|os.O_CREATE, 0755)

			if err != nil {
				fmt.Println("Read Failed: ", err)
				continue
			}

			fp.Write(data_byte)
			fp.Close()

			fmt.Println("Done Accession: ", file)

			metadata["Success"] = append(metadata["Success"], dtos.ResultMetadata{
				Name: accession,
			})
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
