package sns

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *TwitterScraper) CheckApiResponse(r *http.Response) (bool, string) {
	if r.StatusCode == http.StatusTooManyRequests || r.StatusCode == http.StatusForbidden {
		return false, fmt.Sprintf("blocked %d", r.StatusCode)
	}

	if r.Header.Get("content-type") == "" {
		return false, "Content type error"
	}

	if strings.ReplaceAll(r.Header.Get("content-type"), " ", "") != "application/json;charset=utf-8" {
		return false, "content type is not JSON"
	}

	if r.StatusCode != http.StatusOK {
		return false, "non-200 status code"
	}

	return true, ""
}

func (c *TwitterScraper) CheckTokenResponse(r *http.Response) (bool, string) {
	if r.StatusCode != http.StatusOK {
		c.tokenManager.Reset()
		return false, "non-200 response"
	}
	return true, ""
}

func decodeResponse(body io.ReadCloser, target interface{}) error {
	err := json.NewDecoder(body).Decode(target)
	if _, ok := err.(*json.SyntaxError); ok {
		return errors.New("received invalid JSON from Twitter")
	} else if err != nil {
		return fmt.Errorf("json.Decode: %v", err)
	}

	return nil
}
