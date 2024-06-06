package handler

import (
	"net/http"

	"errors"

	"github.com/go-chi/jwtauth/v5"
)

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