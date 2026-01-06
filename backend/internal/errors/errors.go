/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-26 16:24:26
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 15:31:26
 * @FilePath            : frp-web-testbackendinternalerrorserrors.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package errors

import (
	"fmt"
	"net/http"
)

// AppError 统一应用错误类型
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// 预定义错误码
const (
	CodeSuccess      = 0
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeConflict     = 409
	CodeInternal     = 500
	CodeValidation   = 1001
	CodeDatabase     = 1002
	CodeExternal     = 1003
)

// 常用错误构造函数
func NewBadRequest(message string) *AppError {
	return &AppError{Code: CodeBadRequest, Message: message}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{Code: CodeUnauthorized, Message: message}
}

func NewForbidden(message string) *AppError {
	return &AppError{Code: CodeForbidden, Message: message}
}

func NewNotFound(message string) *AppError {
	return &AppError{Code: CodeNotFound, Message: message}
}

func NewConflict(message string) *AppError {
	return &AppError{Code: CodeConflict, Message: message}
}

func NewInternal(message string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: message, Err: err}
}

func NewValidation(message string) *AppError {
	return &AppError{Code: CodeValidation, Message: message}
}

func NewDatabase(message string, err error) *AppError {
	return &AppError{Code: CodeDatabase, Message: message, Err: err}
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeBadRequest, CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Wrap 包装已有错误
func Wrap(err error, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return &AppError{Code: CodeInternal, Message: message, Err: err}
}

// WrapWithCode 使用指定错误码包装错误
func WrapWithCode(err error, code int, message string) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// IsAppError 检查错误是否为 AppError 类型
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError 将错误转换为 AppError，如果不是则返回 nil
func AsAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}
