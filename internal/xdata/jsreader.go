package xdata

import (
	"bufio"
	"io"
)

// StripJSPrefixReader JavaScriptの変数代入式 (例: window.YTD.tweets.part0 = [ ... ]) の先頭プレフィックスを読み飛ばし、
// JSON構造の開始文字 ('[' または '{') からの読み込みを提供する io.Reader です。
type StripJSPrefixReader struct {
	r           *bufio.Reader
	foundPrefix bool
}

func NewStripJSPrefixReader(r io.Reader) *StripJSPrefixReader {
	return &StripJSPrefixReader{
		r: bufio.NewReader(r),
	}
}

func (s *StripJSPrefixReader) Read(p []byte) (n int, err error) {
	if !s.foundPrefix {
		for {
			b, err := s.r.ReadByte()
			if err != nil {
				return 0, err
			}
			if b == '[' || b == '{' {
				if err := s.r.UnreadByte(); err != nil {
					return 0, err
				}
				s.foundPrefix = true
				break
			}
		}
	}
	return s.r.Read(p)
}
