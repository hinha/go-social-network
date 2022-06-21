package sns

import (
	"time"

	"github.com/hinha/go-social-network/entities"
)

type URLParams map[string]string

type (
	// TweetResult of scrapping.
	TweetResult struct {
		*entities.TwitterPost
		Error error
	}

	twitterResponse struct {
		GlobalObjects struct {
			Tweets map[string]TweetRaw   `json:"tweets"`
			Users  map[string]TweetUsers `json:"users"`
		} `json:"globalObjects"`
		Timeline struct {
			Id           string              `json:"id"`
			Instructions []TweetInstructions `json:"instructions"`
		} `json:"timeline"`
		Data struct {
			User *struct {
				Result struct {
					Timeline struct {
						Timeline struct {
							Instructions []TweetInstructions `json:"instructions"`
						} `json:"timeline"`
					} `json:"timeline"`
				} `json:"result"`
			} `json:"user"`
			ThreadedConversationWithInjections struct {
				Instructions []TweetInstructions `json:"instructions"`
			} `json:"threaded_conversation_with_injections"`
		} `json:"data"`
	}
)

type TweetInstructions struct {
	AddEntries   map[string]interface{} `json:"addEntries"`
	ReplaceEntry map[string]interface{} `json:"replaceEntry"`
	// Graphql
	Type    string        `json:"type"`
	Entries []interface{} `json:"entries,omitempty"`
}

type TweetRaw struct {
	Id                int    `json:"id"`
	IdStr             string `json:"id_str"`
	ConversationID    int    `json:"conversation_id"`
	ConversationIDStr string `json:"conversation_id_str"`
	CreatedAt         string `json:"created_at"`
	FavoriteCount     int    `json:"favorite_count"`
	FullText          string `json:"full_text"`
	Entities          struct {
		Mentions []struct {
			Text string `json:"text"`
		} `json:"mentions"`
		Hashtags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		Symbols []struct {
			Text string `json:"text"`
		} `json:"symbols"`
		Media []struct {
			MediaURLHttps string `json:"media_url_https"`
			Type          string `json:"type"`
			URL           string `json:"url"`
		} `json:"media"`
		URLs         []TweetUrls  `json:"urls"`
		UserMentions []TweetUsers `json:"user_mentions"`
	} `json:"entities"`
	ExtendedEntities struct {
		Media []struct {
			IDStr         string `json:"id_str"`
			MediaURLHttps string `json:"media_url_https"`
			Type          string `json:"type"`
			VideoInfo     struct {
				DurationMillis int `json:"duration_millis"`
				Variants       []struct {
					ContentType string `json:"content_type"`
					Bitrate     int    `json:"bitrate"`
					URL         string `json:"url"`
				} `json:"variants"`
			} `json:"video_info"`
			Ext struct {
				MediaStats map[string]interface{} `json:"mediaStats"`
			} `json:"ext"`
		} `json:"media"`
	} `json:"extended_entities"`
	InReplyToStatusIDStr string        `json:"in_reply_to_status_id_str"`
	ReplyCount           int           `json:"reply_count"`
	RetweetCount         int           `json:"retweet_count"`
	QuoteCount           int           `json:"quote_count"`
	Lang                 string        `json:"lang"`
	Source               string        `json:"source"`
	RetweetedStatusIDStr string        `json:"retweeted_status_id_str"`
	QuotedStatusIDStr    string        `json:"quoted_status_id_str"`
	InReplyToUserIdStr   string        `json:"in_reply_to_user_id_str"`
	InReplyToScreenName  string        `json:"in_reply_to_screen_name"`
	Time                 time.Time     `json:"time"`
	UserIDStr            string        `json:"user_id_str"`
	Card                 *TweetRawCard `json:"card,omitempty"`
	Coordinates          *struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"coordinates"`
	Geo *struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geo"`
	Place *struct {
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		PlaceType   string `json:"place_type"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		BoundingBox *struct {
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"bounding_box"`
	} `json:"place"`
}

