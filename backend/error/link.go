package error

import (
	"errors"
	"fmt"
)

var (
	// Query links
	ErrInvalidPage   error = errors.New("invalid page provided")
	ErrInvalidLinkID error = errors.New("invalid link ID provided")
	ErrInvalidPeriod error = errors.New("invalid period provided")
	ErrNoLinkID      error = errors.New("no link ID provided")
	ErrNoLinkWithID  error = errors.New("no link found with given ID")
	ErrNoCats        error = errors.New("no cats provided")
	ErrNoPeriod      error = errors.New("no period provided")
	// Add link
	ErrNoURL             error = errors.New("no URL provided")
	ErrInvalidURL        error = errors.New("invalid URL provided")
	ErrRedirect          error = errors.New("invalid link destination: redirect detected")
	ErrCannotLikeOwnLink error = errors.New("cannot like your own link")
	ErrLinkAlreadyLiked  error = errors.New("link already liked")
	ErrLinkNotLiked      error = errors.New("link not already liked")
	ErrCannotCopyOwnLink error = errors.New("cannot copy your own link to your treasure map")
	ErrLinkAlreadyCopied error = errors.New("link already copied to treasure map")
	ErrLinkNotCopied     error = errors.New("link not already copied")
)

func ErrLinkURLCharsExceedLimit(limit int) error {
	return fmt.Errorf("url too long (max %d chars)", limit)
}

func ErrDuplicateLink(url string, duplicate_link_id string) error {
	return fmt.Errorf(
		"duplicate URL: %s\nsee /tag/%s", 
		url,
		duplicate_link_id,
	)
}
