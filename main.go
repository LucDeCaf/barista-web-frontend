package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

const RECAPTCHA_PROJECT_ID = "barista-1726578784920"
const RECAPTCHA_KEY = "6LcH9kYqAAAAAA13KUvAvKBJmcG9g9xpSsSMFaUz"

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
	Username     string `json:"username"`
	PasswordHash int    `json:"password_hash"`
	Role         string `json:"role"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func generateRiskAnalysis(projectID string, recaptchaKey string, token string, action string) (*recaptchapb.RiskAnalysis, error) {
	ctx := context.Background()
	client, err := recaptcha.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	event := &recaptchapb.Event{
		Token:   token,
		SiteKey: recaptchaKey,
	}

	assessment := &recaptchapb.Assessment{
		Event: event,
	}

	request := &recaptchapb.CreateAssessmentRequest{
		Assessment: assessment,
		Parent:     fmt.Sprintf("projects/%s", projectID),
	}

	resp, err := client.CreateAssessment(
		ctx,
		request,
	)
	if err != nil {
		return nil, err
	}

	if !resp.TokenProperties.Valid {
		return nil, fmt.Errorf("invalid token: %v", token)
	}

	if resp.TokenProperties.Action != action {
		return nil, fmt.Errorf("unexpected action %v, expected %v", resp.TokenProperties.Action, action)
	}

	return resp.RiskAnalysis, nil
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))

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
		var rr RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&rr); err != nil {
			log.Println("err: " + err.Error())
			http.Error(w, "bad request", 400)
			return
		}

		assessment, err := generateRiskAnalysis(RECAPTCHA_PROJECT_ID, RECAPTCHA_KEY, rr.Token, "register")
		if err != nil {
			log.Println("err: " + err.Error())
			http.Error(w, "internal server error", 500)
			return
		}

		if assessment.Score < 0.4 {
			log.Printf("Potential bot found (score: %v)\n", assessment.Score)
			http.Error(w, "disallowed", 402) // TODO: Find proper method
			return
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

		// TODO: Test this / finish this
		resp, err = http.Post(
			"http://localhost:8080/register",
			"application/json",
			strings.NewReader(fmt.Sprintf("{username:\"%v\",password:\"%v\"}", rr.Username, rr.Password)),
		)
		if err != nil {
			log.Println("err: ", err.Error())
			http.Error(w, "internal server error", 500)
			return
		}
	})

	log.Println("app listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func page404(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "404.html", nil)
}
