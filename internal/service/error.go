package service

import "errors"

var (
	ErrShopNotFound      = errors.New("shop not found")
	ErrInvalidReviewData = errors.New("invalid review data")
	ErrInvalidPhone = errors.New("invalid phone")
	ErrInvalidCode  = errors.New("invalid verification code")
	ErrUserNotFound = errors.New("user not found")

	ErrVoucherNotFound     = errors.New("voucher not found")
	ErrVoucherNotStarted   = errors.New("voucher not started")
	ErrVoucherEnded        = errors.New("voucher ended")
	ErrVoucherSoldOut      = errors.New("voucher sold out")
	ErrDuplicateVoucherOrder = errors.New("duplicate voucher order")
)