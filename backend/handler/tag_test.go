package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	m "github.com/julianlk522/fitm/middleware"
	"testing"
)

func TestAddTag(t *testing.T) {
	test_tag_requests := []struct {
		Payload map[string]string
		Valid   bool
	}{
		{
			Payload: map[string]string{
				"link_id": "",
				"cats":    "test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"link_id": "-1",
				"cats":    "test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"link_id": "101010101010101010101010101010101010101",
				"cats":    "test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"link_id": "notanint",
				"cats":    "test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"link_id": "1",
				"cats":    "",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"link_id": "1",
				"cats":    "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123",
			},
			Valid: false,
		},
		// too many cats
		{
			Payload: map[string]string{
				"link_id": "1",
				"cats":    "0,1,2,3,4,5,6,7,8,9,0,1,2",
			},
			Valid: false,
		},
		// duplicate cats
		{
			Payload: map[string]string{
				"link_id": "1",
				"cats":    "0,1,2,3,3",
			},
			Valid: false,
		},
		// should fail because user goolian has already tagged link with ID 1
		{
			Payload: map[string]string{
				"link_id": "1",
				"cats":    "testtest",
			},
			Valid: false,
		},
		// should pass because goolian has _not_ tagged link with ID 10
		{
			Payload: map[string]string{
				"link_id": "10",
				"cats":    "testtest",
			},
			Valid: true,
		},
	}

	const (
		test_user_id    = "3"
		test_login_name = "goolian"
	)

	for _, tr := range test_tag_requests {
		pl, _ := json.Marshal(tr.Payload)
		r := httptest.NewRequest(
			http.MethodPost,
			"/tags",
			bytes.NewReader(pl),
		)
		r.Header.Set("Content-Type", "application/json")

		ctx := context.Background()
		ctx = context.WithValue(ctx, m.UserIDKey, test_user_id)
		ctx = context.WithValue(ctx, m.LoginNameKey, test_login_name)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()
		AddTag(w, r)
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

func TestEditTag(t *testing.T) {
	test_tag_requests := []struct {
		Payload map[string]string
		Valid   bool
	}{
		{
			Payload: map[string]string{
				"tag_id": "1",
				"cats":   "",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"tag_id": "1",
				"cats":   "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123",
			},
			Valid: false,
		},
		// too many cats
		{
			Payload: map[string]string{
				"tag_id": "1",
				"cats":   "0,1,2,3,4,5,6,7,8,9,0,1,2",
			},
			Valid: false,
		},
		// duplicate cats
		{
			Payload: map[string]string{
				"tag_id": "1",
				"cats":   "0,1,2,3,3",
			},
			Valid: false,
		},
		// should fail because user goolian _did not_ submit tag with ID 10
		{
			Payload: map[string]string{
				"tag_id": "10",
				"cats":   "testtest",
			},
			Valid: false,
		},
		// should pass because user goolian _has_ submitted tag with ID 32
		{
			Payload: map[string]string{
				"tag_id": "32",
				"cats":   "hello,kitty",
			},
			Valid: true,
		},
	}

	const (
		test_user_id    = "3"
		test_login_name = "goolian"
	)

	for _, tr := range test_tag_requests {
		pl, _ := json.Marshal(tr.Payload)
		r := httptest.NewRequest(
			http.MethodPut,
			"/tags",
			bytes.NewReader(pl),
		)
		r.Header.Set("Content-Type", "application/json")

		ctx := context.Background()
		ctx = context.WithValue(ctx, m.UserIDKey, test_user_id)
		ctx = context.WithValue(ctx, m.LoginNameKey, test_login_name)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()
		EditTag(w, r)
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
			t.Fatalf("expected status code 400, got %d", res.StatusCode)
		}
	}
}
