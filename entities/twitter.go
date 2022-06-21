package entities

import (
	"fmt"
	"time"
)

type TwitterTextLink struct {
	Text    string `json:"text"`
	Url     string `json:"url"`
	TcoUrl  string `json:"tcourl"`
	Indices []int  `json:"indices"`
}

type TwitterPlace struct {
	FullName    string `json:"full_name"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

type TwitterCoordinates struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type TwitterVideoVariant struct {
	ContentType string `json:"content_type"`
	Url         string `json:"url"`
	BitRate     int    `json:"bit_rate"`
}

type TwitterCard struct {
	SummaryCard map[string]interface{} `json:"summary_card"`
	AppCard     map[string]interface{} `json:"app_card"`
	PlayerCard  map[string]interface{} `json:"player_card"`
}

type twitterSummaryCard struct {
	Title        string      `json:"title"`
	Url          string      `json:"url"`
	Description  string      `json:"description"`
	ThumbnailUrl string      `json:"thumbnail_url"`
	SiteUser     TwitterUser `json:"site_user"`
	CreatorUser  TwitterUser `json:"creator_user"`
}

type TwitterPost struct {
	Id              int               `json:"id"`
	Url             string            `json:"url"`
	Date            *time.Time        `json:"date"`
	RenderedContent string            `json:"rendered_content"`
	ReplyCount      int               `json:"reply_count"`
	RetweetCount    int               `json:"retweet_count"`
	LikeCount       int               `json:"like_count"`
	QuoteCount      int               `json:"quote_count"`
	ConversationId  int               `json:"conversation_id"`
	Lang            string            `json:"lang"`
	Source          string            `json:"source"`
	SourceUrl       string            `json:"source_url"`
	SourceLabel     string            `json:"source_label"`
	Content         string            `json:"content"`
	Links           []TwitterTextLink `json:"links"`
	Media           struct {
		Photo struct {
			PreviewUrl string `json:"preview_url"`
			FullUrl    string `json:"full_url"`
		} `json:"photo"`
		Video struct {
			ThumbnailUrl string                `json:"thumbnail_url"`
			Variants     []TwitterVideoVariant `json:"variants"`
			Duration     float64               `json:"duration"`
			Views        int                   `json:"views"`
		}
		Gif struct {
			ThumbnailUrl string                `json:"thumbnail_url"`
			Variants     []TwitterVideoVariant `json:"variants"`
		} `json:"gif"`
	} `json:"media"`
	RetweetedTweet *TwitterPost `json:"retweeted_tweet"`

	QuotedTweet          *TwitterPost        `json:"quoted_tweet"`
	QuotedTweetRef       *TweetRef           `json:"quoted_tweet_ref"`
	InReplyToTweetId     int                 `json:"in_reply_to_tweet_id"`
	InReplyToStatusIdStr string              `json:"in_reply_to_status_id_str"`
	InReplyToUser        TwitterUser         `json:"in_reply_to_user"`
	MentionedUsers       []TwitterUser       `json:"mentioned_users"`
	Coordinates          *TwitterCoordinates `json:"coordinates"`
	Place                *TwitterPlace       `json:"place"`
	Hashtags             []string            `json:"hashtags"`
	CashTags             []string            `json:"cash_tags"`
	Card                 TwitterCard         `json:"card"`
	User                 TwitterUser         `json:"user"`
}

type TweetRef struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

func (t *TweetRef) SetUrl(id int) {
	t.Url = fmt.Sprintf("https://twitter.com/i/web/status/%d", id)
}

func (t *TweetRef) GetUrl() string {
	return t.Url
}

type TwitterUser struct {
	Id               int        `json:"id"`
	Username         string     `json:"username"`
	DisplayName      string     `json:"display_name"`
	Description      string     `json:"description"`
	RawDescription   string     `json:"raw_description"`
	DescriptionLinks []string   `json:"description_links"`
	Verified         bool       `json:"verified"`
	Created          *time.Time `json:"created"`
	FollowersCount   int        `json:"followers_count"`
	FriendsCount     int        `json:"friends_count"`
	StatusesCount    int        `json:"statuses_count"`
	FavouritesCount  int        `json:"favourites_count"`
	ListedCount      int        `json:"listed_count"`
	MediaCount       int        `json:"media_count"`
	Location         string     `json:"location"`
	Protected        bool       `json:"protected"`
	ProfileImageURL  string     `json:"profile_image_url_https"`
	ProfileBannerURL string     `json:"profile_banner_url"`
	Label            struct {
		Url             string `json:"url"`
		Badge           string `json:"badge"`
		Description     string `json:"description"`
		LongDescription string `json:"long_description"`
	} `json:"label"`
	Url string `json:"url"`
}
