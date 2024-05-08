package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	_ "github.com/mattn/go-sqlite3"

	"oitm/handler"
)

var token_auth *jwtauth.JWTAuth

func init() {
	// new JWT for protected routes
	token_auth = jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(30*time.Second))
}

func main() {	
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Home - check if server running
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})

	// USER ACCOUNTS
	r.Post("/signup", handler.SignUp)
	r.Post("/login", handler.LogIn)
	r.Patch("/users", handler.EditProfile)
	r.Get("/map/{login_name}", handler.GetTreasureMap)

	// PROTECTED (just testing for now)
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(token_auth))
		r.Use(jwtauth.Authenticator(token_auth))
		r.Get("/protected", handler.ProtectedArea)
	})

	// LINKS
	r.Get("/links", handler.GetTopLinks)
	r.Get("/links/{period}", handler.GetTopLinksByPeriod)
	r.Get("/links/cat/{categories}", handler.GetTopLinksByCategories)
	r.Get("/links/cat/{categories}/users", handler.GetTopCategoryContributors)
	r.Get("/links/subcat/{categories}", handler.GetTopSubcategories)
	r.Post("/links", handler.AddLink)
	r.Post("/links/copy", handler.CopyLinkToMap)
	r.Delete("/links/copy", handler.UncopyLink)
	r.Get("/links/{id}/likes", handler.GetLinkLikes)

	// TAGS
	r.Get("/tags/popular", handler.GetTopTagCategories)
	r.Post("/tags", handler.AddTag)
	r.Put("/tags", handler.EditTag)

	// SUMMARIES
	r.Post("/summaries", handler.AddSummaryOrSummaryLike)
	r.Patch("/summaries", handler.EditSummary)
	r.Delete("/summaries", handler.DeleteOrUnlikeSummary)

	// Serve
	// make sure this runs *after* all routes
	if err := http.ListenAndServe("localhost:8000", r); err != nil {
		log.Fatal(err)
	}
}

