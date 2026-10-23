package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oott123/in3words/server/internal/wpfmt"
)

const (
	secondsPerHour   = 3600
	secondsPerMinute = 60
	// maxInitials 是 initials 默认头像在名字不含空格或含 CJK 字符时截取的字数。
	maxInitials = 2
)

// siteOptionNames 返回每次请求需要读取的 wp_options 项。
func siteOptionNames() []string {
	return []string{
		"blogname",
		"blogdescription",
		"siteurl",
		"home",
		"gmt_offset",
		"timezone_string",
		"show_avatars",
		"avatar_default",
		"avatar_rating",
		"use_smilies",
		"https_migration_required",
		"sticky_posts",
		"WPLANG",
		"permalink_structure",
		"default_category",
		"posts_per_rss",
		"close_comments_for_old_posts",
		"close_comments_days_old",
		"upload_path",
		"upload_url_path",
	}
}

// site 是一次请求内用到的站点设置。
type site struct {
	options map[string]string
	tz      *time.Location
	fmt     wpfmt.Options
}

func (s *Server) loadSite(ctx context.Context) (*site, error) {
	rows, err := s.q.ListOptions(ctx, siteOptionNames())
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	opts := make(map[string]string, len(rows))
	for _, row := range rows {
		opts[row.OptionName] = row.OptionValue
	}

	st := &site{options: opts}
	st.tz = st.timezone()
	st.fmt = s.fmtOptions
	st.fmt.HomeURL = st.option("home", "")
	st.fmt.SiteURL = st.option("siteurl", "")
	if st.fmt.UploadsBaseURL == "" {
		st.fmt.UploadsBaseURL = st.uploadsBaseURL()
	}
	st.fmt.UseSmilies = truthy(st.option("use_smilies", "1"))
	st.fmt.HTTPSMigrationRequired = truthy(st.option("https_migration_required", ""))
	return st, nil
}

// uploadsBaseURL 对应单站点、未定义 UPLOADS 常量时 _wp_upload_dir() 的 baseurl。
func (s *site) uploadsBaseURL() string {
	if u := s.option("upload_url_path", ""); u != "" {
		return u
	}
	siteURL := s.option("siteurl", "")
	path := strings.TrimSpace(s.option("upload_path", ""))
	if path == "" || path == "wp-content/uploads" {
		return siteURL + "/wp-content/uploads"
	}
	return strings.TrimRight(siteURL, "/") + "/" + path
}

// option 对应 get_option( $name, $def )：选项不存在时返回默认值。
func (s *site) option(name, def string) string {
	if v, ok := s.options[name]; ok {
		return v
	}
	return def
}

// truthy 按 PHP 的规则判断字符串是否为真。
func truthy(v string) bool {
	return v != "" && v != "0"
}

// gmtOffset 对应 get_option( 'gmt_offset' )：设置了 timezone_string 时，
// WP 会通过 wp_timezone_override_offset 按该时区的当前偏移计算。
func (s *site) gmtOffset() float64 {
	if name := s.option("timezone_string", ""); name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			_, seconds := time.Now().In(loc).Zone()
			return float64(seconds) / secondsPerHour
		}
	}
	offset, err := strconv.ParseFloat(s.option("gmt_offset", "0"), 64)
	if err != nil {
		return 0
	}
	return offset
}

// timezone 对应 wp_timezone()：优先使用 timezone_string，否则按 gmt_offset 构造固定偏移时区。
func (s *site) timezone() *time.Location {
	if name := s.option("timezone_string", ""); name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc
		}
	}
	offset, err := strconv.ParseFloat(s.option("gmt_offset", "0"), 64)
	if err != nil {
		offset = 0
	}
	seconds := int(math.Round(offset * secondsPerHour))
	return time.FixedZone(formatOffset(seconds), seconds)
}

func formatOffset(seconds int) string {
	sign := '+'
	if seconds < 0 {
		sign = '-'
		seconds = -seconds
	}
	return fmt.Sprintf("%c%02d:%02d", sign, seconds/secondsPerHour, seconds%secondsPerHour/secondsPerMinute)
}

// stickyPosts 解析 sticky_posts 选项（PHP 序列化的整数数组）。
var stickyItemRe = regexp.MustCompile(`i:\d+;(?:i:(\d+)|s:\d+:"(\d+)");`)

func (s *site) stickyPosts() map[uint64]bool {
	result := map[uint64]bool{}
	for _, m := range stickyItemRe.FindAllStringSubmatch(s.option("sticky_posts", ""), -1) {
		raw := m[1]
		if raw == "" {
			raw = m[2]
		}
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
			result[id] = true
		}
	}
	return result
}

