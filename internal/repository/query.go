package repository

import (
	"regexp"
	"strings"
)

type parsedToken struct {
	text    string
	isQuote bool
	isNot   bool
	isOp    bool
}

// BuildFTSQuery は入力検索ワードと searchType ("and"|"or") から FTS5 MATCH 式を生成する
func BuildFTSQuery(input string, searchType string) string {
	tokens := parseSearchInput(input)
	if len(tokens) == 0 {
		return ""
	}

	defaultOp := "AND"
	if strings.ToLower(searchType) == "or" {
		defaultOp = "OR"
	}

	var parts []string
	var currentOp string

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if t.isOp {
			if strings.ToUpper(t.text) == "OR" || t.text == "|" {
				currentOp = "OR"
			}
			continue
		}

		op := defaultOp
		if currentOp != "" {
			op = currentOp
			currentOp = ""
		}

		expr := formatFTSToken(t)
		if len(parts) == 0 {
			parts = append(parts, expr)
		} else {
			parts = append(parts, op, expr)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// formatFTSToken はトークンを FTS5 構文用にフォーマットする
func formatFTSToken(t parsedToken) string {
	escaped := strings.ReplaceAll(t.text, "\"", "\"\"")
	term := "\"" + escaped + "\""
	if t.isNot {
		return "NOT " + term
	}
	return term
}

// parseSearchInput は検索文字列をトークン分割する
func parseSearchInput(input string) []parsedToken {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// 全角スペースを半角スペースに正規化
	input = strings.ReplaceAll(input, "　", " ")

	var tokens []parsedToken
	// 1: ダブルクォートで囲まれたフレーズ (例: "hello world" または -"hello world")
	// 2: マイナスで始まる単語 (例: -python)
	// 3: 通常の単語
	re := regexp.MustCompile(`(-?"[^"]+")|(-[^\s]+)|([^\s]+)`)
	matches := re.FindAllString(input, -1)

	for i := 0; i < len(matches); i++ {
		match := matches[i]

		if match == "OR" || match == "or" || match == "|" {
			tokens = append(tokens, parsedToken{text: match, isOp: true})
			continue
		}
		if (match == "NOT" || match == "not") && i+1 < len(matches) {
			// 次のトークンを NOT 扱いにする
			nextToken := matches[i+1]
			i++
			tokens = append(tokens, parseSingleToken(nextToken, true))
			continue
		}

		tokens = append(tokens, parseSingleToken(match, false))
	}

	return tokens
}

func parseSingleToken(match string, forceNot bool) parsedToken {
	isNot := forceNot
	raw := match

	if !isNot && strings.HasPrefix(raw, "-") && len(raw) > 1 {
		isNot = true
		raw = raw[1:]
	}

	isQuote := false
	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") && len(raw) >= 2 {
		isQuote = true
		raw = raw[1 : len(raw)-1]
	}

	raw = strings.TrimSpace(raw)
	return parsedToken{
		text:    raw,
		isQuote: isQuote,
		isNot:   isNot,
	}
}

// BuildLikeQuery は FTS5 が使えない場合のフォールバック SQL 条件節を生成する
func BuildLikeQuery(input string, searchType string) (string, []interface{}) {
	tokens := parseSearchInput(input)
	if len(tokens) == 0 {
		return "", nil
	}

	defaultOp := "AND"
	if strings.ToLower(searchType) == "or" {
		defaultOp = "OR"
	}

	var conditions []string
	var params []interface{}
	var currentOp string

	for _, t := range tokens {
		if t.isOp {
			if strings.ToUpper(t.text) == "OR" || t.text == "|" {
				currentOp = "OR"
			}
			continue
		}

		op := defaultOp
		if currentOp != "" {
			op = currentOp
			currentOp = ""
		}

		cond := "t.full_text LIKE ?"
		if t.isNot {
			cond = "t.full_text NOT LIKE ?"
		}
		param := "%" + t.text + "%"

		if len(conditions) == 0 {
			conditions = append(conditions, cond)
		} else {
			conditions = append(conditions, op+" "+cond)
		}
		params = append(params, param)
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return "(" + strings.Join(conditions, " ") + ")", params
}
