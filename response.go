package sns

import (
	"fmt"
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
