package sns

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hinha/go-social-network/entities"
	"github.com/hinha/go-social-network/utils"
)

var (
	regexLink, _  = regexp.Compile(`href=[\'"]?([^\'" >]+)`)
	regexLabel, _ = regexp.Compile(`>([^<]*)<`)
	regexExt      = regexp.MustCompile("(\\.[^.]+)$") // extension
)

type parseTweets func(timeline twitterResponse, gotPinned *bool, dateRange DateRange) []*entities.TwitterPost

func checkEntries(instruction TweetInstructions) []interface{} {
	var entries interface{}
	if instruction.AddEntries != nil {
		entries = instruction.AddEntries["entries"]
	} else if instruction.ReplaceEntry != nil {
		replaceEntry := instruction.ReplaceEntry["entry"]
		entries = []interface{}{replaceEntry}
	} else if instruction.Type == "TimelineAddEntries" {
		entries = instruction.Entries
	} else {
		return nil
	}

	return entries.([]interface{})
}

func getTweetId(tweet TweetRaw) int {
	if tweet.Id != 0 {
		return tweet.Id
	} else {
		id, _ := strconv.Atoi(tweet.IdStr)
		return id
	}
}
func getUserId(user TweetUsers) int {
	if user.Id != 0 {
		return user.Id
	} else {
		id, _ := strconv.Atoi(user.IDStr)
		return id
	}
}

func renderTextWithUrls(text string, urls []TweetUrls) []string {
	if len(urls) == 0 {
		return []string{text}
	} else {
		var entitiesUrl []string
		for _, url := range urls {
			entitiesUrl = append(entitiesUrl, []string{url.Url, url.DisplayUrl, url.ExpandedURL}...)
		}
		return entitiesUrl
	}
}

func tweetCard(card *TweetRawCard, tweetID int, apiType VersionAPI) entities.TwitterCard {
	userRefs := make(map[string]entities.TwitterUser)
	if apiType == APIStandart {
		for key, _ := range card.Users {
			userRefs[key] = parseUser(card.Users[key], 0)
		}
	} else if apiType == APIGraphql {
		for _, o := range card.Legacy.UserRefs {
			if _, ok := userRefs[o.RestID]; ok {
				// duplicate user card
				continue
			}
			userID, _ := strconv.Atoi(o.RestID)
			if o.Legacy != nil {
				userRefs[o.RestID] = parseUser(*o.Legacy, userID)
			} else {
				userRefs[o.RestID] = entities.TwitterUser{Id: userID}
			}
		}
	}
	mapCard := entities.TwitterCard{}
	bindingValues := make(map[string]interface{})
	for key, value := range card.BindingValues {
		binding := value.(map[string]interface{})
		if _, ok := binding["type"]; !ok {
			continue
		}
		if binding["type"] == "STRING" {
			bindingValues[key] = binding["string_value"]
			if strings.HasSuffix(key, "_datetime_utc") {
				bindingValues[key], _ = time.Parse(time.RFC3339, bindingValues[key].(string))
			}
		} else if binding["type"] == "IMAGE" {
			bindingValues[key] = binding["image_value"].(map[string]interface{})["url"]
		} else if binding["type"] == "BOOLEAN" {
			bindingValues[key] = binding["boolean_value"]
		} else if binding["type"] == "IMAGE_COLOR" {
			// skip this section
		} else if binding["type"] == "USER" {
			bindingValues[key] = userRefs[binding["user_value"].(map[string]interface{})["id_str"].(string)]
		} else {
			log.Printf("WARN: Unsupported card value type on %s in tweet %d:%v", key, tweetID, binding["type"])
		}
	}

	var cardName string
	if apiType == APIStandart {
		cardName = card.Name
	} else if apiType == APIGraphql {
		cardName = card.Legacy.Name
	}

	if cardName == "summary" || cardName == "summary_large_image" || cardName == "app" || cardName == "direct_store_link_app" {
		if cardName == "app" || cardName == "direct_store_link_app" {
			mapCard.AppCard = bindingValues
		} else {
			mapCard.SummaryCard = bindingValues
		}
		//} else if strings.HasPrefix(card.Name, "poll2choice_") || strings.HasPrefix(card.Name, "poll3choice_") || strings.HasPrefix(card.Name, "poll4choice_") {
		//} else if card.Name == "745291183405076480:broadcast" || card.Name == "3691233323:periscope_broadcast" {
		//} else if card.Name == "appplayer" {
	} else if cardName == "player" {
		mapCard.PlayerCard = bindingValues
	} else if cardName == "3337203208:newsletter_publication" {
		panic("3337203208:newsletter_publication")
	} else if cardName == "3337203208:newsletter_issue" {
		panic("3337203208:newsletter_issue")
	} else if cardName == "amplify" {
		panic("amplify")
	} else if cardName == "appplayer" {
		panic("appplayer")
	}
	return mapCard
}

