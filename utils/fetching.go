package utils

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func FetchWithRetry(url string, retry int) (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var res *http.Response
	var err error

	res, err = client.Get(url)

	// Network timeout or Rate Limit, let's do binary exponential backoff
	if err != nil || res.StatusCode == 429 {
		fmt.Printf("retry %s %s\n", url, err)

		time.Sleep(time.Duration(rand.Intn(retry)) * time.Millisecond)

		return FetchWithRetry(url, retry*2)
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("fetch failed: %s", res.Status)

		return nil, err
	}

	return res, nil
}

func FetchSearch(page int, target *dtos.SearchResult) error {
	fetch_url := fmt.Sprintf("%s/arrayexpress/search?page=%d&pageSize=50", constants.API_URL, page)

	res, err := FetchWithRetry(fetch_url, 250)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Read body
	err = json.NewDecoder(res.Body).Decode(target)

	return err
}

func FetchAccession(accession string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%s.sdrf.txt", constants.FILE_BASE_URL, accession, accession)

	res, err := FetchWithRetry(url, 250)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data_byte, err := io.ReadAll(res.Body)

	return data_byte, err
}
