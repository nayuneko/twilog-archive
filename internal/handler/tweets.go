package handler

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"net/http"
	"sort"
	"strings"
	"time"
	"twilog-archive/internal/config"
	"twilog-archive/internal/form"
	"twilog-archive/internal/model"
	"twilog-archive/internal/repository"
)

type (
	TweetResponse struct {
		Date   string               `json:"date"`
		Tweets []TweetResponseTweet `json:"tweets"`
	}
	TweetResponseTweet struct {
		ID         string             `json:"id"`
		Text       string             `json:"text"`
		ScreenName string             `json:"screen_name"`
		Name       *string            `json:"name"`
		Created    string             `json:"created"`
		Retweeted  bool               `json:"retweeted"`
		Replied    bool               `json:"replied"`
		Media      []string           `json:"media,omitempty"`
		URLs       []TweetResponseURL `json:"urls,omitempty"`
		Hashtags   []string           `json:"hashtags,omitempty"`
	}
	TweetResponseURL struct {
		URL         string `json:"url"`
		ExpandedURL string `json:"expanded_url"`
		DisplayURL  string `json:"display_url"`
	}
)

func TweetsLatest(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(form.Pagination)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request: " + err.Error()})
		}
		tweets, err := repository.Latest(db, req.LastID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch tweets: "+err.Error())
		}
		res, err := makeTweetResponse(db, tweets)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to format response: "+err.Error())
		}
		return c.JSON(http.StatusOK, res)
	}
}

// TweetsDates /dates/:date 日付に属する全てのツイートを取得
func TweetsDates(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		date := c.Param("date")
		tweets, err := repository.FindByDates(db, date)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch tweets by date: "+err.Error())
		}
		res, err := makeTweetResponse(db, tweets)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to format response: "+err.Error())
		}
		return c.JSON(http.StatusOK, res)
	}
}

func TweetsSearch(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := c.QueryParam("q")
		searchType := c.QueryParam("type")
		req := &form.SearchRequest{
			SearchWord: q,
			SearchType: searchType,
		}
		tweets, totalCount, err := repository.Search(db, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Search failed: "+err.Error())
		}
		res, err := makeTweetResponse(db, tweets)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to format response: "+err.Error())
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"total_count": totalCount,
			"tweets":      res,
		})
	}
}

func removeMediaURLFromEnd(text string, embedMediaURL *string) string {
	if embedMediaURL == nil {
		return text
	}
	if strings.HasSuffix(text, *embedMediaURL) {
		// 末尾のURLを削除
		return strings.TrimSuffix(text, *embedMediaURL)
	}
	return text
}

func makeTweetResponse(db *sqlx.DB, tweets []model.TweetsWithName) ([]TweetResponse, error) {
	if len(tweets) == 0 {
		return []TweetResponse{}, nil
	}

	tweetIDs := make([]int64, len(tweets))
	for i, t := range tweets {
		tweetIDs[i] = t.ID
	}

	// 一括取得で N+1 を解消
	mediaMap, err := repository.MediaFindByTweetIDs(db, tweetIDs)
	if err != nil {
		return nil, err
	}
	urlsMap, err := repository.URLsFindByTweetIDs(db, tweetIDs)
	if err != nil {
		return nil, err
	}
	hashtagsMap, err := repository.HashtagsFindByTweetIDs(db, tweetIDs)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]TweetResponseTweet)
	for _, t := range tweets {
		tweet := TweetResponseTweet{
			ID:         fmt.Sprint(t.ID),
			Text:       removeMediaURLFromEnd(t.FullText, t.EmbedMediaURL),
			ScreenName: t.ScreenName,
			Name: func() *string {
				if t.UserID != nil && t.Name != nil {
					return t.Name
				}
				if t.ScreenName == config.MyScreenName {
					name := config.MyName
					return &name
				}
				return nil
			}(),
			Created:   t.CreatedAt.In(time.Local).Format("2006/01/02 15:04:05"),
			Retweeted: t.Retweeted,
			Replied:   t.Replied,
		}

		if mediaList, ok := mediaMap[t.ID]; ok && len(mediaList) > 0 {
			var mediaUrls []string
			for _, m := range mediaList {
				mediaUrls = append(mediaUrls, m.MediaURL)
			}
			tweet.Media = mediaUrls
		}

		if urlList, ok := urlsMap[t.ID]; ok && len(urlList) > 0 {
			var respUrls []TweetResponseURL
			for _, u := range urlList {
				respUrls = append(respUrls, TweetResponseURL{
					URL:         u.URL,
					ExpandedURL: u.ExpandURL,
					DisplayURL:  u.DisplayURL,
				})
			}
			tweet.URLs = respUrls
		}

		if hashtagList, ok := hashtagsMap[t.ID]; ok && len(hashtagList) > 0 {
			var respHashTags []string
			for _, h := range hashtagList {
				respHashTags = append(respHashTags, h.Tag)
			}
			tweet.Hashtags = respHashTags
		}

		grouped[t.CreatedDate] = append(grouped[t.CreatedDate], tweet)
	}

	var response []TweetResponse
	for date, tweets := range grouped {
		response = append(response, TweetResponse{
			Date:   date,
			Tweets: tweets,
		})
	}
	sort.Slice(response, func(i, j int) bool {
		return response[i].Date > response[j].Date
	})

	return response, nil
}
