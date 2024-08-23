package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	m "oitm/middleware"
	"testing"
)

func TestGetLinks(t *testing.T) {

	test_get_links_requests := []struct {
		Params map[string]string
		Page int
		Valid bool
	}{
		{
			Params: map[string]string{
				"cats": "",
				"period": "",
				"req_user_id": "",
				"req_login_name": "",
			},
			Page: 0,
			Valid: true,
		},
		{
			Params: map[string]string{
				"cats": "",
				"period": "",
				"req_user_id": "",
				"req_login_name": "",
			},
			Page: 1,
			Valid: true,
		},
		{
			Params: map[string]string{
				"cats": "umvc3",
				"period": "",
				"req_user_id": "",
				"req_login_name": "",
			},
			Page: 1,
			Valid: true,
		},
		{
			Params: map[string]string{
				"cats": "umvc3",
				"period": "day",
				"req_user_id": "",
				"req_login_name": "",
			},
			Page: 1,
			Valid: true,
		},
		{
			Params: map[string]string{
				"cats": "umvc3",
				"period": "poop",
				"req_user_id": "",
				"req_login_name": "",
			},
			Page: 1,
			Valid: false,
		},
		{
			Params: map[string]string{
				"cats": "",
				"period": "",
				"req_user_id": "3",
				"req_login_name": "goolian",
			},
			Page: 1,
			Valid: true,
		},
		// passes because middlware corrects negative pages to 1
		{
			Params: map[string]string{
				"cats": "",
				"period": "",
				"req_user_id": "",
				"req_login_name": "",
			},
			Page: -1,
			Valid: true,
		},
	}


	for _, tglr := range test_get_links_requests {
		r := httptest.NewRequest(
			http.MethodGet,
			"/links/top",
			nil,
		)

		ctx := context.Background()
		ctx = context.WithValue(ctx, m.PageKey, tglr.Page)
		ctx = context.WithValue(ctx, m.UserIDKey, tglr.Params["req_user_id"])
		ctx = context.WithValue(ctx, m.LoginNameKey, tglr.Params["req_login_name"])
		r = r.WithContext(ctx)
		
		q := r.URL.Query()
		for k, v := range tglr.Params {
			q.Add(k, v)
		}
		r.URL.RawQuery = q.Encode()

		w := httptest.NewRecorder()
		GetLinks(w, r)
		res := w.Result()
		defer res.Body.Close()

		if tglr.Valid && res.StatusCode != http.StatusOK {
			text, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal("failed but unable to read request body bytes")
			}

			t.Fatalf(
				"expected status code 200, got %d (test request %+v)\n%s", res.StatusCode,
				tglr.Params,
				text,
			)
		} else if !tglr.Valid && res.StatusCode != http.StatusBadRequest {
			t.Errorf(
				"expected Bad Request, got %d (test request %+v)", 
				res.StatusCode, 
				tglr.Params,
			)
		}
	}
}

func TestAddLink(t *testing.T) {
	test_link_requests := []struct {
		Payload map[string]string
		Valid   bool
	}{
		{
			Payload: map[string]string{
				"url":     "",
				"cats":    "test",
				"summary": "test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"url":     "https://www.google.com/wholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextrachars",
				"cats":    "test",
				"summary": "test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"url":     "notreal",
				"cats":    "test",
				"summary": "bob",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"url":     "https://www.google.com",
				"cats":    "",
				"summary": "",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"url":     "https://www.google.com",
				"cats":    "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
				"summary": "",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"url":     "https://www.google.com",
				"cats":    "0,1,2,3,4,5,6,7,8,9,0,1,2",
				"summary": "",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"url":     "https://www.google.com",
				"cats":    "testtest",
				"summary": "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			},
			Valid: false,
		},
		{
			Payload: map[string]string{
				"url":     "google.com",
				"cats":    "test",
				"summary": "test",
			},
			Valid: true,
		},
		{
			Payload: map[string]string{
				"url":     "about.google.com",
				"cats":    "test",
				"summary": "testy",
			},
			Valid: true,
		},
		{
			Payload: map[string]string{
				"url":     "https://www.google.com/search/howsearchworks/?fg=1",
				"cats":    "test",
				"summary": "testiest",
			},
			Valid: true,
		},
		{
			Payload: map[string]string{
				"url":     "https://www.google.com/search/howsearchworks/features/",
				"cats":    "test",
				"summary": "",
			},
			Valid: true,
		},

		// should fail due to duplicate from previous test with url "google.com"
		{
			Payload: map[string]string{
				"url":     "https://www.google.com",
				"cats":    "test",
				"summary": "",
			},
			Valid: false,
		},
	}

	for _, tr := range test_link_requests {
		pl, _ := json.Marshal(tr.Payload)
		r := httptest.NewRequest(
			http.MethodPost,
			"/links",
			bytes.NewReader(pl),
		)
		r.Header.Set("Content-Type", "application/json")

		ctx := context.Background()
		ctx = context.WithValue(ctx, m.UserIDKey, test_user_id)
		ctx = context.WithValue(ctx, m.LoginNameKey, test_login_name)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()
		AddLink(w, r)
		res := w.Result()
		defer res.Body.Close()

		if tr.Valid && res.StatusCode != 201 {
			text, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal("failed but unable to read request body bytes")
			}

			t.Fatalf(
				"expected status code 201, got %d (test request %+v)\n%s", res.StatusCode,
				tr.Payload,
				text,
			)
		} else if !tr.Valid && res.StatusCode != 400 {
			t.Fatalf("expected status code 400, got %d", res.StatusCode)
		}
	}
}
