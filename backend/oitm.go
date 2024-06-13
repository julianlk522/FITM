package main

import (
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	_ "github.com/mattn/go-sqlite3"

	"oitm/handler"
)

var token_auth *jwtauth.JWTAuth

func init() {
	// new JWT for protected routes (1-day exiration)
	token_auth = jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(24*time.Hour))
	
}

func main() {	
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{
			// "Accept", 
			"Authorization", 
			"Content-Type", 
			// "X-CSRF-Token",
		},
		MaxAge: 300, // Maximum value not ignored by any of major browsers
	  }))

	// Home - check if server running
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})

	// PUBLIC
	// USER ACCOUNTS
	r.Get("/users/{login_name}", handler.GetProfile)
	r.Get("/pic/{file_name}", handler.GetProfilePic)
	r.Post("/signup", handler.SignUp)
	r.Post("/login", handler.LogIn)

	// LINKS
	r.Get("/links/cat/{categories}/users", handler.GetTopCategoryContributors)
	r.Get("/links/subcat/{categories}", handler.GetTopSubcategories)
	r.Get("/links/{id}/likes", handler.GetLinkLikes)

	// TAGS
	r.Get("/tags/popular", handler.GetTopTagCategories)

	// OPTIONAL AUTHENTICATION
	// (bearer token optional; used to get is_liked property for links)
	r.Group(func(r chi.Router) {
		r.Use(handler.VerifierOptional(token_auth))
		r.Use(handler.AuthenticatorOptional(token_auth))

		// USER ACCOUNTS
		r.Get("/map/{login_name}", handler.GetTreasureMap)
		r.Get("/map/{login_name}/{categories}", handler.GetTreasureMapByCategories)

		// LINKS
		r.Get("/links", handler.GetTopLinks)
		r.Get("/links/{period}", handler.GetTopLinksByPeriod)
		r.Get("/links/cat/{categories}", handler.GetTopLinksByCategories)	

		// SUMMARIES
		r.Get("/summaries/{link_id}", handler.GetSummariesForLink)
	})

	// PROTECTED
	// (bearer token required)
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(token_auth))
		r.Use(jwtauth.Authenticator(token_auth))

		// USER ACCOUNTS
		r.Patch("/users", handler.EditProfile)
		r.Post("/pic", handler.UploadProfilePic)

		// LINKS
		r.Post("/links", handler.AddLink)
		r.Post("/links/{link_id}/like", handler.LikeLink)
		r.Delete("/links/{link_id}/like", handler.UnlikeLink)
		r.Post("/links/{link_id}/copy", handler.CopyLink)
		r.Delete("/links/{link_id}/copy", handler.UncopyLink)

		// TAGS
		r.Post("/tags", handler.AddTag)
		r.Put("/tags", handler.EditTag)

		// SUMMARIES
		r.Post("/summaries", handler.AddSummary)
		// r.Put("/summaries", handler.EditSummary)
		r.Delete("/summaries", handler.DeleteSummary)
		r.Post("/summaries/{summary_id}/like", handler.LikeSummary)
		r.Delete("/summaries/{summary_id}/like", handler.UnlikeSummary)

	})

	// Serve
	// make sure this runs after all routes
	if err := http.ListenAndServe("localhost:8000", r); err != nil {
		log.Fatal(err)
	}
}
