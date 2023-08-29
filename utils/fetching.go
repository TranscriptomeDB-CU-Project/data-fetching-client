package utils

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func FetchWithRetry(url string) (*http.Response, error) {
	var res *http.Response
	var err error

	retry := 250

	for {
		res, err = http.Get(url)

		// Network timeout or Rate Limit, let's do binary exponential backoff
		if err != nil || res.StatusCode == 429 {
			fmt.Printf("retry: %s\n", url)

			time.Sleep(time.Duration(retry) * time.Millisecond)

			retry *= 2
			continue
		}

		if res.StatusCode != 200 {
			err = fmt.Errorf("fetch failed: %s", res.Status)

			return nil, err
		}

		break
	}

	return res, nil
}

func FetchSearch(page int, target *dtos.SearchResult) error {
	fetch_url := fmt.Sprintf("%s/arrayexpress/search?page=%d&pageSize=100", constants.API_URL, page)

	res, err := FetchWithRetry(fetch_url)

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

	res, err := FetchWithRetry(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data_byte, err := io.ReadAll(res.Body)

	return data_byte, err
}
