package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/oott123/in3words/server/internal/logging"
)

// Error 是与 WP_Error 经 REST 输出后格式一致的错误响应。
type Error struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Data    ErrorData `json:"data"`

	msg    message
	params map[string]message
}

// ErrorData 对应 WP_Error 的 data 字段。
type ErrorData struct {
	Status int               `json:"status"`
	Params map[string]string `json:"params,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// GetStatus 实现 huma.StatusError。
func (e *Error) GetStatus() int {
	return e.Data.Status
}

// localize 按站点语言重新生成面向最终用户的文字。
func (e *Error) localize(locale string) {
	e.Message = e.msg.render(locale)
	if len(e.params) == 0 {
		return
	}
	e.Data.Params = make(map[string]string, len(e.params))
	for name, m := range e.params {
		e.Data.Params[name] = m.render(locale)
	}
}

func newError(status int, code, format string, args ...any) *Error {
	e := &Error{Code: code, Data: ErrorData{Status: status}, msg: message{format: format, args: args}}
	e.localize("")
	return e
}

func errNoRoute() *Error {
	return newError(http.StatusNotFound, "rest_no_route", "No route was found matching the URL and request method.")
}

// paramError 是参数校验失败的原因，可以被翻译。
type paramError struct {
	name string
	msg  message
}

func (p *paramError) Error() string {
	return p.name + ": " + p.msg.render("")
}

// ErrorDetail 实现 huma.ErrorDetailer。
func (p *paramError) ErrorDetail() *huma.ErrorDetail {
	return &huma.ErrorDetail{Location: "query." + p.name, Message: p.msg.render("")}
}

func newParamError(name, format string, args ...any) *paramError {
	return &paramError{name: name, msg: message{format: format, args: args}}
}

func errInvalidParam(name, format string, args ...any) *Error {
	return invalidParamsError([]error{newParamError(name, format, args...)})
}

// internalError 记录内部错误，只向客户端返回通用信息。
func internalError(ctx context.Context, err error) *Error {
	logging.FromContext(ctx).WithError(err).Error("handle request")
	return newError(
		http.StatusInternalServerError,
		"internal_server_error",
		"There has been a critical error on this website.",
	)
}

// hookHumaErrors 让 huma 自身产生的错误（参数校验失败等）也使用 WP 的错误格式。
// huma.NewError 是包级变量，因此该设置对进程内所有 huma API 生效。
func hookHumaErrors() {
	//nolint:reassign // huma 规定通过替换 NewError 来自定义错误格式
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		if status == http.StatusUnprocessableEntity || status == http.StatusBadRequest {
			return invalidParamsError(errs)
		}
		code := "rest_error"
		if status >= http.StatusInternalServerError {
			code = "internal_server_error"
		}
		return newError(status, code, "%s", msg)
	}
}

func invalidParamsError(errs []error) *Error {
	params := map[string]message{}
	for _, err := range errs {
		if pe, ok := errors.AsType[*paramError](err); ok {
			params[pe.name] = pe.msg
			continue
		}
		var detail *huma.ErrorDetail
		if !errors.As(err, &detail) {
			continue
		}
		name := detail.Location
		name = strings.TrimPrefix(name, "query.")
		name = strings.TrimPrefix(name, "path.")
		params[name] = message{format: "%s", args: []any{detail.Message}}
	}
	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}
	sort.Strings(names)

	e := newError(
		http.StatusBadRequest,
		"rest_invalid_param",
		"Invalid parameter(s): %s",
		strings.Join(names, ", "),
	)
	if len(params) > 0 {
		e.params = params
		e.localize("")
	}
	return e
}

// message 是待翻译的文字，format 使用 WP 英文原文（占位符改为 Go 的格式）。
type message struct {
	format string
	args   []any
}

func (m message) render(locale string) string {
	format := m.format
	if t, ok := translations[locale][format]; ok {
		format = t
	}
	if len(m.args) == 0 {
		return format
	}
	return fmt.Sprintf(format, m.args...)
}

// translations 摘自 WordPress 官方语言包，只包含本服务会输出的文字。
//
//nolint:gochecknoglobals // 常量表
var translations = map[string]map[string]string{
	"zh_CN": {
		"Invalid parameter(s): %s": "无效参数：%s",
		"The page number requested is larger than the number of pages available.": "请求的页码大于总页数。",
		"Invalid post ID.":                                              "文章 ID 无效。",
		"Sorry, you are not allowed to do that.":                        "抱歉，您不能这么做。",
		"No route was found matching the URL and request method.":       "未找到匹配 URL 和请求方式的路由。",
		"Invalid date.":                                                 "无效日期。",
		"Invalid user ID.":                                              "用户 ID 无效。",
		"Sorry, you are not allowed to list users.":                     "抱歉，您不能列出用户。",
		"Sorry, you are not allowed to read the post for this comment.": "抱歉，您不能阅读此评论对应的文章。",
		"Sorry, you are not allowed to read comments without a post.":   "抱歉，您不能阅读没有对应文章的评论。",
		"There has been a critical error on this website.":              "此站点遇到了致命错误。",
		"%s must be between %d (inclusive) and %d (inclusive)":          "%s 必须介于%d（含）和%d（含）之间",
		"%s must be greater than or equal to %d":                        "%s 必须大于或等于%d",
	},
}
