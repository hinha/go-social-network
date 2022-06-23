package sns

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/hinha/go-social-network/logger"
)

type DateRange struct {
	Since string
	Until string
}

type (
	callbackResponse func(r *http.Response) (bool, string)

	Lang int
)

const (
	dateLayout     = "2006-01-02"
	datetimeLayout = "2006-01-02 15:04:05"

	// LangID Lang - default language indonesian
	LangID Lang = iota
	LangEn
)

// Config struct contain extra config of twitter
type Config struct {
	// Logger
	Logger logger.Interface

	Date DateRange
	Lang Lang
	SearchMode
}

// Scraper object
type Scraper struct {
	conf    *Config
	client  *http.Client
	retries int
}

func newScraper(conf *Config) *Scraper {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	s := &Scraper{
		client: &http.Client{
			Timeout: 10 * time.Second, // default timeout
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: jar,
		},
	}

	if conf != nil {
		if _, err := time.Parse(dateLayout, conf.Date.Since); err != nil {
			panic(fmt.Errorf("since %v", err))
		} else {
			conf.Date.Since += " 00:00:00"
		}

		if _, err := time.Parse(dateLayout, conf.Date.Until); err != nil {
			panic(fmt.Errorf("until %v", err))
		} else {
			conf.Date.Until += " 23:59:59"
		}
		s.conf = conf
	}
	s.retries = 3

	return s
}

func (c *Scraper) RequestGET(url string, paramEncode string, header http.Header, cb callbackResponse) (response *http.Response, err error) {
	return c.newRequest("GET", url, paramEncode, nil, header, 10, cb)
}

func (c *Scraper) RequestPOST(url string, paramEncode string, body io.Reader, header http.Header, cb callbackResponse) (response *http.Response, err error) {
	return c.newRequest("POST", url, paramEncode, body, header, 10, cb)
}

func (c *Scraper) GetClient() *http.Client {
	return c.client
}

func (c *Scraper) newRequest(method, urls string, paramEncode string, body io.Reader, header http.Header, timeout int, cb callbackResponse) (response *http.Response, err error) {
	currentLogger, newLogger := c.conf.Logger, logger.Recorder.New()

	c.client.Timeout = time.Duration(timeout) * time.Second
	urls += paramEncode
	req, err := http.NewRequest(method, urls, body)
	if err != nil {
		return nil, err
	}
	currentLogger.Init(map[string]interface{}{
		"subject": "request",
		"method":  method,
		"path":    req.URL.Path,
	})

	req.Header = header
	var redirection []string
	for i := 1; i < c.retries+1; i++ {
		response, err = c.client.Do(req)
		redirection = append(redirection, redirectionUrl(response)) // check redirect url
		if err != nil {
			if cb != nil {
				if len(redirection) != 0 {
					for i, redirect := range redirection {
						directLog := fmt.Sprintf("Request %d: %s: %d (Location: %v)", i, redirect, response.StatusCode, response.Header.Get("location"))
						currentLogger.Debug(newLogger.BeginAt, directLog)
					}
				}

				ok, msg := cb(response)
				if ok {
					currentLogger.Debug(newLogger.BeginAt, "retrieved successfully", msg)
					break
				} else {
					if i < c.retries {
						currentLogger.Info(newLogger.BeginAt, "Error retrieving ", req.URL.String(), ", retrying")
					} else {
						currentLogger.Error(newLogger.BeginAt, "Error retrieving", req.URL.String())
					}
				}
			}

			if i < c.retries {
				sleepTime := 1.0 * int(math.Pow(2, float64(i)))
				currentLogger.Info(newLogger.BeginAt, fmt.Sprintf("Waiting %v seconds", sleepTime))
				time.Sleep(time.Duration(sleepTime) * time.Second)
			}

		}
	}

	currentLogger.Trace(req.Context(), newLogger.BeginAt, func() (string, int) {
		return "", response.StatusCode
	}, err)

	if response == nil || err != nil {
		currentLogger.Error(newLogger.BeginAt, fmt.Sprintf("%d request to %s failed, giving up: %v", c.retries+1, req.URL, err))
		return nil, fmt.Errorf("reached unreachable code: %v", err)
	}

	return response, nil
}

func redirectionUrl(resp *http.Response) string {
	if resp.StatusCode == http.StatusMovedPermanently {
		return resp.Request.URL.String()
	}
	return ""
}
