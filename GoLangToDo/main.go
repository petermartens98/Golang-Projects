package main

import (
	"html/template"
	"log"
	"net/http"
)

type Todo struct {
	Item string
	Done bool
}

type PageData struct {
	Title string
	Todos []Todo
}

var todos []Todo

func todoHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title: "TODO List",
			Todos: todos,
		}
		tmpl.Execute(w, data)
	}
}

func addTodoHandler(w http.ResponseWriter, r *http.Request) {
	item := r.FormValue("item")
	if item != "" {
		todos = append(todos, Todo{Item: item, Done: false})
		log.Printf("Added new item to TODO list: %s", item)
	}
	http.Redirect(w, r, "/todo", http.StatusFound)
}

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.gohtml"))
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/todo", todoHandler(tmpl))
	mux.HandleFunc("/add", addTodoHandler)
	log.Fatal(http.ListenAndServe(":9091", mux))
}