func makeTweet(tweet TweetRaw, user entities.TwitterUser, card entities.TwitterCard, posts ...interface{}) *entities.TwitterPost {
	tw := &entities.TwitterPost{
		Id:           getTweetId(tweet),
		ReplyCount:   tweet.ReplyCount,
		RetweetCount: tweet.RetweetCount,
		LikeCount:    tweet.FavoriteCount,
		QuoteCount:   tweet.QuoteCount,
		Lang:         tweet.Lang,
		Source:       tweet.Source,
	}
	{ // debug
		tw.Content = tweet.FullText
		tw.RenderedContent = strings.Join(renderTextWithUrls(tweet.FullText, tweet.Entities.URLs), "")
		tw.Date = utils.RubyDate(tweet.CreatedAt)
		for _, u := range tweet.Entities.URLs {
			tw.Links = append(tw.Links, entities.TwitterTextLink{
				Text:    u.DisplayUrl,
				Url:     u.ExpandedURL,
				TcoUrl:  u.Url,
				Indices: u.Indices,
			})
		}
		tw.Url = fmt.Sprintf("https://twitter.com/%s/status/%d", tw.User.Username, tw.Id)
		cvId, err := strconv.Atoi(tweet.ConversationIDStr)
		if err != nil {
			tw.ConversationId = tweet.ConversationID
		}
		tw.ConversationId = cvId
		if v := regexLink.FindStringSubmatch(tweet.Source); len(v) > 1 {
			tw.SourceUrl = v[1]
		}
		if v := regexLabel.FindStringSubmatch(tweet.Source); len(v) > 1 {
			tw.SourceLabel = v[1]
		}

		// data media
		for _, media := range tweet.ExtendedEntities.Media {
			if media.Type == "photo" {
				if strings.Contains(media.MediaURLHttps, "?format=") || strings.Contains(media.MediaURLHttps, "&format=") {
					tw.Media.Photo.PreviewUrl = media.MediaURLHttps
					tw.Media.Photo.FullUrl = media.MediaURLHttps
				}
				if !strings.Contains(media.MediaURLHttps, ".") {
					continue
				}
				format := regexExt.FindString(media.MediaURLHttps)[1:]
				baseUrl := regexExt.Split(media.MediaURLHttps, 1)[0] // url
				if format != "jpg" && format != "png" {
					continue
				}
				tw.Media.Photo.PreviewUrl = fmt.Sprintf("%s?format=%s&name=small", baseUrl, format)
				tw.Media.Photo.FullUrl = fmt.Sprintf("%s?format=%s&name=large", baseUrl, format)
			} else if media.Type == "video" || media.Type == "animated_gif" {
				if media.Type == "video" {
					for _, variant := range media.VideoInfo.Variants {
						tw.Media.Video.Variants = append(tw.Media.Video.Variants, entities.TwitterVideoVariant{
							ContentType: variant.ContentType,
							Url:         variant.URL,
							BitRate:     variant.Bitrate,
						})
					}
					tw.Media.Video.ThumbnailUrl = media.MediaURLHttps
					tw.Media.Video.Duration = float64(media.VideoInfo.DurationMillis / 1000)
					if r, ok := media.Ext.MediaStats["r"]; ok {
						if vok, ok := r.(map[string]interface{})["ok"]; ok {
							if count, ok := vok.(map[string]interface{})["viewCount"]; ok {
								switch v := count.(type) {
								case string:
									tw.Media.Video.Views, _ = strconv.Atoi(v)
								case int:
									tw.Media.Video.Views = count.(int)
								}
							}
						}
					}
				} else if media.Type == "animated_gif" {
					for _, variant := range media.VideoInfo.Variants {
						tw.Media.Gif.Variants = append(tw.Media.Gif.Variants, entities.TwitterVideoVariant{
							ContentType: variant.ContentType,
							Url:         variant.URL,
							BitRate:     variant.Bitrate,
						})
					}
					tw.Media.Gif.ThumbnailUrl = media.MediaURLHttps
				}
			}
		}
	}

	tw.InReplyToTweetId, _ = strconv.Atoi(tweet.InReplyToStatusIDStr)
	inReplyToUserIdStr, _ := strconv.Atoi(tweet.InReplyToUserIdStr)
	if inReplyToUserIdStr == tw.User.Id {
		tw.InReplyToUser = tw.User
	}
	if len(tweet.Entities.UserMentions) != 0 {
		for _, u := range tweet.Entities.UserMentions {
			if u.IDStr == tweet.InReplyToUserIdStr {
				tw.InReplyToUser = entities.TwitterUser{Username: u.ScreenName, Id: getUserId(u), DisplayName: u.Name}
			}
			tw.MentionedUsers = append(tw.MentionedUsers, entities.TwitterUser{Username: u.ScreenName, Id: getUserId(u), DisplayName: u.Name})
		}
	}
	if len(tweet.Entities.Hashtags) != 0 {
		for _, tag := range tweet.Entities.Hashtags {
			tw.Hashtags = append(tw.Hashtags, tag.Text)
		}
	}
	if len(tweet.Entities.Symbols) != 0 {
		for _, tag := range tweet.Entities.Symbols {
			tw.CashTags = append(tw.CashTags, tag.Text)
		}
	}
	if tw.InReplyToUser.Username == "" {
		tw.InReplyToUser = entities.TwitterUser{Username: tweet.InReplyToScreenName, Id: inReplyToUserIdStr}
	}
	// https://developer.twitter.com/en/docs/tutorials/filtering-tweets-by-location
	coordinates := new(entities.TwitterCoordinates)
	if tweet.Coordinates != nil {
		if len(tweet.Coordinates.Coordinates) == 2 {
			coordinates.Latitude = tweet.Coordinates.Coordinates[0]
			coordinates.Longitude = tweet.Coordinates.Coordinates[1]
			tw.Coordinates = coordinates
		}
	}
	if tweet.Geo != nil {
		if len(tweet.Geo.Coordinates) == 2 {
			coordinates.Latitude = tweet.Geo.Coordinates[0]
			coordinates.Longitude = tweet.Geo.Coordinates[1]
			tw.Coordinates = coordinates
		}
	}
	if tweet.Place != nil {
		tw.Place = &entities.TwitterPlace{
			FullName:    tweet.Place.FullName,
			Name:        tweet.Place.Name,
			Type:        tweet.Place.PlaceType,
			Country:     tweet.Place.Country,
			CountryCode: tweet.Place.CountryCode,
		}
		if tweet.Place.BoundingBox != nil && coordinates == nil {
			if len(tweet.Place.BoundingBox.Coordinates[0][0]) == 2 {
				// Take the first (longitude, latitude) couple of the "place square"
				coordinates.Latitude = tweet.Place.BoundingBox.Coordinates[0][0][0]
				coordinates.Longitude = tweet.Place.BoundingBox.Coordinates[0][0][1]
				tw.Coordinates = coordinates
			}
		}
	}
	tw.User = user
	tw.Card = card
	for _, t := range posts {
		if v, ok := utils.IsMapKey(t, "quoted_tweet"); ok {
			if data, ok := v.(*entities.TwitterPost); ok {
				tw.QuotedTweet = data
			}
			if data, ok := v.(*entities.TweetRef); ok {
				tw.QuotedTweetRef = data
			}
		}
		if v, ok := utils.IsMapKey(t, "retweeted_tweet"); ok {
			if data, ok := v.(*entities.TwitterPost); ok {
				tw.RetweetedTweet = data
			}
		}
	}

	return tw
}

