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

	h "oitm/handler"
	m "oitm/middleware"
)

var token_auth *jwtauth.JWTAuth

func init() {
	// new JWT for protected/optional routes (1-day exiration)
	// TODO: shorten expiration to idk, 6h
	token_auth = jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(24*time.Hour))
}

func main() {
	r := chi.NewRouter()
	defer func() {
		if err := http.ListenAndServe("localhost:8000", r); err != nil {
			log.Fatal(err)
		}
	}()

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
	r.Get("/pic/{file_name}", h.GetProfilePic)
	r.Post("/signup", h.SignUp)
	r.Post("/login", h.LogIn)

	// LINKS
	r.Get("/links/cat/{categories}/users", h.GetTopCategoryContributors)
	r.Get("/links/{period}/cat/{categories}/users", h.GetTopCategoryContributorsByPeriod)
	r.Get("/links/subcat/{categories}", h.GetSubcategories)
	r.Get("/links/{period}/subcat/{categories}", h.GetSubcategoriesByPeriod)

	// TAGS
	r.Get("/tags/popular", h.GetTopTagCategories)
	r.Get("/tags/popular/{period}", h.GetTopTagCategoriesByPeriod)



	// OPTIONAL AUTHENTICATION
	// (bearer token used optionally to get IsLiked / IsCopied / IsTagged for links)
	r.Group(func(r chi.Router) {
		r.Use(m.VerifierOptional(token_auth))
		r.Use(m.AuthenticatorOptional(token_auth))
		r.Use(m.JWT)

		// USER ACCOUNTS
		r.Get("/map/{login_name}", h.GetTreasureMap)
		r.Get("/map/{login_name}/{categories}", h.GetTreasureMapByCategories)

		// LINKS
		r.Route("/links", func(r chi.Router) {
			r.Use(m.Pagination)
			r.Get("/", h.GetTopLinks)
			r.Get("/{period}", h.GetTopLinksByPeriod)
			r.Get("/cat/{categories}", h.GetTopLinksByCategories)
			r.Get("/{period}/{categories}", h.GetTopLinksByPeriodAndCategories)	
		})

		// SUMMARIES
		r.Get("/summaries/{link_id}", h.GetSummariesForLink)
	})



	// PROTECTED
	// (bearer token required)
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(token_auth))
		r.Use(jwtauth.Authenticator(token_auth))
		r.Use(m.JWT)

		// USER ACCOUNTS
		r.Put("/users/about", h.EditAbout)
		r.Post("/pic", h.UploadNewProfilePic)

		// LINKS
		r.Post("/links", h.AddLink)
		r.Post("/links/{link_id}/like", h.LikeLink)
		r.Delete("/links/{link_id}/like", h.UnlikeLink)
		r.Post("/links/{link_id}/copy", h.CopyLink)
		r.Delete("/links/{link_id}/copy", h.UncopyLink)

		// TAGS
		r.Get("/tags/{link_id}", h.GetTagsForLink)
		r.Post("/tags", h.AddTag)
		r.Put("/tags", h.EditTag)

		// SUMMARIES
		r.Post("/summaries", h.AddSummary)
		r.Delete("/summaries", h.DeleteSummary)
		r.Post("/summaries/{summary_id}/like", h.LikeSummary)
		r.Delete("/summaries/{summary_id}/like", h.UnlikeSummary)

	})
}
