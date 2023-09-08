package orm

import "github.com/Moty1999/web/orm/internal/errs"

// 通过这种形式将内部错误，暴露在外面
var ErrNoRows = errs.ErrNoRows