type TweetRawCard struct {
	Name          string                 `json:"name,omitempty"`
	BindingValues map[string]interface{} `json:"binding_values,omitempty"`
	Users         map[string]TweetUsers  `json:"users,omitempty"`
	CardPlatform  map[string]interface{} `json:"card_platform,omitempty"`
	Legacy        struct {
		BindingValues []interface{} `json:"binding_values"`
		CardPlatform  struct {
			Platform struct {
				Audience struct {
					Name string `json:"name"`
				} `json:"audience"`
				Device struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"device"`
			} `json:"platform"`
		} `json:"card_platform"`
		Name     string `json:"name"`
		URL      string `json:"url"`
		UserRefs []struct {
			AffiliatesHighlightedLabel struct {
			} `json:"affiliates_highlighted_label"`
			HasNftAvatar        bool        `json:"has_nft_avatar"`
			ID                  string      `json:"id"`
			Legacy              *TweetUsers `json:"legacy"`
			RestID              string      `json:"rest_id"`
			SuperFollowEligible bool        `json:"super_follow_eligible"`
			SuperFollowedBy     bool        `json:"super_followed_by"`
			SuperFollowing      bool        `json:"super_following"`
		} `json:"user_refs"`
	} `json:"legacy,omitempty"`
	RestID string `json:"rest_id,omitempty"`
}

type TweetUsers struct {
	Id          int    `json:"id"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
	Entities    struct {
		URL struct {
			Urls []TweetUrls `json:"urls"`
		} `json:"url"`
	} `json:"entities"`
	FavouritesCount      int      `json:"favourites_count"`
	FollowersCount       int      `json:"followers_count"`
	FriendsCount         int      `json:"friends_count"`
	MediaCount           int      `json:"media_count"`
	IDStr                string   `json:"id_str"`
	ListedCount          int      `json:"listed_count"`
	Name                 string   `json:"name"`
	Location             string   `json:"location"`
	PinnedTweetIdsStr    []string `json:"pinned_tweet_ids_str"`
	ProfileBannerURL     string   `json:"profile_banner_url"`
	ProfileImageURLHTTPS string   `json:"profile_image_url_https"`
	Protected            bool     `json:"protected"`
	ScreenName           string   `json:"screen_name"`
	StatusesCount        int      `json:"statuses_count"`
	Verified             bool     `json:"verified"`
	Url                  string   `json:"url"`
	Ext                  struct {
		HighlightedLabel struct {
			R struct {
				Ok struct {
					Label *struct {
						Description     string                 `json:"description"`
						Url             map[string]interface{} `json:"url"`
						Badge           map[string]interface{} `json:"badge"`
						LongDescription map[string]interface{} `json:"long_description"`
					} `json:"label"`
				} `json:"ok"`
			} `json:"r"`
		} `json:"highlightedLabel"`
	} `json:"ext"`
	WithheldInCountries []interface{} `json:"withheld_in_countries"`
	TranslatorType      string        `json:"translator_type"`
}

type TweetUrls struct {
	Url         string `json:"url"`
	DisplayUrl  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
	Indices     []int  `json:"indices"`
}

type TweetGraphqlUser struct {
	Data struct {
		User struct {
			Result struct {
				Typename              string `json:"__typename"`
				HasNftAvatar          bool   `json:"has_nft_avatar"`
				ID                    string `json:"id"`
				IsProfileTranslatable bool   `json:"is_profile_translatable"`
				Legacy                struct {
					CreatedAt           string `json:"created_at"`
					DefaultProfile      bool   `json:"default_profile"`
					DefaultProfileImage bool   `json:"default_profile_image"`
					Description         string `json:"description"`
					Entities            struct {
						Description struct {
							Urls []TweetUrls `json:"urls"`
						} `json:"description"`
					} `json:"entities"`
					FastFollowersCount      int           `json:"fast_followers_count"`
					FavouritesCount         int           `json:"favourites_count"`
					FollowersCount          int           `json:"followers_count"`
					FriendsCount            int           `json:"friends_count"`
					HasCustomTimelines      bool          `json:"has_custom_timelines"`
					IsTranslator            bool          `json:"is_translator"`
					ListedCount             int           `json:"listed_count"`
					Location                string        `json:"location"`
					MediaCount              int           `json:"media_count"`
					Name                    string        `json:"name"`
					NormalFollowersCount    int           `json:"normal_followers_count"`
					PinnedTweetIdsStr       []string      `json:"pinned_tweet_ids_str"`
					ProfileBannerURL        string        `json:"profile_banner_url"`
					ProfileImageURLHTTPS    string        `json:"profile_image_url_https"`
					ProfileInterstitialType string        `json:"profile_interstitial_type"`
					Protected               bool          `json:"protected"`
					ScreenName              string        `json:"screen_name"`
					StatusesCount           int           `json:"statuses_count"`
					TranslatorType          string        `json:"translator_type"`
					Verified                bool          `json:"verified"`
					WithheldInCountries     []interface{} `json:"withheld_in_countries"`
				} `json:"legacy"`
				LegacyExtendedProfile struct {
				} `json:"legacy_extended_profile"`
				RestID              string `json:"rest_id"`
				SuperFollowEligible bool   `json:"super_follow_eligible"`
				SuperFollowedBy     bool   `json:"super_followed_by"`
				SuperFollowing      bool   `json:"super_following"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
}

type TweetGraphqlEntries struct {
	Content struct {
		EntryType   string `json:"entryType"`
		ItemContent struct {
			ItemType         string `json:"itemType"`
			RuxContext       string `json:"ruxContext"`
			TweetDisplayType string `json:"tweetDisplayType"`
			TweetResults     struct {
				Result struct {
					Tweet              interface{} `json:"tweet"`
					Typename           string      `json:"__typename"`
					QuotedStatusResult interface{} `json:"quoted_status_result"` // null
					QuotedRefResult    interface{} `json:"quotedRefResult"`      // null
					Card               interface{} `json:"card"`                 // null
					Core               struct {
						UserResults struct {
							Result struct {
								Typename                   string `json:"__typename"`
								AffiliatesHighlightedLabel struct {
									Label struct {
										Badge struct {
											URL string `json:"url"`
										} `json:"badge"`
										Description string `json:"description"`
										URL         struct {
											URL     string `json:"url"`
											URLType string `json:"urlType"`
										} `json:"url"`
									} `json:"label"`
								} `json:"affiliates_highlighted_label"`
								HasNftAvatar        bool       `json:"has_nft_avatar"`
								ID                  string     `json:"id"`
								Legacy              TweetUsers `json:"legacy"`
								RestID              string     `json:"rest_id"`
								SuperFollowEligible bool       `json:"super_follow_eligible"`
								SuperFollowedBy     bool       `json:"super_followed_by"`
								SuperFollowing      bool       `json:"super_following"`
							} `json:"result"`
						} `json:"user_results"`
					} `json:"core"`
					Legacy struct {
						ConversationIDStr string `json:"conversation_id_str"`
						CreatedAt         string `json:"created_at"`
						DisplayTextRange  []int  `json:"display_text_range"`
						Entities          struct {
							Hashtags []interface{} `json:"hashtags"`
							Media    []struct {
								DisplayURL  string `json:"display_url"`
								ExpandedURL string `json:"expanded_url"`
								Features    struct {
								} `json:"features"`
								IDStr         string `json:"id_str"`
								Indices       []int  `json:"indices"`
								MediaURLHTTPS string `json:"media_url_https"`
								OriginalInfo  struct {
									Height int `json:"height"`
									Width  int `json:"width"`
								} `json:"original_info"`
								Sizes struct {
									Large struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"large"`
									Medium struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"medium"`
									Small struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"small"`
									Thumb struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"thumb"`
								} `json:"sizes"`
								Type string `json:"type"`
								URL  string `json:"url"`
							} `json:"media"`
							Symbols      []interface{} `json:"symbols"`
							Urls         []interface{} `json:"urls"`
							UserMentions []interface{} `json:"user_mentions"`
						} `json:"entities"`
						ExtendedEntities struct {
							Media []struct {
								AdditionalMediaInfo struct {
									Monetizable bool `json:"monetizable"`
								} `json:"additional_media_info"`
								DisplayURL           string `json:"display_url"`
								ExpandedURL          string `json:"expanded_url"`
								ExtMediaAvailability struct {
									Status string `json:"status"`
								} `json:"ext_media_availability"`
								ExtMediaColor struct {
									Palette []struct {
										Percentage float64 `json:"percentage"`
										Rgb        struct {
											Blue  int `json:"blue"`
											Green int `json:"green"`
											Red   int `json:"red"`
										} `json:"rgb"`
									} `json:"palette"`
								} `json:"ext_media_color"`
								Features struct {
								} `json:"features"`
								IDStr      string `json:"id_str"`
								Indices    []int  `json:"indices"`
								MediaStats struct {
									ViewCount int `json:"viewCount"`
								} `json:"mediaStats"`
								MediaKey      string `json:"media_key"`
								MediaURLHTTPS string `json:"media_url_https"`
								OriginalInfo  struct {
									Height int `json:"height"`
									Width  int `json:"width"`
								} `json:"original_info"`
								Sizes struct {
									Large struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"large"`
									Medium struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"medium"`
									Small struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"small"`
									Thumb struct {
										H      int    `json:"h"`
										Resize string `json:"resize"`
										W      int    `json:"w"`
									} `json:"thumb"`
								} `json:"sizes"`
								Type      string `json:"type"`
								URL       string `json:"url"`
								VideoInfo struct {
									AspectRatio    []int `json:"aspect_ratio"`
									DurationMillis int   `json:"duration_millis"`
									Variants       []struct {
										Bitrate     int    `json:"bitrate,omitempty"`
										ContentType string `json:"content_type"`
										URL         string `json:"url"`
									} `json:"variants"`
								} `json:"video_info"`
							} `json:"media"`
						} `json:"extended_entities"`
						FavoriteCount             int         `json:"favorite_count"`
						Favorited                 bool        `json:"favorited"`
						FullText                  string      `json:"full_text"`
						IDStr                     string      `json:"id_str"`
						IsQuoteStatus             bool        `json:"is_quote_status"`
						Lang                      string      `json:"lang"`
						PossiblySensitive         bool        `json:"possibly_sensitive"`
						PossiblySensitiveEditable bool        `json:"possibly_sensitive_editable"`
						QuoteCount                int         `json:"quote_count"`
						ReplyCount                int         `json:"reply_count"`
						RetweetCount              int         `json:"retweet_count"`
						Retweeted                 bool        `json:"retweeted"`
						Source                    string      `json:"source"`
						UserIDStr                 string      `json:"user_id_str"`
						RetweetedStatusResult     interface{} `json:"retweeted_status_result"` // null
						QuotedStatusIdStr         string      `json:"quoted_status_id_str"`
					} `json:"legacy"`
					RestID string `json:"rest_id"`
				} `json:"result"`
			} `json:"tweet_results"`
		} `json:"itemContent"`
	} `json:"content"`
	EntryID   string `json:"entryId"`
	SortIndex string `json:"sortIndex"`
}
