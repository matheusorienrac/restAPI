package main

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func alreadyLoggedIn(req *http.Request) bool {
	_, err := req.Cookie("session")
	if err != nil {
		return false
	}
	return true
}

func signup(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		tpl.ExecuteTemplate(res, "signup.gohtml", nil)
		return
	}

	u := req.FormValue("username")
	p := req.FormValue("password")
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	q := `
		INSERT INTO USERS (USERNAME, PASSWORD)
		VALUES ($1, $2);
	`
	result, err := db.Exec(q, u, hashPassword)
	if err != nil {
		http.Error(res, err.Error(), http.StatusForbidden)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(res, err.Error(), http.StatusForbidden)
		return

	}
	if rowsAffected > 0 {
		tpl.ExecuteTemplate(res, "signup.gohtml", "Account creation successful")
		return
	} else {
		res.Write([]byte("something went wrong."))

	}

}

func login(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")
	var hashed_password string
	q := `SELECT PASSWORD FROM USERS WHERE USERNAME = $1`
	row := db.QueryRow(q, username)
	if err = row.Scan(&hashed_password); err != nil {
		tpl.ExecuteTemplate(res, "index.gohtml", "Invalid Username or Password.")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(password))
	if err != nil {
		tpl.ExecuteTemplate(res, "index.gohtml", "Invalid Username or Password.")
		return
	}

	// if there is no cookie, means we have to create a session
	_, err := req.Cookie("session")
	if err != nil {
		id := uuid.NewV4()
		cookie := &http.Cookie{
			Name:     "session",
			Value:    id.String(),
			Secure:   true,
			HttpOnly: true,
		}
		http.SetCookie(res, cookie)
		_, err = db.Exec("INSERT INTO SESSIONS(sid, username) VALUES ($1, $2)", cookie.Value, cookie.Name)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(res, req, "/pokedex", http.StatusSeeOther)

}

func logout(res http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("session")
	// means we have a cookie, so we should delete it
	if err == nil {
		// delete session from session table
		_, err := db.Exec("DELETE FROM SESSIONS WHERE SID = $1", cookie.Value)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		// delete cookie
		cookie = &http.Cookie{
			Name:   "session",
			Value:  "",
			MaxAge: -1,
		}
		http.SetCookie(res, cookie)
	}
	http.Redirect(res, req, "/", http.StatusSeeOther)
}
