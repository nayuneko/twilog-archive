package xdata

import (
	"encoding/json"
	"testing"
	"time"
)

func Test_TweetHeader(t *testing.T) {
	t.Run("decode", func(t *testing.T) {
		j := `[
  {
    "tweet": {
      "tweet_id": "1951520633868918961",
      "user_id": "87211693",
      "created_at": "Sat Aug 02 05:49:11 +0000 2025"
    }
  },
  {
    "tweet": {
      "tweet_id": "1951510674624094706",
      "user_id": "87211693",
      "created_at": "Sat Aug 02 05:09:36 +0000 2025"
    }
  }
]`
		var given []TweetHeaderWrapper
		if err := json.Unmarshal([]byte(j), &given); err != nil {
			t.Fatal(err)
		}
		want := []TweetHeaderWrapper{
			{
				TweetHeader: TweetHeader{
					1951520633868918961,
					87211693,
					XCreatedAt(time.Date(2025, time.August, 2, 5, 49, 11, 0, time.UTC)),
				},
			},
			{
				TweetHeader: TweetHeader{
					1951510674624094706,
					87211693,
					XCreatedAt(time.Date(2025, time.August, 2, 5, 9, 36, 0, time.UTC)),
				},
			},
		}
		for idx, w := range want {
			g := given[idx]
			if int64(g.TweetHeader.TweetID) != int64(w.TweetHeader.TweetID) {
				t.Fatalf("TweetIDが違う. idx = %d, given = %d, want = %d", idx, g.TweetHeader.TweetID, w.TweetHeader.TweetID)
			}
			if int64(g.TweetHeader.UserID) != int64(w.TweetHeader.UserID) {
				t.Fatalf("UserIDが違う. idx = %d, given = %d, want = %d", idx, g.TweetHeader.TweetID, w.TweetHeader.TweetID)
			}
			if g.TweetHeader.CreatedAt.Format(time.RFC3339) != w.TweetHeader.CreatedAt.Format(time.RFC3339) {
				t.Fatalf("CreatedAtが違う. idx = %d, given = %s, want = %s", idx, g.TweetHeader.CreatedAt.Format(time.RFC3339), w.TweetHeader.CreatedAt.Format(time.RFC3339))
			}
		}
	})
}
