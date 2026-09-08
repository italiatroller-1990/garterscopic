package binding

import (
	"fmt"
	"strings"
)

type Resolver struct {
	frontmatter map[string]any
	page        PageInfo
	site        SiteInfo
	content     ContentInfo
}

type PageInfo struct {
	Route string
	URL   string
	Type  string
}

type SiteInfo struct {
	Name    string
	BaseURL string
}

type ContentInfo struct {
	Posts     []map[string]any
	Pages     []map[string]any
	Sections  []string
	Tags      []string
	TagIndex  map[string][]map[string]any
	SectionPosts map[string][]map[string]any
}

func NewResolver(frontmatter map[string]any, page PageInfo, site SiteInfo) *Resolver {
	return &Resolver{
		frontmatter: frontmatter,
		page:        page,
		site:        site,
	}
}

func NewResolverWithContent(frontmatter map[string]any, page PageInfo, site SiteInfo, content ContentInfo) *Resolver {
	return &Resolver{
		frontmatter: frontmatter,
		page:        page,
		site:        site,
		content:     content,
	}
}

func (r *Resolver) Resolve(path string) (any, error) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("binding '%s' must use a namespace (frontmatter.*, page.*, site.*, posts.*, sections.*, tags.*)", path)
	}

	namespace := parts[0]
	field := parts[1]

	switch namespace {
	case "frontmatter":
		return r.resolveFrontmatter(field)
	case "page":
		return r.resolvePage(field)
	case "site":
		return r.resolveSite(field)
	case "posts":
		return r.resolvePosts(field)
	case "sections":
		return r.resolveSections(field)
	case "tags":
		return r.resolveTags(field)
	default:
		return nil, fmt.Errorf("unknown binding namespace '%s' (must be frontmatter, page, site, posts, sections, or tags)", namespace)
	}
}

func (r *Resolver) resolveFrontmatter(field string) (any, error) {
	if r.frontmatter == nil {
		return nil, nil
	}

	parts := strings.Split(field, ".")
	var current any = r.frontmatter
	for _, p := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[p]
			if !ok {
				return nil, nil
			}
			current = val
		default:
			return nil, nil
		}
	}
	return current, nil
}

func (r *Resolver) resolvePage(field string) (any, error) {
	switch field {
	case "route":
		return r.page.Route, nil
	case "url":
		return r.page.URL, nil
	case "type":
		return r.page.Type, nil
	default:
		return nil, fmt.Errorf("unknown page field '%s'", field)
	}
}

func (r *Resolver) resolveSite(field string) (any, error) {
	switch field {
	case "name":
		return r.site.Name, nil
	case "base_url":
		return r.site.BaseURL, nil
	default:
		return nil, fmt.Errorf("unknown site field '%s'", field)
	}
}

func (r *Resolver) resolvePosts(field string) (any, error) {
	switch field {
	case "all":
		return r.content.Posts, nil
	default:
		return nil, fmt.Errorf("unknown posts field '%s' (use 'all')", field)
	}
}

func (r *Resolver) resolveSections(field string) (any, error) {
	switch field {
	case "all":
		return r.content.Sections, nil
	default:
		if r.content.SectionPosts != nil {
			if posts, ok := r.content.SectionPosts[field]; ok {
				return posts, nil
			}
		}
		return nil, fmt.Errorf("unknown section '%s'", field)
	}
}

func (r *Resolver) resolveTags(field string) (any, error) {
	switch field {
	case "all":
		return r.content.Tags, nil
	default:
		if r.content.TagIndex != nil {
			if posts, ok := r.content.TagIndex[field]; ok {
				return posts, nil
			}
		}
		return nil, fmt.Errorf("unknown tag '%s'", field)
	}
}

func (r *Resolver) ResolveMap(values []string) (map[string]any, error) {
	result := make(map[string]any)
	for _, v := range values {
		resolved, err := r.Resolve(v)
		if err != nil {
			return nil, err
		}
		result[v] = resolved
	}
	return result, nil
}

func ApplyBindings(options map[string]any, resolver *Resolver) (map[string]any, error) {
	if options == nil {
		return map[string]any{}, nil
	}

	result := make(map[string]any)
	for key, value := range options {
		result[key] = value
	}

	if fromValues, ok := options["values"].([]any); ok {
		delete(result, "values")
		for _, v := range fromValues {
			vs, ok := v.(string)
			if !ok {
				continue
			}
			resolved, err := resolver.Resolve(vs)
			if err != nil {
				return nil, fmt.Errorf("binding error for '%s': %w", vs, err)
			}
			parts := strings.Split(vs, ".")
			if len(parts) >= 2 {
				result[parts[len(parts)-1]] = resolved
			}
		}
	}

	if fromMap, ok := options["from"].(map[string]any); ok {
		for target, sourceAny := range fromMap {
			source, ok := sourceAny.(string)
			if !ok {
				continue
			}
			resolved, err := resolver.Resolve(source)
			if err != nil {
				return nil, fmt.Errorf("binding error for '%s': %w", source, err)
			}
			result[target] = resolved
		}
	}

	return result, nil
}
