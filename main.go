package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

const api_url = "https://www.ebi.ac.uk/biostudies/api/v1"
const file_base_url = "https://www.ebi.ac.uk/biostudies/files"

type Study struct {
	Accession    string `json:"accession"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	Links        int    `json:"links"`
	Files        int    `json:"files"`
	Release_date string `json:"release_date"`
	Views        int    `json:"views"`
	IsPublic     bool   `json:"isPublic"`
}

type SearchResult struct {
	TotalHits int     `json:"totalHits"`
	Page      int     `json:"page"`
	PageSize  int     `json:"pageSize"`
	Hits      []Study `json:"hits"`
}

func fetch_search(page int, target interface{}) error {
	fetch_url := fmt.Sprintf("%s/arrayexpress/search?page=%d&pageSize=100", api_url, page)

	println(fetch_url)

	res, err := http.Get(fetch_url)

	if err != nil {
		fmt.Println("Fetch Failed: ", err)

		return err
	}

	if res.StatusCode != 200 {
		fmt.Println("Fetch Failed: ", res.Status)
		return fmt.Errorf("Fetch Failed: %s", res.Status)
	}

	// Read body
	err = json.NewDecoder(res.Body).Decode(target)

	return err
}

func worker_fetch_accession(wg *sync.WaitGroup, queue chan string) {
	defer wg.Done()

	for job := range queue {
		if _, err := os.Stat(fmt.Sprintf("sdrf/%s.sdrf.txt", job)); err == nil {
			fmt.Println("Skip: ", job)
			continue
		}

		fp, err := os.OpenFile(fmt.Sprintf("sdrf/%s.sdrf.txt", job), os.O_RDWR|os.O_CREATE, 0755)

		if err != nil {
			fmt.Println(err)
			return
		}

		fetch_url := fmt.Sprintf("%s/%s/%s.sdrf.txt", file_base_url, job, job)

		res, err := http.Get(fetch_url)

		if res.StatusCode != 200 {
			fmt.Println("Fetch Failed: ", res.Status)
			return
		}

		if err != nil {
			fmt.Println("Fetch Failed: ", err)
			return
		}

		data_byte, err := io.ReadAll(res.Body)

		if err != nil {
			fmt.Println("Read Failed: ", err)
			return
		}

		fp.Write(data_byte)
		fp.Close()

		time.Sleep(100 * time.Millisecond)

		fmt.Println("Fetch Success: ", job)
	}
}

func worker_fetch_search(wg *sync.WaitGroup, queue chan int, result_queue chan string) {
	defer wg.Done()

	for job := range queue {
		var body SearchResult
		err := fetch_search(job, &body)

		if err != nil {
			fmt.Println(err)
			return
		}

		for _, study := range body.Hits {
			result_queue <- study.Accession
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	var body SearchResult
	const max_worker = 10
	err := fetch_search(1, &body)

	if err != nil {
		fmt.Println(err)
		return
	}

	totalPages := int(math.Ceil(float64(body.TotalHits) / float64(body.PageSize)))

	wg := sync.WaitGroup{}
	queue := make(chan int, max_worker)

	wg_fetch_sdrf := sync.WaitGroup{}
	accession_queue := make(chan string, max_worker*7)

	folder_name := fmt.Sprintf("sdrf")

	if _, err := os.Stat(folder_name); err != nil {
		os.Mkdir(folder_name, 0755)
	}

	for i := 1; i <= max_worker; i++ {
		wg.Add(1)
		go worker_fetch_search(&wg, queue, accession_queue)
	}

	for i := 1; i <= max_worker; i++ {
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
