package error

import (
	"errors"
	"fmt"

	"net/http"

	"github.com/go-chi/render"
)

var (
	ErrInvalidPage = errors.New("invalid page provided")
	ErrInvalidLinkID error = errors.New("invalid link ID provided")
	
	ErrNoLinkID error = errors.New("no link ID provided")
	ErrNoCats error = errors.New("no cats provided")
	ErrNoPeriod error = errors.New("no period provided")
	ErrNoURL error = errors.New("no URL provided")
	ErrNoSummaryID error = errors.New("no summary ID provided")
	ErrNoSummaryText error = errors.New("no summary text provided")
	ErrNoSummaryReplacementText error = errors.New("no summary replacement text provided")
	ErrNoTagID error = errors.New("no tag ID provided")
	ErrNoTagCats error = errors.New("no tag cat(s) provided")
	ErrNoLoginName error = errors.New("no login name provided")
	ErrNoPassword error = errors.New("no password provided")
	
	ErrNoLinkWithID error = errors.New("no link found with given ID")
	ErrNoSummaryWithID error = errors.New("no summary found with given ID")
	ErrNoTagWithID error = errors.New("no tag found with given ID")
	ErrNoUserWithLoginName error = errors.New("no user found with given login name")

	ErrDuplicateTag error = errors.New("duplicate tag")
	ErrDuplicateCats error = errors.New("tag contains duplicate cat(s)")
	ErrDoesntOwnTag error = errors.New("cannot edit another user's tag")

	ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
)

// User
func LoginNameExceedsLowerLimit(limit int) error {
	return fmt.Errorf("login name too short (min %d chars)", limit)
}

func LoginNameExceedsUpperLimit(limit int) error {
	return fmt.Errorf("login name too long (max %d chars)", limit)
}

func PasswordExceedsLowerLimit(limit int) error {
	return fmt.Errorf("password too short (min %d chars)", limit)
}

func PasswordExceedsUpperLimit(limit int) error {
	return fmt.Errorf("password too long (max %d chars)", limit)
}

func ProfileAboutLengthExceedsLimit(limit int) error {
	return fmt.Errorf("about text too long (max %d chars)", limit)
}

// Link
func LinkURLCharsExceedLimit(limit int) error {
	return fmt.Errorf("url too long (max %d chars)", limit)
}

// Summary
func SummaryLengthExceedsLimit(limit int) error {
	return fmt.Errorf("summary too long (max %d chars)", limit)
}

// Tag
func CatCharsExceedLimit(limit int) error {
	return fmt.Errorf("cat too long (max %d chars)", limit)
}

func NumCatsExceedsLimit(limit int) error {
	return fmt.Errorf("too many tag cats (%d max)", limit)
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}