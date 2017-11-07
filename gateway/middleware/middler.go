package middleware

import (
	"github.com/julienschmidt/httprouter"
)

type Middler interface {
	Middle() httprouter.Handle
}

type MiddleWrap func() httprouter.Handle
