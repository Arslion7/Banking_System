package controllers

import (
	"html/template"
	"net/http"
)

func StaticHandler(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, nil)
	}
}

func FAQ(tpl Template) http.HandlerFunc {
	questions := []struct{
		Question string
		Answer template.HTML
	}{
		{
			Question: "Which club",
			Answer: "Arsenal",
		},
		{
			Question: "What link",
			Answer: `<a href="arsenal.com">arsenal</a>`,
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, questions)
	}
}