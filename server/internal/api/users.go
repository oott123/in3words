package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oott123/in3words/server/internal/wpdb"
)

// publicPostTypes 返回 has_published_posts 检查的文章类型，对应 show_in_rest 为真的内置类型。
func publicPostTypes() []string {
	return []string{
		"post", "page", "attachment", "nav_menu_item", "wp_block", "wp_template",
		"wp_template_part", "wp_global_styles", "wp_navigation", "wp_font_family", "wp_font_face",
	}
}

// User 是 /wp/v2/users 的响应条目，也用于 _embedded.author。
type User struct {
	ID         uint64            `json:"id"`
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Slug       string            `json:"slug"`
	AvatarURLs map[string]string `json:"avatar_urls,omitempty"`
}

func (s *site) prepareUser(id uint64, login, nicename, email, url, displayName string) User {
	u := User{ID: id, Name: displayName, URL: url, Slug: nicename}
	// show_avatars 关闭时 WP 的 schema 中没有 avatar_urls 字段。
	if truthy(s.option("show_avatars", "1")) {
		name := displayName
		if name == "" {
			name = login
		}
		u.AvatarURLs = s.avatarURLs(email, name)
	}
	return u
}

type listUsersInput struct {
	Pagination

	Slug []string `query:"slug" doc:"用户别名（user_nicename）"`
}

type listUsersOutput struct {
	Total      int64 `header:"X-WP-Total"`
	TotalPages int64 `header:"X-WP-TotalPages"`
	Body       []User
}

// listUsers 对应 WP_REST_Users_Controller::get_items() 的访客视角：只列出发表过文章的用户，按显示名排序。
func (s *Server) listUsers(ctx context.Context, in *listUsersInput) (*listUsersOutput, error) {
	st, err := s.loadSite(ctx)
	if err != nil {
		return nil, internalError(ctx, err)
	}
	slugs := sanitizeSlugs(in.Slug)
	filterSlugs := flag(len(in.Slug) > 0)

	total, err := s.q.CountPublishedAuthors(ctx, wpdb.CountPublishedAuthorsParams{
		FilterSlugs: filterSlugs,
		Slugs:       slugs,
		PostTypes:   publicPostTypes(),
	})
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("count users: %w", err))
	}

	out := &listUsersOutput{Total: total, TotalPages: in.totalPages(total), Body: []User{}}
	if total == 0 {
		return out, nil
	}

	rows, err := s.q.ListPublishedAuthors(ctx, wpdb.ListPublishedAuthorsParams{
		FilterSlugs: filterSlugs,
		Slugs:       slugs,
		PostTypes:   publicPostTypes(),
		Limit:       int32(in.PerPage), //nolint:gosec // per_page 已由参数校验限定范围
		Offset:      in.offset(),
	})
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("list users: %w", err))
	}
	for _, r := range rows {
		u := st.prepareUser(r.ID, r.UserLogin, r.UserNicename, r.UserEmail, r.UserUrl, r.DisplayName)
		out.Body = append(out.Body, u)
	}
	return out, nil
}

// embedUsers 返回 _embedded.author 的内容：可见的用户为 User，否则为与 WP 相同的错误对象。
func (s *Server) embedUsers(ctx context.Context, st *site, ids []uint64) (map[uint64]any, error) {
	rows, err := s.q.ListUsersByIDs(ctx, wpdb.ListUsersByIDsParams{PostTypes: publicPostTypes(), Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("list users by id: %w", err)
	}

	result := make(map[uint64]any, len(ids))
	for _, id := range ids {
		result[id] = newError(http.StatusNotFound, "rest_user_invalid_id", "Invalid user ID.")
	}
	for _, r := range rows {
		if !r.HasPublishedPosts {
			result[r.ID] = newError(
				http.StatusUnauthorized,
				"rest_user_cannot_view",
				"Sorry, you are not allowed to list users.",
			)
			continue
		}
		result[r.ID] = st.prepareUser(r.ID, r.UserLogin, r.UserNicename, r.UserEmail, r.UserUrl, r.DisplayName)
	}
	return result, nil
}
