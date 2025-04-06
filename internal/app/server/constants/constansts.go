package consts

type ctxKey int8

const (
	CtxRequestKey ctxKey = iota
	CtxUserIdKey  ctxKey = iota
)
