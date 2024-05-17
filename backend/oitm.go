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
	// new JWT for protected routes (1 day)
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
		// ExposedHeaders:   []string{"Link"},
		// AllowCredentials: false,
		MaxAge: 300, // Maximum value not ignored by any of major browsers
	  }))

	// Home - check if server running
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})

	// PUBLIC
	// USER ACCOUNTS
	r.Get("/users/{login_name}", handler.GetProfile)
	r.Get("/map/{login_name}", handler.GetTreasureMap)
	r.Post("/signup", handler.SignUp)
	r.Post("/login", handler.LogIn)

	// LINKS
	r.Get("/links/cat/{categories}/users", handler.GetTopCategoryContributors)
	r.Get("/links/subcat/{categories}", handler.GetTopSubcategories)
	r.Get("/links/{id}/likes", handler.GetLinkLikes)

	// TAGS
	r.Get("/tags/popular", handler.GetTopTagCategories)
	
	// SUMMARIES
	r.Get("/summaries/{link_id}", handler.GetSummariesForLink)

	// OPTIONAL AUTHENTICATION
	// e.g., links include is_liked property
	r.Group(func(r chi.Router) {
		r.Use(VerifierOptional(token_auth))
		r.Use(AuthenticatorOptional(token_auth))

		r.Get("/links", handler.GetTopLinks)
		r.Get("/links/{period}", handler.GetTopLinksByPeriod)
		r.Get("/links/cat/{categories}", handler.GetTopLinksByCategories)	

	})

	// PROTECTED
	// (bearer token required)
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(token_auth))
		r.Use(jwtauth.Authenticator(token_auth))

		// USER ACCOUNTS
		r.Patch("/users", handler.EditProfile)

		// LINKS
		r.Post("/links", handler.AddLink)
		r.Post("/links/{link_id}/like", handler.LikeLink)
		r.Delete("/links/{link_id}/like", handler.UnlikeLink)
		r.Post("/links/copy", handler.CopyLinkToMap)
		r.Delete("/links/copy", handler.UncopyLink)

		// TAGS
		r.Post("/tags", handler.AddTag)
		r.Put("/tags", handler.EditTag)

		// SUMMARIES
		r.Post("/summaries", handler.AddSummaryOrSummaryLike)
		r.Put("/summaries", handler.EditSummary)
		r.Delete("/summaries", handler.DeleteOrUnlikeSummary)

	})

	// Serve
	// make sure this runs after all routes
	if err := http.ListenAndServe("localhost:8000", r); err != nil {
		log.Fatal(err)
	}
}

// MODIFIED JWT VERIFIER
// (requests with no token allowed, but getting isLiked on links requires a token)

// From "github.com/go-chi/jwtauth/v5":
// Verifier http middleware handler will verify a JWT string from a http request.
//
// Verifier will search for a JWT token in a http request, in the order:
//  1. 'Authorization: BEARER T' request header
//  2. Cookie 'jwt' value
//
// The first JWT string that is found as a query parameter, authorization header
// or cookie header is then decoded by the `jwt-go` library and a *jwt.Token
// object is set on the request context. In the case of a signature decoding error
// the Verifier will also set the error on the request context.
//
// The Verifier always calls the next http handler in sequence, which can either
// be the generic `jwtauth.Authenticator` middleware or your own custom handler
// which checks the request context jwt token and error to prepare a custom
// http response.
func VerifierOptional(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return VerifyOptional(ja, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie)
}

func VerifyOptional(ja *jwtauth.JWTAuth, findTokenFns ...func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			token, err := VerifyRequestOptional(ja, r, findTokenFns...)
			ctx = jwtauth.NewContext(ctx, token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}

func VerifyRequestOptional(ja *jwtauth.JWTAuth, r *http.Request, findTokenFns ...func(r *http.Request) string) (jwt.Token, error) {
	var tokenString string

	// Extract token string from the request by calling token find functions in
	// the order they where provided. Further extraction stops if a function
	// returns a non-empty string.
	for _, fn := range findTokenFns {
		tokenString = fn(r)
		if tokenString != "" {
			break
		}
	}
	if tokenString == "" {
		// return nil, jwtauth.ErrNoTokenFound
		return nil, nil
	}

	return jwtauth.VerifyToken(ja, tokenString)
}

// MODIFIED JWT AUTHENTICATOR
// (requests with no token allowed, but getting isLiked on links requires a token)

// From "github.com/go-chi/jwtauth/v5":
// Authenticator is a default authentication middleware to enforce access from the
// Verifier middleware request context values. The Authenticator sends a 401 Unauthorized
// response for any unverified tokens and passes the good ones through. It's just fine
// until you decide to write something similar and customize your client response.
func AuthenticatorOptional(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			} else if token != nil && jwt.Validate(token, ja.ValidateOptions()...) != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// No token or valid token, either way pass through
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
