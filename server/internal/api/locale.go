package api

import (
	"context"
	"sync"
	"time"

	"github.com/oott123/in3words/server/internal/logging"
)

const localeTTL = time.Minute

// localeCache 缓存站点语言（WPLANG 选项），错误响应在未加载站点设置时也需要它。
type localeCache struct {
	mu      sync.Mutex
	value   string
	expires time.Time
}

func (s *Server) locale(ctx context.Context) string {
	s.lang.mu.Lock()
	defer s.lang.mu.Unlock()
	if time.Now().Before(s.lang.expires) {
		return s.lang.value
	}
	rows, err := s.q.ListOptions(ctx, []string{"WPLANG"})
	if err != nil {
		logging.FromContext(ctx).WithError(err).Warn("load site locale")
		return s.lang.value
	}
	s.lang.value = ""
	for _, row := range rows {
		s.lang.value = row.OptionValue
	}
	s.lang.expires = time.Now().Add(localeTTL)
	return s.lang.value
}
