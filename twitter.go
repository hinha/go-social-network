package sns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hinha/go-social-network/entities"
	"github.com/hinha/go-social-network/logger"
	"github.com/hinha/go-social-network/utils"
)

const (
	ApiAuthorizationHeader   = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
	TwitterAPIToken          = "https://api.twitter.com/1.1/guest/activate.json"
	TwitterAPISearch         = "https://api.twitter.com/2/search/adaptive.json"
	TwitterAPITrends         = "https://twitter.com/i/api/2/guide.json"
	TwitterAPIUserScreenName = "https://twitter.com/i/api/graphql/7mjxD3-C6BxitPMVQ6w0-Q/UserByScreenName"
	TwitterAPIUserTweets     = "https://twitter.com/i/api/graphql/BSKxQ9_IaCoVyIvQHQROIQ/UserTweetsAndReplies"
)

// var (
// 	reHashtag    = regexp.MustCompile(`\B(\#\S+\b)`)
// 	reTwitterURL = regexp.MustCompile(`https:(\/\/t\.co\/([A-Za-z0-9]|[A-Za-z]){10})`)
// 	reUsername   = regexp.MustCompile(`\B(\@\S{1,15}\b)`)
// )

type SearchMode int

const (
	SearchKeyword SearchMode = iota
	SearchUser
	SearchHashtag
)

type VersionAPI string

const (
	APIStandart VersionAPI = "API.v2"
	APIGraphql  VersionAPI = "GRAPHQL.v2"
)

var (
	wg sync.WaitGroup
)

type TwitterScraper struct {
	scraper    *Scraper
	config     *Config
	apiHeaders http.Header
	userAgent  string
	//guestToken string
	query        string
	tokenManager *utils.GuestTokenManager
	ensureToken  struct {
		baseUrl string
		params  string
	}
}

func NewTwitterScraper(conf *Config) *TwitterScraper {
	s := &TwitterScraper{}

	if conf != nil {
		if conf.Logger == nil {
			conf.Logger = logger.Default
		}
		conf.Logger.SetField("media", "twitter")
		s.scraper = newScraper(conf)
		s.config = conf
	}

	rand.Seed(time.Now().UnixNano())
	header := http.Header{}
	header.Set("Authorization", ApiAuthorizationHeader)
	header.Set("Accept-Language", "en-US,en;q=0.5")
	s.apiHeaders = header
	s.randomUserAgent()
	s.tokenManager = utils.TokenManager()

	return s
}

func (c *TwitterScraper) randomUserAgent() {
	c.userAgent = fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.%d Safari/537.%d", rand.Intn(9999), rand.Intn(99))
	c.apiHeaders.Set("User-Agent", c.userAgent)
}

func (c *TwitterScraper) ensureGuestToken(baseUrl string) error {
	beginAt := time.Now()

	header := http.Header{}
	if c.tokenManager.Token == "" {
		header.Add("User-Agent", c.apiHeaders.Get("User-Agent"))
		r, err := c.scraper.RequestGET(baseUrl, "", header, c.CheckTokenResponse)
		if err != nil {
			c.config.Logger.Error(beginAt, err)
			return err
		}
		defer r.Body.Close()
		c.config.Logger.SetField("subject", "token")

		resp, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		full, _ := regexp.Compile(`document\.cookie = decodeURIComponent\("gt=(\d+); Max-Age=10800; Domain=\.twitter\.com; Path=/; Secure"\);`)
		match := full.FindStringSubmatch(string(resp))
		if len(match) > 1 {
			c.config.Logger.Debug(beginAt, "Found guest token in HTML")
			c.tokenManager.SetToken(match[1])
		}
		for _, cookie := range r.Cookies() {
			c.config.Logger.Debug(beginAt, "Found guest token in cookies")
			if cookie.Name == "gt" {
				c.tokenManager.SetToken(cookie.Value)
				break
			}
		}
		if c.tokenManager.Token == "" {
			c.config.Logger.Debug(beginAt, "No guest token in response")
			c.config.Logger.Info(beginAt, "Retrieving guest token via API")
			r, err := c.scraper.RequestPOST(TwitterAPIToken, "", bytes.NewReader([]byte("")), c.apiHeaders, c.CheckTokenResponse)
			if err != nil {
				return err
			}
			defer r.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
				return err
			}

			if val, ok := result["guest_token"]; ok {
				c.tokenManager.SetToken(val.(string))
			}
		}
		c.config.Logger.Debug(beginAt, "Using guest token ", c.tokenManager.Token)
	}
	cookie := make([]*http.Cookie, 0)
	cookie = append(cookie, &http.Cookie{
		Name:    "gt",
		Domain:  ".twitter.com",
		Path:    "/",
		Secure:  true,
		Value:   c.tokenManager.Token,
		Expires: c.tokenManager.GetTime().Add(time.Duration(10800) * time.Second),
	})
	URL, _ := url.Parse(baseUrl)
	c.scraper.GetClient().Jar.SetCookies(URL, cookie)
	c.apiHeaders.Add("x-guest-token", c.tokenManager.Token)
	return nil
}