// avatarURLs 对应 rest_get_avatar_urls()；email 为空时生成不带哈希的默认头像地址。
// name 仅在默认头像类型为 initials 时使用。
func (s *site) avatarURLs(email, name string) map[string]string {
	hash := ""
	if email != "" {
		sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
		hash = hex.EncodeToString(sum[:])
	}

	def := s.option("avatar_default", "mystery")
	if def == "" {
		def = "mystery"
	}
	switch def {
	case "mm", "mystery", "mysteryman":
		def = "mm"
	case "gravatar_default":
		def = ""
	}
	rating := strings.ToLower(s.option("avatar_rating", ""))

	// 尺寸对应 rest_get_avatar_sizes() 的默认值。
	sizes := [...]int{24, 48, 96}
	urls := make(map[string]string, len(sizes))
	for _, size := range sizes {
		// 参数顺序与 get_avatar_data() 中 add_query_arg 的结果一致：s、d、r、initials。
		query := []string{"s=" + strconv.Itoa(size)}
		if def != "" {
			query = append(query, "d="+rawURLEncode(def))
		}
		if rating != "" {
			query = append(query, "r="+rawURLEncode(rating))
		}
		if def == "initials" && name != "" {
			query = append(query, "initials="+rawURLEncode(avatarInitials(name)))
		}
		urls[strconv.Itoa(size)] = "https://secure.gravatar.com/avatar/" + hash + "?" + strings.Join(query, "&")
	}
	return urls
}

var cjkRe = regexp.MustCompile(`\p{Han}|\p{Hiragana}|\p{Katakana}|\p{Hangul}`)

func avatarInitials(name string) string {
	runes := []rune(name)
	if !strings.Contains(name, " ") || cjkRe.MatchString(name) {
		return string(runes[:min(maxInitials, len(runes))])
	}
	last := []rune(name[strings.LastIndex(name, " ")+1:])
	if len(last) == 0 {
		return string(runes[:1])
	}
	return string(runes[0]) + string(last[0])
}

// rawURLEncode 对应 PHP 的 rawurlencode（RFC 3986）。
func rawURLEncode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// mysqlZero 判断 MySQL 的 0000-00-00 00:00:00 日期（驱动解析为零值）。
func mysqlZero(t time.Time) bool {
	return t.IsZero() || t.Year() <= 0
}

// formatDate 对应 mysql_to_rfc3339()，输出不带时区的 Y-m-d\TH:i:s。
func formatDate(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

// prepareDate 对应 WP_REST_Posts_Controller::prepare_date_response()。
func prepareDate(gmt, local time.Time) *string {
	if mysqlZero(gmt) {
		return nil
	}
	s := formatDate(local)
	return &s
}

// gmtFromLocal 对应 get_gmt_from_date()：把站点时区的墙上时间转换为 UTC。
func (s *site) gmtFromLocal(local time.Time) time.Time {
	t := time.Date(
		local.Year(), local.Month(), local.Day(),
		local.Hour(), local.Minute(), local.Second(), 0, s.tz,
	)
	return t.UTC()
}

// localWallClock 把任意时刻转换成站点时区的墙上时间，并以 UTC 表示，用于和 post_date 比较。
func (s *site) localWallClock(t time.Time) time.Time {
	l := t.In(s.tz)
	return time.Date(l.Year(), l.Month(), l.Day(), l.Hour(), l.Minute(), l.Second(), 0, time.UTC)
}

// parseRESTDate 对应 rest_parse_date() 接受的格式；没有时区时按站点时区解释。
var restDateRe = regexp.MustCompile(
	`^\d{4}-\d{2}-\d{2}[Tt ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}(?::\d{2})?)?$`,
)

func (s *site) parseRESTDate(v string) (time.Time, bool) {
	if !restDateRe.MatchString(v) {
		return time.Time{}, false
	}
	v = strings.Replace(strings.Replace(v, "t", "T", 1), " ", "T", 1)
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05Z07"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t, true
		}
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05.999999999", v, s.tz)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// siteIndex 是 GET /wp-json/ 的响应，只包含本服务支持的字段。
type siteIndex struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	URL            string  `json:"url"`
	Home           string  `json:"home"`
	GMTOffset      float64 `json:"gmt_offset"`
	TimezoneString string  `json:"timezone_string"`
}

type siteIndexOutput struct {
	Body siteIndex
}

func (s *Server) getIndex(ctx context.Context, _ *struct{}) (*siteIndexOutput, error) {
	st, err := s.loadSite(ctx)
	if err != nil {
		return nil, internalError(ctx, err)
	}
	return &siteIndexOutput{Body: siteIndex{
		Name:           st.option("blogname", ""),
		Description:    st.option("blogdescription", ""),
		URL:            st.option("siteurl", ""),
		Home:           st.option("home", ""),
		GMTOffset:      st.gmtOffset(),
		TimezoneString: st.option("timezone_string", ""),
	}}, nil
}
