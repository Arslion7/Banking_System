package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"myapp/context"
	"myapp/models"
	"net/http"

	"github.com/gorilla/csrf"
)

type Template struct {
	htmlTpl *template.Template
}

func Parse(fs embed.FS, pattern ...string) (Template, error) {
	tpl := template.New(pattern[0]).Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", fmt.Errorf("Implement csrfField in Execute")
		},
		"currentUser": func() (template.HTML, error) {
			return "", fmt.Errorf("Implement currntUser in Execute")
		},
	})
	tmp, err := tpl.ParseFS(fs, pattern...)

	if err != nil {
		return Template{}, fmt.Errorf("parsing error: %w", err)
	}

	return Template{
		htmlTpl: tmp,
	}, err
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return 
	}

	tpl = tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.TemplateField(r)
		},
		"currentUser": func() *models.User {
			return context.User(r.Context())
		},
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		http.Error(w, "Error in executing the template", http.StatusInternalServerError)
		return 
	}
	io.Copy(w, &buf)
}