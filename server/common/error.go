package common

import "errors"

var (
	ErrContentTypeNotMultipartFormData = errors.New("Content-Type is not multipart/form-data")
	ErrRequestParamEmpty               = errors.New("request param is empty")
	ErrRequestParamInvalid             = errors.New("request param is invalid")
)
