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
	if retry > 10000 {
		return nil, fmt.Errorf("retry limit exceeded")
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	var res *http.Response
	var err error

	res, err = client.Get(url)

	// Network timeout or Rate Limit, let's do binary exponential backoff
	if err != nil || res.StatusCode == 500 {
		fmt.Printf("retry %s %s\n", url, err)

		time.Sleep(time.Duration(rand.Intn(retry)) * time.Millisecond)

		res.Body.Close()

		return FetchWithRetry(url, retry*2)
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("fetch failed: %s %s", res.Status, url)

		return nil, err
	}

	return res, nil
}

func FetchSearch(page int, target *dtos.SearchResult) error {
	fetch_url := fmt.Sprintf("%s/arrayexpress/search?page=%d&pageSize=100", constants.API_URL, page)

	res, err := FetchWithRetry(fetch_url, 250)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Read body
	err = json.NewDecoder(res.Body).Decode(target)

	return err
}

func FetchAccessionInfo(accession string, target *dtos.StudyInfo) error {
	fetch_url := fmt.Sprintf("%s/studies/%s/info", constants.API_URL, accession)

	res, err := FetchWithRetry(fetch_url, 250)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(target)

	return err
}

func FetchSDRFFileList(accession string) ([]string, error) {
	fetch_url := fmt.Sprintf("%s/%s/%s.json", constants.FILE_BASE_URL, accession, accession)

	res, err := FetchWithRetry(fetch_url, 250)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var x dtos.AccessionMetadata

	err = json.NewDecoder(res.Body).Decode(&x)

	if err != nil {
		return nil, err
	}

	file_name := ExtractSDRFFileName(&x)

	return file_name, nil
}

func FetchAccessionSDRFFile(accession string, filename string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%s", constants.FILE_BASE_URL, accession, filename)

	res, err := FetchWithRetry(url, 250)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.ContentLength == 0 {
		return nil, fmt.Errorf("empty body %s", filename)
	}

	data_byte, err := io.ReadAll(res.Body)

	return data_byte, err
}
