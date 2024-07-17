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
	r.Get("/users/{login_name}", handler.GetProfile)
	r.Get("/pic/{file_name}", handler.GetProfilePic)
	r.Post("/signup", handler.SignUp)
	r.Post("/login", handler.LogIn)

	// LINKS
	r.Get("/links/cat/{categories}/users", handler.GetTopCategoryContributors)
	r.Get("/links/{period}/cat/{categories}/users", handler.GetTopCategoryContributorsByPeriod)
	r.Get("/links/subcat/{categories}", handler.GetSubcategories)
	r.Get("/links/{period}/subcat/{categories}", handler.GetSubcategoriesByPeriod)

	// TAGS
	r.Get("/tags/popular", handler.GetTopTagCategories)
	r.Get("/tags/popular/{period}", handler.GetTopTagCategoriesByPeriod)



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
		r.Get("/links/{period}/{categories}", handler.GetTopLinksByPeriodAndCategories)	

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
		r.Put("/users/about", handler.EditAbout)
		r.Post("/pic", handler.UploadProfilePic)

		// LINKS
		r.Post("/links", handler.AddLink)
		r.Post("/links/{link_id}/like", handler.LikeLink)
		r.Delete("/links/{link_id}/like", handler.UnlikeLink)
		r.Post("/links/{link_id}/copy", handler.CopyLink)
		r.Delete("/links/{link_id}/copy", handler.UncopyLink)

		// TAGS
		r.Get("/tags/{link_id}", handler.GetTagsForLink)
		r.Post("/tags", handler.AddTag)
		r.Put("/tags", handler.EditTag)

		// SUMMARIES
		r.Post("/summaries", handler.AddSummary)
		r.Delete("/summaries", handler.DeleteSummary)
		r.Post("/summaries/{summary_id}/like", handler.LikeSummary)
		r.Delete("/summaries/{summary_id}/like", handler.UnlikeSummary)

	})
}
