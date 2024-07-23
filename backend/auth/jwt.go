package auth

import (
	"net/http"

	"errors"

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

// Retrieve JWT claims if they are passed in request context
// claims = {"user_id":"1234","login_name":"johndoe"}
// TODO: remove and replace repeated logic if claims can simply be passed to request context and retrieved from it directly in handlers
func GetJWTClaims(r *http.Request) (string, string, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if len(claims) == 0 {
		return "", "", nil
	} else if err != nil {
		return "", "", err
	}
	
	req_user_id, ok := claims["user_id"]
	req_login_name, ok2 := claims["login_name"]
	if !ok || !ok2 {
		return "", "", errors.New("invalid auth token")
	}
	
	return req_user_id.(string), req_login_name.(string), nil
}