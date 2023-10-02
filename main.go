package main

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"arrayexpress-fetch/utils"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	constants.LoadFileBasePath()

	godotenv.Load("./.env")

	mongo_uri := os.Getenv("MONGO_URI")

	client, _err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongo_uri))

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if _err != nil {
		log.Fatal(_err)
	}

	var body dtos.SearchResult

	start := time.Now()

	accession_metadata := sync.Map{}
	time_stamp := utils.ReadMetadata(client)

	err := utils.FetchSearch("homo sapiens", 1, &body)

	if err != nil {
		fmt.Println(err)
		return
	}

	totalPages := int(math.Min(math.Ceil(float64(body.TotalHits)/float64(body.PageSize)), 1))

	wg := sync.WaitGroup{}
	queue := make(chan int, constants.FETCH_SEARCH_WORKER)

	wg_fetch_sdrf := sync.WaitGroup{}
	accession_queue := make(chan string, constants.FETCH_FILE_WORKER)

	folder_name := fmt.Sprintf("%ssdrf", constants.FILE_BASE_PATH)

	if _, err := os.Stat(folder_name); err != nil {
		os.Mkdir(folder_name, 0755)
	}

	for i := 1; i <= constants.FETCH_SEARCH_WORKER; i++ {
		wg.Add(1)
		go utils.WorkerFetchSearch("homo sapiens", &wg, queue, accession_queue)
	}

	for i := 1; i <= constants.FETCH_FILE_WORKER; i++ {
		wg_fetch_sdrf.Add(1)
		go utils.WorkerFetchAccession(&wg_fetch_sdrf, accession_queue, &accession_metadata, time_stamp)
	}

	for i := 1; i <= totalPages; i++ {
		queue <- i
	}

	close(queue)
	wg.Wait()

	close(accession_queue)
	wg_fetch_sdrf.Wait()

	utils.WriteMetadata(time_stamp, &accession_metadata, client)

	log.Printf("Time took: %s", time.Since(start))
}
