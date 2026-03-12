package service

import "errors"

var (
	ErrShopNotFound = errors.New("shop not found")
	ErrInvalidReviewData= errors.New("invalid review data")
)