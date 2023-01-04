package render

import (
	"html/template"
	"net/http"
)

type HTMLRender interface {
	Instance(name string, data any, debug bool) Render
}

type HTMLEngine struct {
	Files    []string
	Glob     string
	FuncMap  template.FuncMap
	template *template.Template
}

type HTML struct {
	Template *template.Template
	Name     string
	Data     any
}

func NewHTMLRender(funcMap template.FuncMap, files ...string) HTMLRender {
	return &HTMLEngine{
		Files:   files,
		FuncMap: funcMap,
	}
}

func NewGlobHTMLRender(funcMap template.FuncMap, glob string) HTMLRender {
	return &HTMLEngine{
		Glob:    glob,
		FuncMap: funcMap,
	}
}

func (r *HTMLEngine) Instance(name string, data any, debug bool) Render {
	return HTML{Template: r.loadTemplate(debug), Name: name, Data: data}
}

func (r *HTMLEngine) loadTemplate(debug bool) *template.Template {
	if r.FuncMap == nil {
		r.FuncMap = template.FuncMap{}
	}
	if r.template != nil && !debug {
		return r.template
	}

	if len(r.Files) > 0 {
		r.template = template.Must(template.New("").Funcs(r.FuncMap).ParseFiles(r.Files...))
	} else if r.Glob != "" {
		r.template = template.Must(template.New("").Funcs(r.FuncMap).ParseGlob(r.Glob))
	} else {
		panic("the HTMLEngine was created without files or glob pattern")
	}

	return r.template
}

func (r HTML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	if r.Name == "" {
		return r.Template.Execute(w, r.Data)
	}
	return r.Template.ExecuteTemplate(w, r.Name, r.Data)
}

func (r HTML) WriteContentType(w http.ResponseWriter) {
	w.Header().Set(headerContentType, mimeTextHTMLCharsetUTF8)
}
