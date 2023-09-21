package orm

type Assignment struct {
	col string
	val any
}

func Assign(col string, val any) Assignable {
	return Assignment{
		col: col,
		val: val,
	}
}

func (a Assignment) assign() {

}
