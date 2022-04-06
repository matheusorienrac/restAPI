package main

import "net/http"

func main() {
	http.ListenAndServe("/", "cert.pem", "key.pem", index)
}
