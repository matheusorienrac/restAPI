package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"

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
	tpl = template.Must(template.ParseFiles("index.gohtml"))
	postgreConnString, err := ioutil.ReadFile("postgreConnString")
	if err != nil {
		panic(err)
	}
	db, err = sql.Open("postgres", string(postgreConnString))
	if err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

func main() {
	http.HandleFunc("/pokemons/create", create)
	http.HandleFunc("/pokemons/read", read)
	http.HandleFunc("/pokemons/update", update)
	http.HandleFunc("/pokemons/delete", delete)
	http.HandleFunc("/pokedex", pokedex)
	http.HandleFunc("/", index)
	http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
}

func index(res http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(res, "index.gohtml", "Tem nada aqui kkkkkkkkkkkkkk")
}

func create(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	ID_int, ID_interr := strconv.Atoi(req.FormValue("ID"))

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

func read(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	pkmn_id := req.FormValue("ID")
	fmt.Println(pkmn_id)
	row := db.QueryRow("SELECT * FROM pokemons WHERE id = $1", pkmn_id)
	pkmn := Pokemon{}
	err = row.Scan(&pkmn.ID, &pkmn.Name, &pkmn.Type, &pkmn.Category)
	if err != nil {
		http.Error(res, http.StatusText(500), 500)
	}
	res.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(res).Encode(pkmn)
	if err != nil {
		http.Error(res, http.StatusText(500), 500)
	}
}

func update(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	ID_int, ID_interr := strconv.Atoi(req.FormValue("ID"))

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
		UPDATE pokemons SET ID = $1, Name=$2, Type=$3, Category=$4 WHERE ID=$1;
		`
	result, err := db.Exec(q, new_pkmn.ID, new_pkmn.Name, new_pkmn.Type, new_pkmn.Category)
	if err != nil {
		panic(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		panic(err)
	} else if rows > 0 {
		res.Write([]byte("Row updated successfully."))
		fmt.Println("redirected to /pokedex")
		http.Redirect(res, req, "/pokedex", http.StatusSeeOther)
	} else {
		res.Write([]byte("No rows were affected."))
	}
	return

}

func delete(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	idToDelete := req.FormValue("ID")
	q := `
		DELETE FROM pokemons
		WHERE ID = $1;
	`
	result, err := db.Exec(q, idToDelete)
	if err != nil {
		panic(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		panic(err)
	} else if rows > 0 {
		res.Write([]byte(fmt.Sprintf("Row with ID: %s removed successfully.", idToDelete)))
		http.Redirect(res, req, "/pokemons", http.StatusSeeOther)
	} else {
		res.Write([]byte("No rows were affected."))
	}

	return

}
func pokedex(res http.ResponseWriter, req *http.Request) {
	rows, err := db.Query("SELECT * FROM pokemons;")
	if err != nil {
		http.Error(res, http.StatusText(500), 500)
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
		panic(err)
	}

	res.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(res).Encode(pkmns)
	if err != nil {
		panic(err)
	}
	for _, pkmn := range pkmns {
		fmt.Println(pkmn.ID, pkmn.Name, pkmn.Type, pkmn.Category)
	}

}