func tweetToTweet(tweet TweetRaw, obj twitterResponse) *entities.TwitterPost {
	user := parseUser(obj.GlobalObjects.Users[tweet.UserIDStr], 0)
	// retweet data
	tweetList := make(map[string]*entities.TwitterPost)
	if tweet.RetweetedStatusIDStr != "" {
		tweetList["retweeted_tweet"] = tweetToTweet(obj.GlobalObjects.Tweets[tweet.RetweetedStatusIDStr], obj)
	}
	if tweet.QuotedStatusIDStr != "" {
		tweetList["quoted_tweet"] = tweetToTweet(obj.GlobalObjects.Tweets[tweet.QuotedStatusIDStr], obj)
	}

	var card entities.TwitterCard
	if tweet.Card != nil {
		card = tweetCard(tweet.Card, getTweetId(tweet), APIStandart)
	}
	return makeTweet(tweet, user, card, tweetList)
}

func retrieveTweetData(entryID string, content map[string]interface{}, obj twitterResponse) *entities.TwitterPost {
	var tweet TweetRaw
	if v, ok := content["tweet"]; ok {
		val := v.(map[string]interface{})
		if _, ok := val["promotedMetadata"]; ok {
			return nil
		}
		if _, ok := obj.GlobalObjects.Tweets[val["id"].(string)]; !ok {
			log.Println("WARN: skipping tweet", val["id"].(string), "which is not in globalObjects")
			return nil
		}
		tweet = obj.GlobalObjects.Tweets[val["id"].(string)]
	} else if v, ok := content["tombstone"]; ok {
		val := v.(map[string]interface{})
		if _, ok := val["tweet"]; ok { // E.g. deleted reply
			return nil
		}
		id := val["tweet"].(map[string]interface{})["id"].(string)
		if _, ok := obj.GlobalObjects.Tweets[id]; !ok {
			log.Println("WARN: skipping tweet", val["id"].(string), "which is not in globalObjects")
		}
		tweet = obj.GlobalObjects.Tweets[id]
	} else {
		//raise error
		log.Println("ERROR: unable to handle entry", entryID)
		return nil
	}

	return tweetToTweet(tweet, obj)
}

