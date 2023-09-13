package valuer

import (
	"database/sql"
	"github.com/Moty1999/web/orm/internal/errs"
	"github.com/Moty1999/web/orm/model"
	"reflect"
	"unsafe"
)

type UnsafeValue struct {
	model *model.Model

	// 对应 T 的指针
	val any
}

// 这就是判断 NewUnsafeValue 是否符合 Creator 签名, 当 Creator 发生变化的时候, 这里会飘红提醒
var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *model.Model, val any) Value {
	return UnsafeValue{
		model: model,
		val:   val,
	}
}

func (u UnsafeValue) SetColumns(rows *sql.Rows) error {

	// 在这里处理结果集

	// 我怎么知道你 SELECT 出来了哪些列？
	// 拿到了 SELECT 出来的列
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	var vals []any
	address := reflect.ValueOf(u.val).UnsafePointer()
	for _, c := range cs {
		fd, ok := u.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknowFieldColumn(c)
		}

		// 是不是要计算字段的地址
		// 起始地址 + 偏移量
		fdAddress := unsafe.Pointer(uintptr(address) + fd.Offset)
		val := reflect.NewAt(fd.Type, fdAddress)
		vals = append(vals, val.Interface())
	}

	err = rows.Scan(vals...)
	return err
}