func (c *TwitterScraper) get_api_data(endpoint string, params url.Values, apiType VersionAPI) (twitterResponse, error) {
	var paramsEncode string
	if apiType == APIStandart {
		if err := c.ensureGuestToken(c.ensureToken.baseUrl + c.ensureToken.params); err != nil {
			return twitterResponse{}, err
		}
		paramsEncode = params.Encode()
	} else if apiType == APIGraphql {
		mapParams := make(map[string]string)
		for k, v := range params {
			mapParams[k] = v[0]
		}
		strMap, _ := json.Marshal(mapParams)
		paramsEncode += "variables=" + url.PathEscape(string(strMap))
	}

	resp, err := c.scraper.RequestGET(endpoint, paramsEncode, c.apiHeaders, c.CheckTokenResponse)
	if err != nil {
		return twitterResponse{}, err
	}
	defer resp.Body.Close()

	var tweetResult twitterResponse
	if err := decodeResponse(resp.Body, &tweetResult); err != nil {
		return twitterResponse{}, err
	}

	return tweetResult, nil
}

func (c *TwitterScraper) iteratorApiData(ctx context.Context, endpoint string, params url.Values, paginationParams url.Values, cursor string, maxTweet int, apiType VersionAPI, channel chan *TweetResult, fn parseTweets) {
	beginAt := time.Now()
	defer close(channel)

	var reqParams url.Values
	if cursor == "" {
		reqParams = params
	} else {
		reqParams = paginationParams
		reqParams.Add("cursor", cursor)
	}

	var bottomCursorAndStop interface{}
	var stopOnEmptyResponse bool
	var emptyResponseOnCursor int
	var tweetNum int

	for {
		c.config.Logger.Info(beginAt, "Retrieving scroll page ", cursor)
		obj, err := c.get_api_data(endpoint, reqParams, apiType)
		if err != nil {
			channel <- &TweetResult{Error: err}
			return
		}

		var goPinned bool

		wg.Add(1)
		go func(obj twitterResponse) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				channel <- &TweetResult{Error: ctx.Err()}
				return
			default:
			}

			tweets := fn(obj, &goPinned, c.config.Date)
			if len(tweets) == 0 {
				return
			}

			for _, tweet := range tweets {
				select {
				case <-ctx.Done():
					if tweet != nil {
						channel <- &TweetResult{Error: ctx.Err()}
						return
					}
				default:
				}

				if tweetNum < maxTweet {
					channel <- &TweetResult{TwitterPost: tweet}
				} else {
					return
				}
				tweetNum++
			}
		}(obj)
		wg.Wait()

		if tweetNum == maxTweet {
			break
		}

		var newCursor string
		var promptCursor string
		var newBottomCursorAndStop interface{}
		var tweetCount int

		var instructions []TweetInstructions
		if apiType == APIStandart {
			instructions = obj.Timeline.Instructions
		} else if apiType == APIGraphql {
			if obj.Data.User != nil {
				instructions = obj.Data.User.Result.Timeline.Timeline.Instructions
			} else {
				panic("todo")
				//instructions = obj.Data.ThreadedConversationWithInjections.Instructions
			}
		}

		for _, instruction := range instructions {
			entries := checkEntries(instruction)
			if entries == nil {
				continue
			}

			for _, obj := range entries {
				entry := utils.Dict(obj)

				if strings.HasPrefix(entry.M("entryId").String(), "sq-I-t-") || strings.HasPrefix(entry.M("entryId").String(), "tweet-") {
					tweetCount += 1
				}

				if !(strings.HasPrefix(entry.M("entryId").String(), "sq-cursor-") || strings.HasPrefix(entry.M("entryId").String(), "cursor-")) {
					continue
				}

				var entryCursor string
				var entryCursorStop bool
				if apiType == APIStandart {
					entryCursor = entry.M("content").M("operation").M("cursor").M("value").String()
					if _, ok := entry.M("content").M("operation").M("cursor").Exists("stopOnEmptyResponse"); ok {
						entryCursorStop = entry.M("content").M("operation").M("cursor").M("stopOnEmptyResponse").Bool()
					}
				} else if apiType == APIGraphql {
					cursorContent := entry.M("content").MapInterface()
					var itemType string
					var entryType string

					if v, ok := cursorContent["itemType"].(string); ok {
						itemType = v
					}
					if v, ok := cursorContent["entryType"].(string); ok {
						entryType = v
					}
					for itemType == "TimelineTimelineItem" || entryType == "TimelineTimelineItem" {
						cursorContent = cursorContent["itemContent"].(map[string]interface{})
					}
					entryCursor = cursorContent["value"].(string)
					if v, ok := cursorContent["stopOnEmptyResponse"]; ok {
						entryCursorStop = v.(bool)
					}

					if entryCursor == "" {
						c.config.Logger.Debug(beginAt, "emtpty cursor")
					}
				}

				if entry.M("entryId").String() == "sq-cursor-bottom" || strings.HasPrefix(entry.M("entryId").String(), "cursor-bottom-") {
					newCursor = entryCursor
					stopOnEmptyResponse = entryCursorStop
				} else if strings.HasPrefix(entry.M("entryId").String(), "cursor-showMoreThreadsPrompt-") {
					promptCursor = entryCursor
					panic("prompt")
				} else if bottomCursorAndStop == nil && (entry.M("entryId").String() == "sq-cursor-bottom" || strings.HasPrefix(entry.M("entryId").String(), "cursor-bottom-")) {
					newBottomCursorAndStop = entryCursor
					panic(entryCursor)
					//entryCursorStop.(bool) = true
				}
			}
		}

		if bottomCursorAndStop == nil && newBottomCursorAndStop != nil {
			panic(newBottomCursorAndStop)
		}
		if newCursor == cursor && tweetCount == 0 {
			emptyResponseOnCursor += 1
			if emptyResponseOnCursor > c.scraper.retries {
				break
			}
		}

		if newCursor == "" || (stopOnEmptyResponse && tweetCount == 0) {
			// end of pagination
			if promptCursor != "" {
				newCursor = promptCursor
			} else if bottomCursorAndStop != nil {
				panic("bottomCursorAndStop")
			} else {
				break
			}
		}

		cursor = newCursor
		reqParams = paginationParams
		reqParams.Set("cursor", cursor)
	}
}

