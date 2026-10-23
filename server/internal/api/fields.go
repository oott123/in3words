package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

var listSeparatorRe = regexp.MustCompile(`[\s,]+`)

// filterFields 对应 rest_filter_response_fields()：按 _fields 参数裁剪响应，支持 a.b 形式的嵌套字段。
func filterFields(ctx huma.Context, _ string, v any) (any, error) {
	u := ctx.URL()
	raw := u.Query().Get("_fields")
	if raw == "" {
		return v, nil
	}
	if _, isErr := v.(*Error); isErr {
		return v, nil
	}
	keyed := fieldsTree(listSeparatorRe.Split(strings.TrimSpace(raw), -1))
	if len(keyed) == 0 {
		return v, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode response for _fields: %w", err)
	}
	var data any
	if err = json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("decode response for _fields: %w", err)
	}
	if list, ok := data.([]any); ok {
		for i, item := range list {
			list[i] = intersectFields(item, keyed)
		}
		return list, nil
	}
	return intersectFields(data, keyed), nil
}

// fieldsTree 把 ["a", "b.c"] 转成 {a: true, b: {c: true}}；父字段已被整体选中时忽略其子字段。
func fieldsTree(fields []string) map[string]any {
	tree := map[string]any{}
	for _, field := range fields {
		if field == "" {
			continue
		}
		parts := strings.Split(field, ".")
		node := tree
		for _, part := range parts[:len(parts)-1] {
			next, ok := node[part].(map[string]any)
			if !ok {
				if node[part] == true {
					node = nil
					break
				}
				next = map[string]any{}
				node[part] = next
			}
			node = next
		}
		if node != nil {
			node[parts[len(parts)-1]] = true
		}
	}
	return tree
}

func intersectFields(data any, tree map[string]any) any {
	obj, ok := data.(map[string]any)
	if !ok {
		return data
	}
	out := make(map[string]any, len(tree))
	for key, sub := range tree {
		value, exists := obj[key]
		if !exists {
			continue
		}
		if subtree, nested := sub.(map[string]any); nested {
			out[key] = intersectFields(value, subtree)
		} else {
			out[key] = value
		}
	}
	return out
}
