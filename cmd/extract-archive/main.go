package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const (
	zipPath  = `/Volumes/home/twitter-2025-09-02-31664232fed404b8a51511a2281218889ed09926e92ef278f526b21d708c4688.zip`
	outDir   = `data/tweets/`
	parallel = 2
)

var (
	fixedFiles = map[string]struct{}{
		"like.js":          {},
		"tweets.js":        {},
		"tweet-headers.js": {},
	}
	reTweetsPart = regexp.MustCompile(`^tweets-part(\d+)\.js$`)
)

func shouldPick(zipName string) bool {
	base := filepath.Base(zipName) // ディレクトリ無視
	if _, ok := fixedFiles[base]; ok {
		return true
	}
	return reTweetsPart.MatchString(base)
}

func extractSelected(ctx context.Context) error {
	f, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	// 中央ディレクトリを末尾から読み込む（ファイル全体は読まない）
	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		return err
	}

	// 抽出対象だけ拾う
	var targets []*zip.File
	for _, zf := range r.File {
		if shouldPick(zf.Name) {
			targets = append(targets, zf)
		}
	}
	if len(targets) == 0 {
		fmt.Println("no matches")
		return nil
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	sem := make(chan struct{}, parallel)
	errCh := make(chan error, len(targets))

	for _, zf := range targets {
		zf := zf
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			// contextキャンセル対応（任意）
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}
			if err := extractOne(zf); err != nil {
				errCh <- fmt.Errorf("%s: %w", zf.Name, err)
			} else {
				fmt.Println("extracted:", zf.Name)
			}
		}()
	}

	// 待ち合わせ
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	} // 全ゴルーチン終了待ち
	close(errCh)
	for e := range errCh {
		if e != nil {
			return e
		}
	}
	return nil
}

func extractOne(zf *zip.File) error {
	rc, err := zf.Open() // 必要なファイルだけストリーム読取
	if err != nil {
		return err
	}
	defer rc.Close()

	dst := filepath.Join(outDir, filepath.Base(zf.Name))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// 大きめバッファでNAS往復を減らす
	buf := make([]byte, 4<<20) // 4MB
	_, err = io.CopyBuffer(out, rc, buf)
	return err
}

func main() {
	if err := extractSelected(context.Background()); err != nil {
		fmt.Println("ERROR:", err)
	}
}
