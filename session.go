package main

import "net/http"

func isLoggedIn(req *http.Request) bool {
	cookie, err := req.Cookie()

}
