package orm

import (
	"github.com/Moty1999/web/orm/internal/valuer"
	"github.com/Moty1999/web/orm/model"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry
}
