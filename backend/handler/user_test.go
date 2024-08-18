package handler

import (
	"bytes"
	"context"
	"encoding/json"
	m "oitm/middleware"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignUp(t *testing.T) {
	test_signup_requests := []struct {
		Payload map[string]string
		Valid   bool
	}{
		{
			Payload: map[string]string{
				"login_name": "",
				"password":   "testtest",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"login_name": "p",
				"password":   "testtest",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"login_name": "123456789012345678901234567890123",
				"password":   "testtest",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"login_name": "test",
				"password":   "",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"login_name": "test",
				"password":   "pp",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"login_name": "test",
				"password":   "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"login_name": "goolian",
				"password":   "testtest",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"login_name": "test",
				"password":   "testtest",
			},
			Valid: true,
		},
	}

	for _, tr := range test_signup_requests {
		pl, _ := json.Marshal(tr.Payload)
		r := httptest.NewRequest(
			http.MethodPost,
			"/signup",
			bytes.NewReader(pl),
		)
		r.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		SignUp(w, r)
		res := w.Result()
		defer res.Body.Close()

		if tr.Valid && res.StatusCode != 201 {
			text, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal("failed but unable to read request body bytes")
			} else {
				t.Fatalf(
					"expected status code 201, got %d (test request %+v)\n%s", res.StatusCode,
					tr.Payload,
					text,
				)
			}
		} else if !tr.Valid && res.StatusCode != 400 {
			t.Fatalf("expected status code 400, got %d", res.StatusCode)
		}
	}
}

func TestEditAbout(t *testing.T) {
	test_edit_about_requests := []struct {
		Payload map[string]string
		Valid   bool
	}{
		{
			Payload: map[string]string{
				"about": "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"about": "hello",
			},
			Valid: true,
		},
		{
			Payload: map[string]string{
				"about": "",
			},
			Valid: true,
		},
	}

	for _, tr := range test_edit_about_requests {
		pl, _ := json.Marshal(tr.Payload)
		r := httptest.NewRequest(
			http.MethodPut,
			"/users/about",
			bytes.NewReader(pl),
		)
		r.Header.Set("Content-Type", "application/json")

		ctx := context.Background()
		ctx = context.WithValue(ctx, m.UserIDKey, test_user_id)
		ctx = context.WithValue(ctx, m.LoginNameKey, test_login_name)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()
		EditAbout(w, r)
		res := w.Result()
		defer res.Body.Close()

		if tr.Valid && res.StatusCode != 200 {
			text, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal("failed but unable to read request body bytes")
			} else {
				t.Fatalf(
					"expected status code 200, got %d (test request %+v)\n%s", res.StatusCode,
					tr.Payload,
					text,
				)
			}
		} else if !tr.Valid && res.StatusCode != 400 {
			t.Fatalf("expected status code 400, got %d (test request %+v)", res.StatusCode, tr.Payload)
		}
	}
}
