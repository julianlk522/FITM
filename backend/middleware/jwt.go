package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// MODIFIED JWT VERIFIER / AUTHENTICATOR
// (requests with no token are allowed, but getting isLiked / isCopied / isTagged on links requires a token)
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

	for _, fn := range findTokenFns {
		tokenString = fn(r)
		if tokenString != "" {
			break
		}
	}

	if tokenString == "" {
		return nil, nil
	}

	return jwtauth.VerifyToken(ja, tokenString)
}

func AuthenticatorOptional(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			// Error decoding token
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return

			// Invalid token
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

// Retrieve JWT claims if they are passed in request context
// claims = {"user_id":"1234","login_name":"johndoe"}
func JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user_id, login_name interface{}
		_, claims, err := jwtauth.FromContext(r.Context())
		if len(claims) == 0 || err != nil {
			user_id = ""
			login_name = ""
		} else {
			var ok, ok2 bool
			user_id, ok = claims["user_id"].(string)
			login_name, ok2 = claims["login_name"].(string)

			if !ok || !ok2 {
				user_id = ""
				login_name = ""
			}
		}
				
		ctx := context.WithValue(r.Context(), UserIDKey, user_id)
		ctx = context.WithValue(ctx, LoginNameKey, login_name)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}