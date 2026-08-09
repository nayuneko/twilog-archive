package repository

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"twilog-archive/internal/form"
)

// trigramMinChars は FTS5 Trigram Tokenizer が検索可能な最小文字数 (Unicode rune 単位)
const trigramMinChars = 3

// BuildTypeFilterSQL は TweetTypeFilter から SQL WHERE 条件節を生成する
func BuildTypeFilterSQL(filter form.TweetTypeFilter) string {
	// 全て true または全て false（ゼロ値）の場合は絞り込みなし
	if (filter.IncludeNormal && filter.IncludeReply && filter.IncludeRT) ||
		(!filter.IncludeNormal && !filter.IncludeReply && !filter.IncludeRT) {
		return ""
	}

	var conds []string
	if filter.IncludeNormal {
		conds = append(conds, "(t.retweeted = 0 AND t.replied = 0)")
	}
	if filter.IncludeReply {
		conds = append(conds, "t.replied = 1")
	}
	if filter.IncludeRT {
		conds = append(conds, "t.retweeted = 1")
	}

	if len(conds) == 0 {
		return "1 = 0"
	}

	return "(" + strings.Join(conds, " OR ") + ")"
}

type parsedToken struct {
	text    string
	isQuote bool
	isNot   bool
	isOp    bool
}

// BuildFTSQuery は入力検索ワード、除外ワード、searchType ("and"|"or") から FTS5 MATCH 式を生成する
func BuildFTSQuery(input string, excludeInput string, searchType string) string {
	incTokens, excTokens := parseCombinedInput(input, excludeInput)
	if len(incTokens) == 0 {
		return ""
	}

	defaultOp := "AND"
	if strings.ToLower(searchType) == "or" {
		defaultOp = "OR"
	}

	var incParts []string
	var currentOp string

	for _, t := range incTokens {
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
		if len(incParts) == 0 {
			incParts = append(incParts, expr)
		} else {
			incParts = append(incParts, op, expr)
		}
	}

	if len(incParts) == 0 {
		return ""
	}

	var mainExpr string
	if len(incParts) > 1 {
		mainExpr = "(" + strings.Join(incParts, " ") + ")"
	} else {
		mainExpr = incParts[0]
	}

	if len(excTokens) == 0 {
		return mainExpr
	}

	var excParts []string
	for _, t := range excTokens {
		t.isNot = true
		excParts = append(excParts, formatFTSToken(t))
	}

	return mainExpr + " " + strings.Join(excParts, " ")
}

