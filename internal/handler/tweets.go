package handler

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"net/http"
	"sort"
	"strings"
	"time"
	"twilog-archive/internal/constant"
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
			return c.String(http.StatusBadRequest, "Request is failed: "+err.Error())
		}
		tweets, err := repository.Latest(db, req.LastID)
		if err != nil {
			return err
		}
		res, err := makeTweetResponse(db, tweets)
		if err != nil {
			return err
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
			return err
		}
		res, err := makeTweetResponse(db, tweets)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}

func TweetsSearch(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		/*
			req := new(SearchRequest)
			if err := c.Bind(req); err != nil {
				return c.String(http.StatusBadRequest, "Request is failed: "+err.Error())
			}
		*/
		q := c.QueryParam("q")
		req := &form.SearchRequest{
			SearchWord: q,
		}
		tweets, err := repository.Search(db, req)
		if err != nil {
			return err
		}
		res, err := makeTweetResponse(db, tweets)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
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
	grouped := make(map[string][]TweetResponseTweet)
	for _, t := range tweets {
		tweet := TweetResponseTweet{
			ID:         fmt.Sprint(t.ID),
			Text:       removeMediaURLFromEnd(t.FullText, t.EmbedMediaURL),
			ScreenName: t.ScreenName,
			Name: func() *string {
				if t.UserID == nil {
					name := constant.MyName
					return &name
				}
				return t.Name
			}(),
			Created:   t.CreatedAt.In(time.Local).Format("2006/01/02 15:04:05"),
			Retweeted: t.Retweeted,
			Replied:   t.Replied,
		}
		media, err := repository.MediaFindByTweetID(db, t.ID)
		if err != nil {
			return nil, err
		}
		if media != nil && len(media) > 0 {
			var mediaUrls []string
			for _, m := range media {
				mediaUrls = append(mediaUrls, m.MediaURL)
			}
			tweet.Media = mediaUrls
		}
		urls, err := repository.URLsFindByTweetID(db, t.ID)
		if err != nil {
			return nil, err
		}
		if urls != nil && len(urls) > 0 {
			var respUrls []TweetResponseURL
			for _, u := range urls {
				respUrls = append(respUrls, TweetResponseURL{
					URL:         u.URL,
					ExpandedURL: u.ExpandURL,
					DisplayURL:  u.DisplayURL,
				})
			}
			tweet.URLs = respUrls
		}
		hashtags, err := repository.HashtagsFindByTweetID(db, t.ID)
		if err != nil {
			return nil, err
		}
		if hashtags != nil && len(hashtags) > 0 {
			var respHashTags []string
			for _, h := range hashtags {
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