func retrieveGraphqlTimeline(result *utils.DictType) (*entities.TwitterPost, error) {
	if result.M("__typename").String() == "Tweet" {
	} else if result.M("__typename").String() == "TweetWithVisibilityResults" {
		panic("TODO")
	} else {
		return nil, fmt.Errorf("unknown result type %s", result.M("__typename").String())
	}
	userId, _ := strconv.Atoi(result.M("core").M("user_results").M("result").M("rest_id").String())
	legacy, err := json.Marshal(result.M("core").M("user_results").M("result").M("legacy").Interface())
	if err != nil {
		return nil, fmt.Errorf("json.Marshal unknown result type %s", result.M("__typename").String())
	}
	var userRaw TweetUsers
	_ = json.Unmarshal(legacy, &userRaw)
	user := parseUser(userRaw, userId)

	// tweet := result["legacy"].(map[string]interface{})
	tweet := result.M("legacy")
	tweetList := make(map[string]interface{})
	if v, ok := tweet.Exists("retweeted_status_result"); ok {
		retweeted, err := retrieveGraphqlTimeline(v.M("result"))
		if err == nil {
			tweetList["retweeted_tweet"] = retweeted
		}
	}

	if v, ok := result.Exists("quoted_status_result"); ok {
		if v.M("result").M("__typename").String() == "TweetTombstone" {
			panic("TweetTombstone")
		} else {
			quotedTweet, err := retrieveGraphqlTimeline(result)
			if err == nil {
				tweetList["quoted_tweet"] = quotedTweet
			}
		}
	} else if v, ok := result.Exists("quotedRefResult"); ok {
		tf := &entities.TweetRef{}
		if v.M("result").M("__typename").String() == "TweetTombstone" {
			tf.Id, _ = strconv.Atoi(tweet.M("quoted_status_id_str").String())
		} else {
			tf.Id, _ = strconv.Atoi(result.M("rest_id").String())
		}
		tf.SetUrl(tf.Id)
		tweetList["quoted_tweet"] = tf
	} else if v, ok := tweet.Exists("quoted_status_id_str"); ok {
		id, _ := strconv.Atoi(v.String())
		tf := &entities.TweetRef{Id: id}
		tf.SetUrl(id)
		tweetList["quoted_tweet"] = tf
	}

	var tc entities.TwitterCard
	if v, ok := result.Exists("card"); ok {
		raw, _ := json.Marshal(v.Interface())
		var card *TweetRawCard
		_ = json.Unmarshal(raw, &card)

		tweetId, _ := strconv.Atoi(tweet.M("id_str").String())
		tc = tweetCard(card, tweetId, APIGraphql)
	}

	js, err := json.Marshal(tweet.Interface())
	if err != nil {
		return nil, fmt.Errorf("json.Marshal unknown tweet result")
	}
	var tweetRaw TweetRaw
	if err := json.Unmarshal(js, &tweetRaw); err != nil {
		return nil, fmt.Errorf("json.Unmarshal unknown tweet result")
	}

	return makeTweet(tweetRaw, user, tc, tweetList), nil //todo panic error
}

