package service

type VoucherOrderMessage struct {
	UserID    int64 `json:"user_id"`
	VoucherID int64 `json:"voucher_id"`
}