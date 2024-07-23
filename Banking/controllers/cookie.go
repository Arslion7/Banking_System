package controllers

import "net/http"

const CookieSession = "session"

func newCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name: name,
		Value: value,
		Path: "/",
		HttpOnly: true,
	}
}

func setCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, newCookie(name, value))
}

func deleteCookie(w http.ResponseWriter, name string) {
	cookie := newCookie(name, "")
	cookie.MaxAge = -1
	http.SetCookie(w,cookie)
}