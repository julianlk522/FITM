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

func TestAddTag(t *testing.T) {
	test_tag_requests := []struct {
		Payload map[string]string
		Valid bool
	}{
		{
			Payload: map[string]string {
					"link_id":"",
					"categories":"test",
				},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"link_id":"-1",
				"categories":"test",
				},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"link_id":"101010101010101010101010101010101010101",
				"categories":"test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"link_id":"notanint",
				"categories":"test",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"link_id":"1",
				"categories":"",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"link_id":"1",
				"categories":"0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123",
			},
			Valid: false,
		},
		{
			Payload: map[string]string {
				"link_id":"1",
				"categories":"0,1,2,3,4,5,6,7,8,9,0,1,2",
			},
			Valid: false,
		},
		// should fail because user goolian has already tagged link with ID 1
		{
			Payload: map[string]string {
				"link_id":"1",
				"categories":"testtest",
			},
			Valid: false,
		},
		// should pass because goolian has _not_ tagged link with ID 10
		{
			Payload: map[string]string {
				"link_id":"10",
				"categories":"testtest",
			},
			Valid: true,
		},
	}

	const (
		test_user_id = "3"
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