func (c *TwitterScraper) TweetSearch(ctx context.Context, query string, maxTweets int) <-chan *TweetResult {
	channel := make(chan *TweetResult)

	c.query = query
	query, paginationParams := c.params(query)
	paginationParams.Add("q", query)
	paginationParams.Add("f", "top")

	// copy value
	params := paginationParams
	params.Del("cursor")

	// make ensure token
	param := url.Values{}
	param.Add("q", c.query)
	param.Add("f", "live")
	param.Add("lang", "en")
	param.Add("src", "spelling_expansion_revert_click")
	c.ensureToken.baseUrl = "https://twitter.com/search?"
	c.ensureToken.params = params.Encode()

	go c.iteratorApiData(ctx, TwitterAPISearch+"?", params, paginationParams, "", maxTweets, APIStandart, channel, parseTimeline)
	return channel
}

func (c *TwitterScraper) TweetHastag(ctx context.Context, hashtag string, maxTweets int) <-chan *TweetResult {
	channel := make(chan *TweetResult)

	c.query = hashtag
	query, paginationParams := c.params(hashtag)
	paginationParams.Add("q", "#"+query)
	paginationParams.Add("f", "top")

	// copy value
	params := paginationParams
	params.Del("cursor")

	// make ensure token
	param := url.Values{}
	param.Add("q", c.query)
	param.Add("f", "live")
	param.Add("lang", "en")
	param.Add("src", "spelling_expansion_revert_click")
	c.ensureToken.baseUrl = "https://twitter.com/search?"
	c.ensureToken.params = params.Encode()

	go c.iteratorApiData(ctx, TwitterAPISearch+"?", params, paginationParams, "", maxTweets, APIStandart, channel, parseTimeline)
	return channel
}

