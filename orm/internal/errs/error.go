package errs

import (
	"errors"
	"fmt"
)

var (
	// ErrPointerOnly 只支持一级指针作为输入
	// 看到这个 error 说明你输入了其他的东西
	// 我们并不希望用户直接使用 err == ErrPointerOnly
	// 所以放在我们的 internal 包里
	ErrPointerOnly           = errors.New("orm: 只支持指向结构体的一级指针")
	ErrUnsupportedExpression = errors.New("orm: 不支持的表达式类型")
	ErrNoRows                = errors.New("orm: 没有数据")
)

func NewErrUnsupportedExpressionV1(expr any) error {
	return fmt.Errorf("%w %v", ErrUnsupportedExpression, expr)
}

// @ErrUnsupportedExpression 40001 原因是你输入了乱七八糟的类型
// 解决方案: 使用正确的类型
func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("orm: 不支持的表达式类型 %v", expr)
}

func NewErrUnknowField(name string) error {
	return fmt.Errorf("orm: 未知字段 %s", name)
}

func NewErrInvalidTagContext(pair string) error {
	return fmt.Errorf("orm: 非法标签值 %s", pair)
}
func NewErrUnknowFieldColumn(name string) error {
	return fmt.Errorf("orm: 未知列 %s", name)
}
