package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/oott123/in3words/server/internal/wpdb"
)

// Comment 是 /wp/v2/comments 的响应条目。
type Comment struct {
	ID               uint64            `json:"id"`
	Post             uint64            `json:"post"`
	Parent           uint64            `json:"parent"`
	Author           uint64            `json:"author"`
	AuthorName       string            `json:"author_name"`
	AuthorURL        string            `json:"author_url"`
	Date             string            `json:"date"`
	DateGMT          string            `json:"date_gmt"`
	Content          Rendered          `json:"content"`
	Status           string            `json:"status"`
	Type             string            `json:"type"`
	AuthorAvatarURLs map[string]string `json:"author_avatar_urls,omitempty"`
}

type listCommentsInput struct {
	Pagination

	Post   []uint64 `query:"post"   doc:"文章 ID"`
	Parent []uint64 `query:"parent" doc:"父评论 ID"`
	Order  string   `query:"order"  doc:"按时间排序的方向" enum:"asc,desc" default:"desc"`
}

type listCommentsOutput struct {
	Total      int64 `header:"X-WP-Total"`
	TotalPages int64 `header:"X-WP-TotalPages"`
	Body       []Comment
}

// listComments 对应 WP_REST_Comments_Controller::get_items() 的访客视角（status=approve、type=comment）。
func (s *Server) listComments(ctx context.Context, in *listCommentsInput) (*listCommentsOutput, error) {
	if err := s.checkCommentPosts(ctx, in.Post); err != nil {
		return nil, err
	}
	st, err := s.loadSite(ctx)
	if err != nil {
		return nil, internalError(ctx, err)
	}

	count := wpdb.CountCommentsParams{
		FilterPosts:   flag(len(in.Post) > 0),
		Posts:         in.Post,
		FilterParents: flag(len(in.Parent) > 0),
		Parents:       in.Parent,
	}
	total, err := s.q.CountComments(ctx, count)
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("count comments: %w", err))
	}

	out := &listCommentsOutput{Total: total, TotalPages: in.totalPages(total), Body: []Comment{}}
	if total == 0 {
		return out, nil
	}

	rows, err := s.q.ListComments(ctx, wpdb.ListCommentsParams{
		FilterPosts:   count.FilterPosts,
		Posts:         count.Posts,
		FilterParents: count.FilterParents,
		Parents:       count.Parents,
		Ascending:     flag(in.Order == "asc"),
		Limit:         int32(in.PerPage), //nolint:gosec // per_page 已由参数校验限定范围
		Offset:        in.offset(),
	})
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("list comments: %w", err))
	}
	for _, r := range rows {
		out.Body = append(out.Body, s.prepareComment(st, r))
	}
	return out, nil
}

// checkCommentPosts 对应 get_items_permissions_check() 中对 post 参数的检查：
// 请求的文章存在但访客无权阅读（未发布或有密码）时整个请求返回 401。
func (s *Server) checkCommentPosts(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	if slices.Contains(ids, 0) {
		return newError(
			http.StatusUnauthorized,
			"rest_cannot_read",
			"Sorry, you are not allowed to read comments without a post.",
		)
	}
	posts, err := s.q.ListPostsByIDs(ctx, ids)
	if err != nil {
		return internalError(ctx, fmt.Errorf("list comment posts: %w", err))
	}
	for _, p := range posts {
		if p.PostStatus != "publish" || p.PostPassword != "" {
			return newError(
				http.StatusUnauthorized,
				"rest_cannot_read_post",
				"Sorry, you are not allowed to read the post for this comment.",
			)
		}
	}
	return nil
}

func (s *Server) prepareComment(st *site, r wpdb.ListCommentsRow) Comment {
	commentType := r.CommentType
	if commentType == "" {
		commentType = "comment"
	}
	c := Comment{
		ID:         r.CommentID,
		Post:       r.CommentPostID,
		Parent:     r.CommentParent,
		Author:     r.UserID,
		AuthorName: r.CommentAuthor,
		AuthorURL:  r.CommentAuthorUrl,
		Date:       formatDate(r.CommentDate),
		DateGMT:    formatDate(r.CommentDateGmt),
		Content:    Rendered{Rendered: s.render(renderComment, r.CommentID, r.CommentContent, st.fmt, nil)},
		Status:     commentStatus(r.CommentApproved),
		Type:       commentType,
	}
	if truthy(st.option("show_avatars", "1")) {
		// 与 get_avatar_data() 一致：有关联用户时用用户邮箱，否则用评论者填写的邮箱。
		email := r.UserEmail
		if email == "" {
			email = r.CommentAuthorEmail
		}
		c.AuthorAvatarURLs = st.avatarURLs(email, r.CommentAuthor)
	}
	return c
}

// commentStatus 对应 prepare_status_response()。
func commentStatus(approved string) string {
	switch approved {
	case "hold", "0":
		return "hold"
	case "approve", "1":
		return "approved"
	}
	return approved
}
