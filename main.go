package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var tpl *template.Template
var db *sql.DB
var err error

type Pokemon struct {
	ID       int
	Name     string
	Type     string
	Category string
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
	db, err = sql.Open("postgres", "postgres://postgres:password@localhost:5432/pokedex?sslmode=disable")
	if err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

func main() {
	http.HandleFunc("/pokedex", pokedex)
	http.HandleFunc("/pokedex/", singlePokemon)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/", index)
	http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
}

func index(res http.ResponseWriter, req *http.Request) {
	// if LoggedIn, redirects to My Pokedex, if not request to log in and offer sign up button
	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/pokedex", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(res, "index.gohtml", nil)
}

func pokedex(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT * FROM pokemons ORDER BY ID;")
		if err != nil {
			fmt.Println("caceta meu 1")
			http.Error(res, err.Error(), 500)
			return
		}
		defer rows.Close()
		pkmns := make([]Pokemon, 0)
		for rows.Next() {
			pkmn := Pokemon{}
			err := rows.Scan(&pkmn.ID, &pkmn.Name, &pkmn.Type, &pkmn.Category)
			if err != nil {
				panic(err)
			}
			pkmns = append(pkmns, pkmn)
		}
		if err = rows.Err(); err != nil {
			fmt.Println("caceta meu")
			panic(err)
		}
		res.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(res).Encode(pkmns)
		if err != nil {
			http.Error(res, http.StatusText(500), 500)
		}
		return
	case http.MethodPost:
		ID_int, ID_interr := strconv.Atoi(req.FormValue("ID"))
		new_pkmn := Pokemon{
			ID:       ID_int,
			Name:     req.FormValue("Name"),
			Type:     req.FormValue("Type"),
			Category: req.FormValue("Category"),
		}

		// multipart/form data curls also add a boundary string to the header, so if we dont do it like this we get an error
		if ct := req.Header.Get("content-type"); !strings.Contains(ct, "multipart/form-data") {
			res.WriteHeader(http.StatusUnsupportedMediaType)
			res.Write([]byte(fmt.Sprintf("need content-type 'multipart/form-data', but got '%s'", ct)))
			return
		}
		// validate form values
		if ID_interr != nil || new_pkmn.Name == "" || new_pkmn.Type == "" || new_pkmn.Category == "" {
			http.Error(res, http.StatusText(400), http.StatusBadRequest)
			return
		}
		q := `
			INSERT INTO POKEMONS(id, name, type, category)
			VALUES($1,$2,$3,$4);
			`
		_, err = db.Exec(q, new_pkmn.ID, new_pkmn.Name, new_pkmn.Type, new_pkmn.Category)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		} else {
			res.Write([]byte("New record added succesfully."))
			res.Write([]byte(fmt.Sprintf("ID: %d, Name: %s, Type: %s, Category: %s", new_pkmn.ID, new_pkmn.Name, new_pkmn.Type, new_pkmn.Category)))
		}
	}

}

func singlePokemon(res http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.String(), "/")
	if len(parts) != 3 {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	switch req.Method {
	case http.MethodGet:
		row := db.QueryRow("SELECT * FROM pokemons WHERE id = $1", parts[2])
		pkmn := Pokemon{}
		err = row.Scan(&pkmn.ID, &pkmn.Name, &pkmn.Type, &pkmn.Category)
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(res).Encode(pkmn)
		if err != nil {
			http.Error(res, http.StatusText(500), 500)
		}
		return
	case http.MethodPut:
		ID_int, ID_interr := strconv.Atoi(parts[2])

		// multipart/form data curls also add a boundary string to the header, so if we dont do it like this we get an error
		if ct := req.Header.Get("content-type"); !strings.Contains(ct, "multipart/form-data") {
			res.WriteHeader(http.StatusUnsupportedMediaType)
			res.Write([]byte(fmt.Sprintf("need content-type 'multipart/form-data', but got '%s'", ct)))
			return
		}
		new_pkmn := Pokemon{
			ID:       ID_int,
			Name:     req.FormValue("Name"),
			Type:     req.FormValue("Type"),
			Category: req.FormValue("Category"),
		}

		// validate form values
		if ID_interr != nil || new_pkmn.Name == "" || new_pkmn.Type == "" || new_pkmn.Category == "" {
			http.Error(res, http.StatusText(400), http.StatusBadRequest)
			return
		}
		q := `
			UPDATE pokemons SET Name=$1, Type=$2, Category=$3 WHERE ID=$4;
			`
		result, err := db.Exec(q, new_pkmn.Name, new_pkmn.Type, new_pkmn.Category, new_pkmn.ID)
		if err != nil {
			panic(err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			panic(err)
		} else if rows > 0 {
			res.Write([]byte("Row updated successfully."))
		} else {
			res.Write([]byte("No rows were affected."))
		}
		return
	case http.MethodDelete:
		q := `
			DELETE FROM pokemons
			WHERE ID = $1;
		`
		result, err := db.Exec(q, parts[2])
		if err != nil {
			panic(err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			panic(err)
		} else if rows > 0 {
			res.Write([]byte("Pokemon ID:" + parts[2] + " Deleted Sucessfully"))
		} else {
			res.Write([]byte("No rows were affected. Id doesn't exist."))
		}
		return
	default:
		http.Error(res, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

}
