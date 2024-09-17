package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var templates *template.Template

type Blog struct {
	Id            int           `json:"id"`
	OwnerUsername string        `json:"owner_username"`
	Title         string        `json:"title"`
	Content       template.HTML `json:"content"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type User struct {
	Username     int    `json:"username"`
	PasswordHash int    `json:"password_hash"`
	Role         string `json:"role"`
}

type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmpassword"`
}

func serverRequest(url string, action string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Server-Action", action)

	client := http.Client{}

	return client.Do(req)
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

	http.HandleFunc("/", page404)

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("http://localhost:8080/v1/blogs")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var blogs []Blog

		if err := json.NewDecoder(resp.Body).Decode(&blogs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.ExecuteTemplate(w, "home.html", map[string]any{
			"Blogs": blogs,
		})
	})

	http.HandleFunc("/blogs/{$}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("http://localhost:8080/v1/blogs")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var blogs []Blog

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
			page404(w, r)
			return
		}

		var blog Blog
		if err = json.NewDecoder(resp.Body).Decode(&blog); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.ExecuteTemplate(w, "blog.html", blog)
	})

	http.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "register.html", nil)
	})

	http.HandleFunc("POST /register", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/register", http.StatusMovedPermanently)
			return
		}

		if !(r.Form.Has("username") && r.Form.Has("password") && r.Form.Has("confirmpassword")) {
			http.Error(w, "bad request", 400)
			return
		}

		rr := RegisterRequest{
			Username:        r.FormValue("username"),
			Password:        r.FormValue("password"),
			ConfirmPassword: r.FormValue("confirmpassword"),
		}

		resp, err := serverRequest("http://localhost:8080/v1/users",
			"GetByUsername",
			strings.NewReader(rr.Username),
		)

		if err != nil {
			log.Println("err: " + err.Error())
			http.Error(w, "internal server error", 500)
			return
		}

		if resp.StatusCode <= 200 || resp.StatusCode > 299 {
			// Request body is not needed, we are just checking
			// if the user already exists
			resp.Body.Close()
		}

		if resp.StatusCode != 404 {
			log.Println("err: " + err.Error())
			http.Error(w, "user already exists", 400)
			return
		}

		// TODO: Make POST request to http://localhost:8080/v1/users with registration details
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("app listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func page404(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "404.html", nil)
}
