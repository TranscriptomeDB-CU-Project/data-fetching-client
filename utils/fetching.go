package utils

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func FetchWithRetry(url string, retry int) (*http.Response, *dtos.ErrorResponse) {
	if retry > 10000 {
		return nil, &dtos.ErrorResponse{
			Code:    500,
			Message: "Retry limit exceeded",
		}
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
		errMsg := fmt.Sprintf("fetch failed: %d", res.StatusCode)

		return nil, &dtos.ErrorResponse{
			Code:    res.StatusCode,
			Message: errMsg,
		}
	}

	return res, nil
}

func FetchSearch(species string, page int, target *dtos.SearchResult) *dtos.ErrorResponse {
	species_query := strings.ReplaceAll(species, " ", "+")

	if species_query != "" {
		species_query = fmt.Sprintf("&facet.organism=%s", species_query)
	}

	fetch_url := fmt.Sprintf("%s/arrayexpress/search?page=%d&pageSize=100%s", constants.API_URL, page, species_query)

	res, err := FetchWithRetry(fetch_url, 250)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Read body
	_err := json.NewDecoder(res.Body).Decode(target)

	if _err != nil {
		return &dtos.ErrorResponse{
			Code:    500,
			Message: _err.Error(),
		}
	}

	return nil
}

func FetchAccessionInfo(accession string, target *dtos.StudyInfo) *dtos.ErrorResponse {
	fetch_url := fmt.Sprintf("%s/studies/%s/info", constants.API_URL, accession)

	res, err := FetchWithRetry(fetch_url, 250)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	_err := json.NewDecoder(res.Body).Decode(target)

	if _err != nil {
		return &dtos.ErrorResponse{
			Code:    500,
			Message: _err.Error(),
		}
	}

	return nil
}

func FetchSDRFFileList(accession string) ([]string, *dtos.ErrorResponse) {
	fetch_url := fmt.Sprintf("%s/%s/%s.json", constants.FILE_BASE_URL, accession, accession)

	res, err := FetchWithRetry(fetch_url, 250)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	b, _err := io.ReadAll(res.Body)

	if err != nil {
		return nil, &dtos.ErrorResponse{
			Code:    500,
			Message: _err.Error(),
		}
	}

	file_name := ExtractSDRFFileName(string(b))

	return file_name, nil
}

func FetchAccessionSDRFFile(accession string, filename string) ([]byte, *dtos.ErrorResponse) {
	url := fmt.Sprintf("%s/%s/%s", constants.FILE_BASE_URL, accession, filename)

	res, err := FetchWithRetry(url, 250)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.ContentLength == 0 {
		return nil, &dtos.ErrorResponse{
			Code:    404,
			Message: "File not found",
		}
	}

	data_byte, _err := io.ReadAll(res.Body)

	if _err != nil {
		return nil, &dtos.ErrorResponse{
			Code:    500,
			Message: _err.Error(),
		}
	}

	return data_byte, nil
}
