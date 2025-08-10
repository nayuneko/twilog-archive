package xdata

import (
	"encoding/json"
	"strconv"
	"time"
)

type (
	TweetWrapper struct {
		Tweet Tweet `json:"tweet"`
	}

	Tweet struct {
		ID              XID        `json:"id"`
		Source          string     `json:"source"`
		CreatedAt       XCreatedAt `json:"created_at"`
		FullText        string     `json:"full_text"`
		Lang            string     `json:"lang"`
		InReplyToUserID *string    `json:"in_reply_to_user_id,omitempty"`
		Entities        struct {
			Hashtags     []Entity  `json:"hashtags"`
			UserMentions []Mention `json:"user_mentions"`
			URLs         []URL     `json:"urls"`
			Media        []Media   `json:"media"`
		} `json:"entities"`
		EditInfo struct {
			Initial struct {
				EditTweetIds   []string  `json:"editTweetIds"`
				EditableUntil  time.Time `json:"editableUntil"`
				EditsRemaining string    `json:"editsRemaining"`
				IsEditEligible bool      `json:"isEditEligible"`
			} `json:"initial"`
		} `json:"edit_info"`
		ExtendedEntities struct {
			Media []Media `json:"media"`
		} `json:"extended_entities,omitempty"`
	}

	Media struct {
		ID             XID    `json:"id"`
		MediaURL       string `json:"media_url"`
		MediaURLHttps  string `json:"media_url_https"`
		URL            string `json:"url"`
		DisplayURL     string `json:"display_url"`
		ExpandedURL    string `json:"expanded_url"`
		Type           string `json:"type"` // "photo", "video", "animated_gif"
		Indices        []Sint `json:"indices"`
		SourceStatusID string `json:"source_status_id_str,omitempty"`
		SourceUserID   string `json:"source_user_id_str,omitempty"`
	}

	Entity struct {
		Text    string `json:"text"`
		Indices []Sint `json:"indices"`
	}

	Mention struct {
		ID         XID    `json:"id"`
		ScreenName string `json:"screen_name"`
		Name       string `json:"name"`
		Indices    []Sint `json:"indices"`
	}

	URL struct {
		URL         string `json:"url"`
		ExpandedURL string `json:"expanded_url"`
		DisplayURL  string `json:"display_url"`
		Indices     []Sint `json:"indices"`
	}
)

// XID XのIDは基本的にstringでくるが、int64で保持する
type XID int64

func (xid *XID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*xid = XID(n)
		return nil
	}

	// fallback: number directly
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*xid = XID(n)
	return nil
}

type Sint int

func (si *Sint) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		n, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*si = Sint(n)
		return nil
	}

	// fallback: number directly
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*si = Sint(n)
	return nil
}

type XCreatedAt time.Time

func (tt *XCreatedAt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// RubyDate形式でパース
	t, err := time.Parse(time.RubyDate, s)
	if err != nil {
		return err
	}
	*tt = XCreatedAt(t)
	return nil
}

func (tt *XCreatedAt) Time() time.Time {
	return time.Time(*tt)
}
