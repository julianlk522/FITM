package main

import (
	"log"
	"net/http"
	"os"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"

	h "github.com/julianlk522/fitm/handler"
	m "github.com/julianlk522/fitm/middleware"
)

var (
	token_auth *jwtauth.JWTAuth
	// test_api_url = "localhost:1999"
	api_url = "api.fitm.online:1999"
)

func init() {
	token_auth = jwtauth.New("HS256", []byte(os.Getenv("FITM_JWT_SECRET")), nil, jwt.WithAcceptableSkew(6*time.Hour))
}

func main() {
	r := chi.NewRouter()
	defer func() {
		if err := http.ListenAndServeTLS(
		api_url, 
			"/etc/letsencrypt/live/api.fitm.online/fullchain.pem", 
			"/etc/letsencrypt/live/api.fitm.online/privkey.pem", 
			r,
		); err != nil {
			log.Fatal(err)
		}
		// if err := http.ListenAndServe(test_api_url, r); err != nil {
		// 	log.Fatal(err)
		// }
	}()

	// Router-wide middleware
	// Logger
	r.Use(middleware.Logger)

	// Rate Limit
	httprate_options := httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message": "too many requests"}`))
	})
	r.Use(httprate.Limit(
		60,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	))
	r.Use(
		httprate.Limit(
			60,
			time.Minute,
			httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
				return r.Header.Get("Authorization"), nil
			}),
			httprate_options,
		))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://fitm.online", 
			"https://fitm.online/*",
			"https://www.fitm.online",
			"https://www.fitm.online/*",
			// "http://localhost:4321",
			// "http://localhost:4321/*",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		// Debug: true,
	}))

	// ROUTES
	// PUBLIC
	r.Post("/ghwh", h.HandleGitHubWebhook)
	
	r.Post("/signup", h.SignUp)
	r.Post("/login", h.LogIn)
	r.Get("/pic/{file_name}", h.GetProfilePic)

	r.Get("/cats", h.GetTopGlobalCats) // includes subcats
	r.Get("/contributors", h.GetTopContributors)

	// OPTIONAL AUTHENTICATION
	// (bearer token used optionally to get IsLiked / IsCopied for links)
	r.Group(func(r chi.Router) {
		r.Use(m.VerifierOptional(token_auth))
		r.Use(m.AuthenticatorOptional(token_auth))
		r.Use(m.JWT)

		r.Get("/map/{login_name}", h.GetTreasureMap)

		r.
			With(m.Pagination).
			Get("/links", h.GetLinks)

		r.Get("/summaries/{link_id}", h.GetSummaryPage)
		r.Get("/tags/{link_id}", h.GetTagPage)
	})

	// PROTECTED
	// (bearer token required)
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(token_auth))
		r.Use(jwtauth.Authenticator(token_auth))
		r.Use(m.JWT)

		// Users
		r.Put("/about", h.EditAbout)
		r.Post("/pic", h.UploadProfilePic)

		// Links
		r.Post("/links", h.AddLink)
		r.Delete("/links", h.DeleteLink)
		r.Post("/links/{link_id}/like", h.LikeLink)
		r.Delete("/links/{link_id}/like", h.UnlikeLink)
		r.Post("/links/{link_id}/copy", h.CopyLink)
		r.Delete("/links/{link_id}/copy", h.UncopyLink)

		// Tags
		r.Post("/tags", h.AddTag)
		r.Put("/tags", h.EditTag)

		// Summaries
		r.Post("/summaries", h.AddSummary)
		r.Delete("/summaries", h.DeleteSummary)
		r.Post("/summaries/{summary_id}/like", h.LikeSummary)
		r.Delete("/summaries/{summary_id}/like", h.UnlikeSummary)
	})
}