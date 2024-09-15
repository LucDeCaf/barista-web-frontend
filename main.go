package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"
)

var templates *template.Template

type blog struct {
	Id            int           `json:"id"`
	OwnerUsername string        `json:"owner_username"`
	Title         string        `json:"title"`
	Content       template.HTML `json:"content"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

func main() {
	portPtr := flag.String(
		"port",
		"8000",
		"the port the app will run on",
	)
	flag.Parse()
	port := *portPtr

	templates = template.Must(template.ParseGlob("templates/*.html"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			handle404(w)
			return
		}

		resp, err := http.Get("http://localhost:8080/v1/blogs")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var blogs []blog

		if err := json.NewDecoder(resp.Body).Decode(&blogs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.ExecuteTemplate(w, "index.html", map[string]any{
			"Blogs": blogs,
		})
	})

	http.HandleFunc("/blogs", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("http://localhost:8080/v1/blogs")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var blogs []blog

		if err := json.NewDecoder(resp.Body).Decode(&blogs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.ExecuteTemplate(w, "blogs.html", map[string]any{
			"Blogs": blogs,
		})
	})

	http.HandleFunc("/blogs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		resp, err := http.Get("http://localhost:8080/v1/blogs/" + id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			handle404(w)
			return
		}

		var blog blog
		if err = json.NewDecoder(resp.Body).Decode(&blog); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.ExecuteTemplate(w, "blog.html", blog)
	})

	log.Println("app listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handle404(w http.ResponseWriter) {
	templates.ExecuteTemplate(w, "404.html", nil)
}
