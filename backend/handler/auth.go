package handler

import (
	"net/http"

	"errors"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// MODIFIED JWT VERIFIER
// (requests with no token are allowed, but getting isLiked / isCopied / isTagged on links requires a token)

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
// (requests with no token are allowed, but getting isLiked / isCopied / isTagged on links requires a token)

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

// Retrieve signed-in user login_name and user_id from JWT claims if they are passed in request context
func GetJWTClaims(r *http.Request) (map[string]interface{}, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if len(claims) == 0 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	
	// claims = {"user_id":"1234","login_name":"johndoe"}
	req_login_name, ok := claims["login_name"]
	req_user_id, ok2 := claims["user_id"]
	if !ok || !ok2 {
		return nil, errors.New("invalid auth token")
	}
	
	return map[string]interface{}{"login_name": req_login_name, "user_id": req_user_id}, nil
}