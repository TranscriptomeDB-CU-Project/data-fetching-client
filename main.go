package main

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"arrayexpress-fetch/utils"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

func worker_fetch_accession(wg *sync.WaitGroup, queue chan string, metadata map[string]int) {
	defer wg.Done()

	for accession := range queue {
		fmt.Println("Start Accession: ", accession)

		var file_result dtos.SearchFileResult

		err := utils.FetchSDRFFileList(accession, &file_result)

		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, file := range file_result.Files {
			new_file_name, _ := strings.CutSuffix(file.Name, ".sdrf.txt")

			if _, err := os.Stat(fmt.Sprintf("sdrf/%s.sdrf.csv", new_file_name)); err == nil {
				fmt.Println("Skip: ", file.Name)
				continue
			}

			data_byte, err := utils.FetchAccessionSDRFFile(accession, file.Name)

			if err != nil {
				fmt.Println(err)
				continue
			}

			fp, err := os.OpenFile(fmt.Sprintf("sdrf/%s.sdrf.csv", new_file_name), os.O_RDWR|os.O_CREATE, 0755)

			if err != nil {
				fmt.Println("Read Failed: ", err)
				continue
			}

			fp.Write(data_byte)
			fp.Close()

			var target dtos.StudyInfo

			err = utils.FetchAccessionInfo(accession, &target)

			if err != nil {
				fmt.Println("Failed to fetch", err)
				continue
			}

			fmt.Println("Done Accession: ", file.Name)

			metadata[new_file_name] = target.Modified

			time.Sleep(200 * time.Millisecond)
		}

		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Shutdown Accession worker")
}

func worker_fetch_search(wg *sync.WaitGroup, queue chan int, result_queue chan string) {
	defer wg.Done()

	for accession := range queue {
		var body dtos.SearchResult

		fmt.Println("Start Page: ", accession)

		err := utils.FetchSearch(accession, &body)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Done Page: ", accession)

		for _, study := range body.Hits {
			result_queue <- study.Accession
		}

		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Shutdown Search worker")
}

func main() {
	var body dtos.SearchResult

	accession_metadata := make(map[string]int)

	err := utils.FetchSearch(1, &body)

	if err != nil {
		fmt.Println(err)
		return
	}

	totalPages := int(math.Ceil(float64(body.TotalHits) / float64(body.PageSize)))

	wg := sync.WaitGroup{}
	queue := make(chan int, constants.WORKER_NUMBER)

	wg_fetch_sdrf := sync.WaitGroup{}
	accession_queue := make(chan string, constants.WORKER_NUMBER*7)

	folder_name := "sdrf"

	if _, err := os.Stat(folder_name); err != nil {
		os.Mkdir(folder_name, 0755)
	}

	for i := 1; i <= constants.WORKER_NUMBER; i++ {
		wg.Add(1)
		go worker_fetch_search(&wg, queue, accession_queue)
	}

	for i := 1; i <= constants.WORKER_NUMBER; i++ {
		wg_fetch_sdrf.Add(1)
		go worker_fetch_accession(&wg_fetch_sdrf, accession_queue, accession_metadata)
	}

	for i := 1; i <= totalPages; i++ {
		queue <- i
	}

	close(queue)
	wg.Wait()

	fmt.Println("Done Search")

	close(accession_queue)
	wg_fetch_sdrf.Wait()

	fmt.Println("Done Accession")

	// Write map[string]int to file
	fp, err := os.OpenFile("metadata.csv", os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		fmt.Println("Read Failed: ", err)
		return
	}

	for k := range accession_queue {
		fmt.Printf("Remaining: %s\n", k)
	}
	defer fp.Close()

	fp.WriteString("accession,modified\n")
	for k, v := range accession_metadata {
		fp.WriteString(fmt.Sprintf("%s,%d\n", k, v))
	}

	fmt.Println("Total Accession: ", len(accession_metadata))
}
