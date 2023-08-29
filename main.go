package main

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"arrayexpress-fetch/utils"
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

func worker_fetch_accession(wg *sync.WaitGroup, queue chan string) {
	defer wg.Done()

	for job := range queue {
		data_byte, err := utils.FetchAccession(job)

		if err != nil {
			fmt.Println(err)
			return
		}

		if _, err := os.Stat(fmt.Sprintf("sdrf/%s.sdrf.csv", job)); err == nil {
			fmt.Println("Skip: ", job)
			continue
		}

		fp, err := os.OpenFile(fmt.Sprintf("sdrf/%s.sdrf.csv", job), os.O_RDWR|os.O_CREATE, 0755)

		if err != nil {
			fmt.Println("Read Failed: ", err)
			return
		}

		fp.Write(data_byte)
		fp.Close()

		fmt.Println("Done Accession: ", job)

		time.Sleep(50 * time.Millisecond)
	}
}

func worker_fetch_search(wg *sync.WaitGroup, queue chan int, result_queue chan string) {
	defer wg.Done()

	for job := range queue {
		var body dtos.SearchResult

		err := utils.FetchSearch(job, &body)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Done Page: ", job)

		for _, study := range body.Hits {
			result_queue <- study.Accession
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	var body dtos.SearchResult

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

	for i := 1; i <= constants.WORKER_NUMBER*5; i++ {
		wg_fetch_sdrf.Add(1)
		go worker_fetch_accession(&wg_fetch_sdrf, accession_queue)
	}

	for i := 1; i <= totalPages; i++ {
		wg_fetch_sdrf.Add(1)
		queue <- i
	}

	wg.Wait()
	wg_fetch_sdrf.Wait()

	close(queue)
	close(accession_queue)
}
