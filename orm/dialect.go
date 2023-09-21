package orm

import (
	"github.com/Moty1999/web/orm/internal/errs"
)

var (
	DialectMySQL       Dialect = mysqlDialect{}
	DialectSQLite      Dialect = sqliteDialect{}
	DialectPostgresSQL Dialect = postgresDialect{}
)

type Dialect interface {
	// quoter 就是为了解决引号问题
	quoter() byte
	buildUpsert(b *builder, upsert *Upsert) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildUpsert(b *builder, upsert *Upsert) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) quoter() byte {
	return '`'
}

func (s mysqlDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range upsert.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			// 字段不同, 或者说列不对
			if !ok {
				return errs.NewErrUnknowField(a.col)
			}

			b.quote(fd.ColName)

			b.sb.WriteString("=?")
			b.args = append(b.args, a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			// 字段不同, 或者说列不对
			if !ok {
				return errs.NewErrUnknowField(a.name)
			}

			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		default:
			return errs.NewErrUnsupportAssignable(assign)
		}
	}

	return nil
}

type sqliteDialect struct {
	standardSQL
}

func (s sqliteDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON CONFLICT(")
	for i, col := range upsert.conflictColumns {
		if i > 0 {
			b.sb.WriteByte(',')
		}

		err := b.buildColumn(col)
		if err != nil {
			return err
		}

	}
	b.sb.WriteString(") DO UPDATE SET ")
	for idx, assign := range upsert.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			// 字段不同, 或者说列不对
			if !ok {
				return errs.NewErrUnknowField(a.col)
			}

			b.quote(fd.ColName)

			b.sb.WriteString("=?")
			b.args = append(b.args, a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			// 字段不同, 或者说列不对
			if !ok {
				return errs.NewErrUnknowField(a.name)
			}

			b.quote(fd.ColName)
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		default:
			return errs.NewErrUnsupportAssignable(assign)
		}
	}
	return nil
}
func (s sqliteDialect) quoter() byte {
	return '`'
}

type postgresDialect struct {
	standardSQL
}
