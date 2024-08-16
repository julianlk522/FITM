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

func TestAddLink(t *testing.T) {
	test_link_requests := []struct {
		Payload map[string]string
		Valid bool
	}{
		{
			Payload: map[string]string {
					"url":"",
					"categories":"test",
					"summary":"test",
				},
			Valid: false,
		},
		{
			Payload: map[string]string {
					"url":"https://www.google.com/wholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextracharswholebunchofextrachars",
					"categories":"test",
					"summary":"test",
				},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"url":"notreal",
				"categories":"test",
				"summary":"bob",
				},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"url":"https://www.google.com",
				"categories":"",
				"summary":"",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"url":"https://www.google.com",
				"categories":"01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
				"summary":"",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"url":"https://www.google.com",
				"categories":"0,1,2,3,4,5,6,7,8,9,0,1,2",
				"summary":"",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"url":"https://www.google.com",
				"categories":"testtest",
				"summary":"01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
					"url":"google.com",
					"categories":"test",
					"summary":"test",
				},
			Valid: true,
		},
		{
			Payload: map[string]string {
					"url":"about.google.com",
					"categories":"test",
					"summary":"testy",
				},
			Valid: true,
		},
		{
			Payload: map[string]string {
					"url":"https://www.google.com/search/howsearchworks/?fg=1",
					"categories":"test",
					"summary":"testiest",
				},
			Valid: true,
		},
		{
			Payload: map[string]string {
					"url":"https://www.google.com/search/howsearchworks/features/",
					"categories":"test",
					"summary":"",
				},
			Valid: true,
		},

		// should fail due to duplicate from previous test with url "google.com"
		{
			Payload: map[string]string {
					"url":"https://www.google.com",
					"categories":"test",
					"summary":"",
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