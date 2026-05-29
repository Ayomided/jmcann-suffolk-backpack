package ui

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type TemplateCache struct {
	Cache map[string]*template.Template
}

func humanDate(t time.Time) string {
	return t.Format("3:04am on 02 Jan 2006")
}

func notificationArgs(values ...any) (map[string]any, error) {
	m := make(map[string]any)
	for i := 0; i < len(values); i += 2 {
		m[values[i].(string)] = values[i+1]
	}
	return m, nil
}

func statusToLabel(status string) string {
	switch status {
	case "inProgress":
		return "In Progress"
	default:
		return status
	}
}

var functions = template.FuncMap{
	"humanDate":   humanDate,
	"dict":        notificationArgs,
	"statToLabel": statusToLabel,
}

func NewTemplateCache() (*TemplateCache, error) {
	cache := map[string]*template.Template{}

	var pages []string
	err := filepath.WalkDir("./cmd/ui/templates/pages", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".templ.html") {
			pages = append(pages, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles("./cmd/ui/templates/base.templ.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./cmd/ui/templates/partials/*.templ.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return &TemplateCache{
		Cache: cache,
	}, nil
}
