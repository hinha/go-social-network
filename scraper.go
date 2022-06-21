package sns

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/cookiejar"
	"time"
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

	s.retries = 3
	s.conf = conf

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
	c.client.Timeout = time.Duration(timeout) * time.Second
	//req.URL.RawQuery = paramEncode
	urls += paramEncode
	req, err := http.NewRequest(method, urls, body)
	if err != nil {
		return nil, err
	}

	req.Header = header

	for i := 1; i < c.retries+1; i++ {
		response, err = c.client.Do(req)
		if err != nil {
			if i < c.retries {
				sleepTime := 1.0 * int(math.Pow(2, float64(i)))
				log.Printf("Waiting %v seconds", sleepTime)
				time.Sleep(time.Duration(sleepTime) * time.Second)
			}

			if i >= c.retries {
				msg := fmt.Sprintf("%d request to %s failed, giving up: %v", i, req.URL, err)
				log.Printf(msg)
				return nil, errors.New(msg)
			}
		}

	}

	if response == nil {
		return nil, fmt.Errorf("reached unreachable code: %v", err)
	}

	if cb != nil {
		ok, msg := cb(response)
		if ok {
			log.Println(urls, "retrieved successfully", msg)
		} else {
			// TODO logging
			//	if attempt < self._retries:
			//		retrying = ', retrying'
			//		level = logging.INFO
			//	else:
			//		retrying = ''
			//		level = logging.ERROR
			//	logger.log(level, f'Error retrieving {req.url}{msg}{retrying}')
		}
	}

	return response, nil
}