func parseUser(user TweetUsers, userId int) entities.TwitterUser {

	entities := entities.TwitterUser{
		Id:               getUserId(user),
		Username:         user.ScreenName,
		DisplayName:      user.Name,
		RawDescription:   user.Description,
		Description:      strings.Join([]string{user.Description, strings.Join(renderTextWithUrls(user.Description, user.Entities.URL.Urls), ", ")}, " | "),
		DescriptionLinks: renderTextWithUrls(user.Description, user.Entities.URL.Urls),
	}
	entities.Created = utils.RubyDate(user.CreatedAt)
	entities.FollowersCount = user.FollowersCount
	entities.FriendsCount = user.FriendsCount
	entities.StatusesCount = user.StatusesCount
	entities.FavouritesCount = user.FavouritesCount
	entities.ListedCount = user.ListedCount
	entities.MediaCount = user.MediaCount
	entities.Location = user.Location
	entities.Protected = user.Protected
	entities.Verified = user.Verified
	entities.Url = user.Url
	entities.ProfileImageURL = user.ProfileImageURLHTTPS
	entities.ProfileBannerURL = user.ProfileBannerURL
	if user.Ext.HighlightedLabel.R.Ok.Label != nil {
		entities.Label.Description = user.Ext.HighlightedLabel.R.Ok.Label.Description
		if user.Ext.HighlightedLabel.R.Ok.Label.Url != nil {
			entities.Label.Url = user.Ext.HighlightedLabel.R.Ok.Label.Url["url"].(string)
		}
		if user.Ext.HighlightedLabel.R.Ok.Label.Badge != nil {
			entities.Label.Badge = user.Ext.HighlightedLabel.R.Ok.Label.Badge["url"].(string)
		}
		if user.Ext.HighlightedLabel.R.Ok.Label.LongDescription != nil {
			entities.Label.LongDescription = user.Ext.HighlightedLabel.R.Ok.Label.LongDescription["text"].(string)
		}
	}
	return entities
}