func (c *TwitterScraper) TweetUser(ctx context.Context, username string, maxTweets int) <-chan *TweetResult {
	channel := make(chan *TweetResult)

	baseUrl := "https://twitter.com/i/user/" + username
	if err := c.ensureGuestToken(baseUrl); err != nil {
		panic(err)
	}

	paramsStr := "variables=%7B%22screen_name%22%3A%22" + username + "%22%2C%22withSafetyModeUserFields%22%3Atrue%2C%22withSuperFollowsUserFields%22%3Atrue%7D"
	resp, err := c.scraper.RequestGET(TwitterAPIUserScreenName+"?", paramsStr, c.apiHeaders, c.CheckTokenResponse)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result TweetGraphqlUser
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}
	if result.Data.User.Result.Typename == "UserUnavailable" {
		return channel
	}

	// TODO
	user := result.Data.User.Result
	//rawDescription := strings.Join([]string{user.Legacy.Description, strings.Join(renderTextWithUrls(user.Legacy.Description, user.Legacy.Entities.Description.Urls), ", ")}, " | ")
	//fmt.Println(rawDescription)

	paginationVariables := url.Values{}
	paginationVariables.Add("userId", user.RestID)
	paginationVariables.Add("count", "100")
	paginationVariables.Add("cursor", "")
	paginationVariables.Add("includePromotedContent", "true")
	paginationVariables.Add("withCommunity", "true")
	paginationVariables.Add("withSuperFollowsUserFields", "true")
	paginationVariables.Add("withDownvotePerspective", "false")
	paginationVariables.Add("withReactionsMetadata", "false")
	paginationVariables.Add("withReactionsPerspective", "false")
	paginationVariables.Add("withSuperFollowsTweetFields", "true")
	paginationVariables.Add("withVoice", "true")
	paginationVariables.Add("withV2Timeline", "false")

	variables := paginationVariables
	variables.Del("cursor")

	go c.iteratorApiData(ctx, TwitterAPIUserTweets+"?", variables, paginationVariables, "", maxTweets, APIGraphql, channel, parseTimelineV2)
	return channel
}

func (c *TwitterScraper) TweetTrends(ctx context.Context) ([]*entities.TwitterTrend, error) {
	_, paginationParams := c.params("")
	paginationParams.Del("cursor")
	paginationParams.Del("tweet_search_mode")
	paginationParams.Del("query_source")
	paginationParams.Del("pc")
	paginationParams.Del("spelling_corrections")
	paginationParams.Set("count", "20")
	paginationParams.Add("candidate_source", "trends")
	paginationParams.Add("include_page_configuration", "dalse")
	paginationParams.Add("entity_tokens", "false")
	paginationParams.Set("ext", "mediaStats,highlightedLabel,voiceInfo")

	// make ensure token
	c.ensureToken.baseUrl = "https://twitter.com/i/trends"
	c.ensureToken.params = ""

	response, err := c.get_api_data(TwitterAPITrends+"?", paginationParams, APIStandart)
	if err != nil {
		return nil, err
	}

	return parseTrends(response), nil
}

func (c *TwitterScraper) params(query string) (string, url.Values) {
	paginationParams := url.Values{}
	switch c.config.Lang {
	case LangID:
		query = fmt.Sprintf("%s lang:%s", query, "id")
	case LangEn:
		query = fmt.Sprintf("%s lang:%s", query, "en")
	}
	if c.config.Date.Since != "" && c.config.Date.Until != "" {
		spSince := strings.Split(c.config.Date.Since, " ")
		spUntil := strings.Split(c.config.Date.Until, " ")
		query += fmt.Sprintf(" until:%s since:%s ", spUntil[0], spSince[0])
	}
	paginationParams.Add("include_profile_interstitial_type", "1")
	paginationParams.Add("include_blocking", "1")
	paginationParams.Add("include_blocked_by", "1")
	paginationParams.Add("include_followed_by", "1")
	paginationParams.Add("include_want_retweets", "1")
	paginationParams.Add("include_mute_edge", "1")
	paginationParams.Add("include_can_dm", "1")
	paginationParams.Add("include_can_media_tag", "1")
	paginationParams.Add("skip_status", "1")
	paginationParams.Add("cards_platform", "Web-12")
	paginationParams.Add("include_cards", "1")
	paginationParams.Add("include_ext_alt_text", "true")
	paginationParams.Add("include_quote_count", "true")
	paginationParams.Add("include_reply_count", "1")
	paginationParams.Add("tweet_mode", "extended")
	paginationParams.Add("include_entities", "true")
	paginationParams.Add("include_user_entities", "true")
	paginationParams.Add("include_ext_media_color", "true")
	paginationParams.Add("include_ext_media_availability", "true")
	paginationParams.Add("send_error_codes", "true")
	paginationParams.Add("simple_quoted_tweets", "true")
	paginationParams.Add("tweet_search_mode", "live")
	paginationParams.Add("count", "100")
	paginationParams.Add("query_source", "spelling_expansion_revert_click")
	paginationParams.Add("cursor", "")
	paginationParams.Add("pc", "1")
	paginationParams.Add("spelling_corrections", "1")
	paginationParams.Add("ext", "mediaStats,highlightedLabel")

	return query, paginationParams
}