// BuildHashtagExcludeSQL は除外指定されたハッシュタグ（#タグ）についてのサブクエリ WHERE 条件を生成する
func BuildHashtagExcludeSQL(input string, excludeInput string) (string, []interface{}) {
	_, excTokens := parseCombinedInput(input, excludeInput)
	var tags []string

	for _, t := range excTokens {
		clean := strings.TrimSpace(t.text)
		if strings.HasPrefix(clean, "#") || strings.HasPrefix(clean, "＃") {
			rawTag := strings.TrimPrefix(strings.TrimPrefix(clean, "#"), "＃")
			if rawTag != "" {
				tags = append(tags, rawTag, "#"+rawTag, "＃"+rawTag)
			}
		}
	}

	if len(tags) == 0 {
		return "", nil
	}

	uniqueMap := make(map[string]bool)
	var placeholders []string
	var params []interface{}

	for _, tag := range tags {
		if !uniqueMap[tag] {
			uniqueMap[tag] = true
			placeholders = append(placeholders, "?")
			params = append(params, tag)
		}
	}

	sqlCond := fmt.Sprintf("t.id NOT IN (SELECT tweet_id FROM hashtags WHERE tag IN (%s))", strings.Join(placeholders, ", "))
	return sqlCond, params
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
	// 2: マイナスまたはハッシュで始まる単語
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
func BuildLikeQuery(input string, excludeInput string, searchType string) (string, []interface{}) {
	incTokens, excTokens := parseCombinedInput(input, excludeInput)
	if len(incTokens) == 0 {
		return "", nil
	}

	defaultOp := "AND"
	if strings.ToLower(searchType) == "or" {
		defaultOp = "OR"
	}

	var incConditions []string
	var params []interface{}
	var currentOp string

	for _, t := range incTokens {
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
		param := "%" + t.text + "%"

		if len(incConditions) == 0 {
			incConditions = append(incConditions, cond)
		} else {
			incConditions = append(incConditions, op+" "+cond)
		}
		params = append(params, param)
	}

	if len(incConditions) == 0 {
		return "", nil
	}

	var mainCond string
	if len(incConditions) > 1 {
		mainCond = "(" + strings.Join(incConditions, " ") + ")"
	} else {
		mainCond = incConditions[0]
	}

	if len(excTokens) == 0 {
		return mainCond, params
	}

	var excConditions []string
	for _, t := range excTokens {
		excConditions = append(excConditions, "t.full_text NOT LIKE ?")
		params = append(params, "%"+t.text+"%")
	}

	fullCond := mainCond + " AND " + strings.Join(excConditions, " AND ")
	return "(" + fullCond + ")", params
}

func parseCombinedInput(input string, excludeInput string) ([]parsedToken, []parsedToken) {
	tokens := parseSearchInput(input)
	excTokensFromInput := parseSearchInput(excludeInput)

	var incTokens []parsedToken
	var excTokens []parsedToken

	for _, t := range tokens {
		if t.isNot {
			excTokens = append(excTokens, t)
		} else {
			incTokens = append(incTokens, t)
		}
	}

	for _, t := range excTokensFromInput {
		if !t.isOp {
			t.isNot = true
			excTokens = append(excTokens, t)
		}
	}

	return incTokens, excTokens
}

// isFTSCompatible はトークンが FTS5 Trigram Tokenizer で検索可能かどうかを返す。
// Trigram Tokenizer は最低3文字（Unicode rune 単位）必要。
func isFTSCompatible(text string) bool {
	return utf8.RuneCountInString(text) >= trigramMinChars
}

// HybridSearchResult は FTS5 と LIKE を組み合わせたハイブリッド検索の構成要素を保持する
type HybridSearchResult struct {
	// FTSQuery は FTS5 MATCH 式（3文字以上のトークンのみ）。空文字の場合は FTS5 不使用。
	FTSQuery string
	// LikeConds は 2文字以下トークン用の SQL LIKE 条件式の配列 (例: "t.full_text LIKE ?")
	LikeConds []string
	// LikeParams は LikeConds に対応するパラメータ
	LikeParams []interface{}
	// HasFTS は FTS5 を使用する検索トークンが1つ以上あるかどうか
	HasFTS bool
}

// BuildHybridSearchQuery は入力を解析し、FTS5 互換トークン（3文字以上）は MATCH 式に、
// 2文字以下のトークンは LIKE 条件に分離した HybridSearchResult を返す。
func BuildHybridSearchQuery(input string, excludeInput string, searchType string) HybridSearchResult {
	incTokens, excTokens := parseCombinedInput(input, excludeInput)
	if len(incTokens) == 0 {
		return HybridSearchResult{}
	}

	defaultOp := "AND"
	if strings.ToLower(searchType) == "or" {
		defaultOp = "OR"
	}

	var ftsIncParts []string
	var likeIncConds []string
	var likeParams []interface{}
	var currentOp string

	for _, t := range incTokens {
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

		if isFTSCompatible(t.text) {
			// 3文字以上: FTS5 MATCH 式に追加
			expr := formatFTSToken(t)
			if len(ftsIncParts) == 0 {
				ftsIncParts = append(ftsIncParts, expr)
			} else {
				ftsIncParts = append(ftsIncParts, op, expr)
			}
		} else {
			// 2文字以下: LIKE 条件に追加
			if op == "OR" && len(likeIncConds) > 0 {
				likeIncConds = append(likeIncConds, "OR t.full_text LIKE ?")
			} else {
				likeIncConds = append(likeIncConds, "t.full_text LIKE ?")
			}
			likeParams = append(likeParams, "%"+t.text+"%")
		}
	}

	// FTS5 側の MATCH 式を組み立て
	var ftsQuery string
	if len(ftsIncParts) > 0 {
		if len(ftsIncParts) > 1 {
			ftsQuery = "(" + strings.Join(ftsIncParts, " ") + ")"
		} else {
			ftsQuery = ftsIncParts[0]
		}

		// 除外トークン（3文字以上）は FTS MATCH 式に NOT で追加
		var ftsExcParts []string
		var likeExcConds []string
		for _, t := range excTokens {
			t.isNot = true
			if isFTSCompatible(t.text) {
				ftsExcParts = append(ftsExcParts, formatFTSToken(t))
			} else {
				likeExcConds = append(likeExcConds, "t.full_text NOT LIKE ?")
				likeParams = append(likeParams, "%"+t.text+"%")
			}
		}
		if len(ftsExcParts) > 0 {
			ftsQuery += " " + strings.Join(ftsExcParts, " ")
		}
		likeIncConds = append(likeIncConds, likeExcConds...)
	} else {
		// FTS5 対象トークンが無い場合、除外は全て LIKE で処理
		for _, t := range excTokens {
			likeIncConds = append(likeIncConds, "t.full_text NOT LIKE ?")
			likeParams = append(likeParams, "%"+t.text+"%")
		}
	}

	return HybridSearchResult{
		FTSQuery:   ftsQuery,
		LikeConds:  likeIncConds,
		LikeParams: likeParams,
		HasFTS:     len(ftsIncParts) > 0,
	}
}

