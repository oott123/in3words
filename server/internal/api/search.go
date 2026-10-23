package api

import (
	"database/sql"
	"strings"
	"unicode/utf8"

	"github.com/dlclark/regexp2"

	"github.com/oott123/in3words/server/internal/wpdb"
)

// maxSearchTerms 与 WP_Query::parse_search 一致：超过 9 个词时把整句当作一个词。
const maxSearchTerms = 9

// searchTermRe 是 WP_Query::parse_search 中的分词正则，带后行断言，所以用 regexp2。
//
//nolint:gochecknoglobals // 预编译的正则
var searchTermRe = regexp2.MustCompile(`".*?("|$)|((?<=[\t ",+])|^)[^\t ",+]+`, regexp2.None)

// isSearchStopword 对应 WP_Query::get_search_stopwords() 的默认列表（zh_CN 未翻译该列表）。
func isSearchStopword(word string) bool {
	switch word {
	case "about", "an", "are", "as", "at", "be", "by", "com", "for", "from", "how", "in", "is", "it", "of",
		"on", "or", "that", "the", "this", "to", "was", "what", "when", "where", "who", "will", "with", "www":
		return true
	}
	return false
}

type searchTerm struct {
	like    string
	exclude bool
}

// parseSearch 对应 WP_Query::parse_search 与 parse_search_terms，返回每个词对应的 LIKE 模式。
func parseSearch(s string) []searchTerm {
	s = strings.NewReplacer("\r", "", "\n", "").Replace(s)
	if s == "" {
		return nil
	}

	var raw []string
	m, _ := searchTermRe.FindStringMatch(s)
	for m != nil {
		raw = append(raw, m.String())
		m, _ = searchTermRe.FindNextMatch(m)
	}

	words := []string{s}
	if len(raw) > 0 {
		checked := checkSearchTerms(raw)
		if len(checked) > 0 && len(checked) <= maxSearchTerms {
			words = checked
		}
	}

	terms := make([]searchTerm, 0, len(words))
	for _, w := range words {
		t := searchTerm{}
		if strings.HasPrefix(w, "-") {
			t.exclude = true
			w = w[1:]
		}
		t.like = "%" + escapeLike(w) + "%"
		terms = append(terms, t)
	}
	return terms
}

func checkSearchTerms(raw []string) []string {
	checked := make([]string, 0, len(raw))
	for _, term := range raw {
		// 精确匹配（整体带引号）的词保留首尾空格。
		if utf8.RuneCountInString(term) > 2 && strings.HasPrefix(term, `"`) && strings.HasSuffix(term, `"`) {
			term = strings.Trim(term, `"'`)
		} else {
			term = strings.Trim(term, `"' `)
		}
		// 跳过单个字母和单个短横线。
		if term == "" || (len(term) == 1 && isASCIILetterOrDash(term[0])) {
			continue
		}
		if isSearchStopword(strings.ToLower(term)) {
			continue
		}
		checked = append(checked, term)
	}
	return checked
}

func isASCIILetterOrDash(c byte) bool {
	return c == '-' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// escapeLike 对应 wpdb::esc_like()。
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `_`, `\_`, `%`, `\%`).Replace(s)
}

// applySearch 把搜索词依次填入查询的 s1..s9（包含）与 x1..x9（排除）参数。
func applySearch(p *wpdb.ListPostsParams, terms []searchTerm) {
	includes := []*sql.NullString{&p.S1, &p.S2, &p.S3, &p.S4, &p.S5, &p.S6, &p.S7, &p.S8, &p.S9}
	excludes := []*sql.NullString{&p.X1, &p.X2, &p.X3, &p.X4, &p.X5, &p.X6, &p.X7, &p.X8, &p.X9}
	ni, nx := 0, 0
	for _, t := range terms {
		if t.exclude {
			*excludes[nx] = sql.NullString{String: t.like, Valid: true}
			nx++
		} else {
			*includes[ni] = sql.NullString{String: t.like, Valid: true}
			ni++
		}
	}
}
