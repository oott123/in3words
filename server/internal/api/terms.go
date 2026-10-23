package api

import (
	"context"
	"fmt"

	"github.com/oott123/in3words/server/internal/wpdb"
)

const (
	taxonomyCategory = "category"
	taxonomyTag      = "post_tag"
)

// Term 是 /wp/v2/categories 与 /wp/v2/tags 的响应条目。
type Term struct {
	ID          uint64  `json:"id"`
	Count       int64   `json:"count"`
	Description string  `json:"description"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Taxonomy    string  `json:"taxonomy"`
	Parent      *uint64 `json:"parent,omitempty"`
	Meta        []any   `json:"meta"`
}

// EmbedTerm 是 embed 上下文中的分类或标签。
type EmbedTerm struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Taxonomy string `json:"taxonomy"`
}

func embedTerm(id uint64, name, slug, taxonomy string) EmbedTerm {
	return EmbedTerm{ID: id, Name: name, Slug: slug, Taxonomy: taxonomy}
}

type listTermsInput struct {
	Pagination

	OrderBy string   `query:"orderby" enum:"name,count" default:"name" doc:"排序字段"`
	Order   string   `query:"order"   enum:"asc,desc"   default:"asc"  doc:"排序方向"`
	Slug    []string `query:"slug"                                     doc:"别名"`
}

type listTermsOutput struct {
	Total      int64 `header:"X-WP-Total"`
	TotalPages int64 `header:"X-WP-TotalPages"`
	Body       []Term
}

func (s *Server) listCategories(ctx context.Context, in *listTermsInput) (*listTermsOutput, error) {
	return s.listTerms(ctx, taxonomyCategory, in)
}

func (s *Server) listTags(ctx context.Context, in *listTermsInput) (*listTermsOutput, error) {
	return s.listTerms(ctx, taxonomyTag, in)
}

// listTerms 对应 WP_REST_Terms_Controller::get_items()；hide_empty 默认为 false，超出页数时返回空列表。
func (s *Server) listTerms(ctx context.Context, taxonomy string, in *listTermsInput) (*listTermsOutput, error) {
	slugs := sanitizeSlugs(in.Slug)
	filterSlugs := flag(len(in.Slug) > 0)

	total, err := s.q.CountTerms(ctx, wpdb.CountTermsParams{
		Taxonomy:    taxonomy,
		FilterSlugs: filterSlugs,
		Slugs:       slugs,
	})
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("count %s terms: %w", taxonomy, err))
	}

	out := &listTermsOutput{Total: total, TotalPages: in.totalPages(total), Body: []Term{}}
	if total == 0 {
		return out, nil
	}

	rows, err := s.q.ListTerms(ctx, wpdb.ListTermsParams{
		Taxonomy:    taxonomy,
		FilterSlugs: filterSlugs,
		Slugs:       slugs,
		Sort:        in.OrderBy + "_" + in.Order,
		Limit:       int32(in.PerPage), //nolint:gosec // per_page 已由参数校验限定范围
		Offset:      in.offset(),
	})
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("list %s terms: %w", taxonomy, err))
	}
	for _, row := range rows {
		t := Term{
			ID:          row.TermID,
			Count:       row.Count,
			Description: row.Description,
			Name:        row.Name,
			Slug:        row.Slug,
			Taxonomy:    row.Taxonomy,
			Meta:        []any{},
		}
		// 只有层级分类法（category）才输出 parent。
		if taxonomy == taxonomyCategory {
			parent := row.Parent
			t.Parent = &parent
		}
		out.Body = append(out.Body, t)
	}
	return out, nil
}
