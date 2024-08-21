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

	h "oitm/handler"
	m "oitm/middleware"
)

var token_auth *jwtauth.JWTAuth

func init() {
	token_auth = jwtauth.New("HS256", []byte(os.Getenv("FITM_JWT_SECRET")), nil, jwt.WithAcceptableSkew(6*time.Hour))
}

func main() {
	r := chi.NewRouter()
	defer func() {
		if err := http.ListenAndServe("localhost:8000", r); err != nil {
			log.Fatal(err)
		}
	}()

	// Router-wide middleware
	httprate_options := httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message": "too many requests"}`))
	})

	r.Use(middleware.Logger)
	r.Use(httprate.Limit(
		20,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	))
	r.Use(
		httprate.Limit(
			20,
			time.Minute,
			httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
				return r.Header.Get("Authorization"), nil
			}),
			httprate_options,
		))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		MaxAge: 300,
		Debug: true,
	}))

	// ROUTES
	// PUBLIC
	// Users
	r.Post("/signup", h.SignUp)
	r.Post("/login", h.LogIn)

	r.Get("/pic/{file_name}", h.GetProfilePic)

	// Cats
	r.Get("/cats", h.GetTopGlobalCats)
	r.Get("/subcats/{cats}", h.GetSubcats)
	r.Get("/contributors/{cats}", h.GetCatsContributors)

	// OPTIONAL AUTHENTICATION
	// (bearer token used optionally to get IsLiked / IsCopied / IsTagged for links)
	r.Group(func(r chi.Router) {
		r.Use(m.VerifierOptional(token_auth))
		r.Use(m.AuthenticatorOptional(token_auth))
		r.Use(m.JWT)

		r.Get("/map/{login_name}", h.GetTreasureMap)

		r.Route("/links", func(r chi.Router) {
			r.Use(m.Pagination)
			r.Get("/top", h.GetLinks)
		})

		r.Get("/summaries/{link_id}", h.GetSummaryPage)
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
		r.Post("/links/{link_id}/like", h.LikeLink)
		r.Delete("/links/{link_id}/like", h.UnlikeLink)
		r.Post("/links/{link_id}/copy", h.CopyLink)
		r.Delete("/links/{link_id}/copy", h.UncopyLink)

		// Tags
		r.Get("/tags/{link_id}", h.GetTagPage)
		r.Post("/tags", h.AddTag)
		r.Put("/tags", h.EditTag)

		// Summaries
		r.Post("/summaries", h.AddSummary)
		r.Delete("/summaries", h.DeleteSummary)
		r.Post("/summaries/{summary_id}/like", h.LikeSummary)
		r.Delete("/summaries/{summary_id}/like", h.UnlikeSummary)
	})
}