func parseTimeline(timeline twitterResponse, gotPinned *bool, dateRange DateRange) []*entities.TwitterPost {
	tweets := make([]*entities.TwitterPost, 0)

	for _, instruction := range timeline.Timeline.Instructions {
		entries := checkEntries(instruction)
		if entries == nil {
			continue
		}

		for _, obj := range entries {
			entry := obj.(map[string]interface{})
			if !(strings.HasPrefix(entry["entryId"].(string), "sq-I-t-") || strings.HasPrefix(entry["entryId"].(string), "tweet-")) {
				continue
			}
			tweets = append(tweets, retrieveTweetData(entry["entryId"].(string), entry["content"].(map[string]interface{})["item"].(map[string]interface{})["content"].(map[string]interface{}), timeline))
		}
	}
	return tweets
}

func parseTimelineV2(timeline twitterResponse, gotPinned *bool, dateRange DateRange) []*entities.TwitterPost {
	tweets := make([]*entities.TwitterPost, 0)
	if !*gotPinned {
		for _, instruction := range timeline.Data.User.Result.Timeline.Timeline.Instructions {
			if len(instruction.Entries) == 0 {
				continue
			}

			if instruction.Type == "TimelinePinEntry" {
				setPinned := true
				gotPinned = &setPinned
				panic("TimelinePinEntry")
				// parse timeline func
			}
		}
	}
	// func
	for _, instruction := range timeline.Data.User.Result.Timeline.Timeline.Instructions {
		entries := checkEntries(instruction)
		if entries == nil {
			continue
		}

		if instruction.Type != "TimelineAddEntries" {
			continue
		}

		for _, obj := range entries {
			entry := utils.Dict(obj)

			if strings.HasPrefix(entry.M("entryId").String(), "tweet-") {
				entryType := entry.M("content").M("entryType").String()
				itemType := entry.M("content").M("itemContent").M("itemType").String()
				if entryType == "TimelineTimelineItem" && itemType == "TimelineTweet" {
					result, err := retrieveGraphqlTimeline(entry.M("content").M("itemContent").M("tweet_results").M("result"))
					if err != nil {
						continue
					}

					start, _ := time.Parse(datetimeLayout, dateRange.Since)
					end, _ := time.Parse(datetimeLayout, dateRange.Until)
					if inTimeSpan(start, end, *result.Date) {
						tweets = append(tweets, result)
					}

				} else {
					log.Println("WARN: got unrecognised timeline tweet item(s)")
				}
			}
		}
	}
	return tweets
}

func parseTrends(timeline twitterResponse) []*entities.TwitterTrend {
	var trends []*entities.TwitterTrend
	urlEncode := url.Values{}
	for _, instruction := range timeline.Timeline.Instructions {
		entries := checkEntries(instruction)
		if entries == nil {
			continue
		}

		for _, obj := range entries {
			entry := utils.Dict(obj)
			if entry.M("entryId").String() != "trends" {
				continue
			}

			for _, item := range entry.M("content").M("timelineModule").M("items").Slice() {
				trend := utils.Dict(item).M("item").M("content").M("trend")
				trendName := trend.M("name").String()
				urlEncode.Set("q", trendName)

				var metaDescription string
				if meta, ok := trend.M("trendMetadata").Exists("metaDescription"); ok {
					metaDescription = meta.String()
				}
				trends = append(trends, &entities.TwitterTrend{
					Name:            trendName,
					MetaDescription: metaDescription,
					DomainContext:   trend.M("trendMetadata").M("domainContext").String(),
					Url:             fmt.Sprintf("https://twitter.com/search?%s", urlEncode.Encode()),
				})
			}
		}
	}

	return trends
}
func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}
