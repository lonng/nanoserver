{{if not .IsAppend}}package {{.Model}}{{end}}

{{$ilen := len .Imports}}{{if gt $ilen 0}}
import (
	{{range .Imports}}"{{.}}"{{end}}
){{end}}

{{range .Tables}}
type {{Mapper .Name}} struct {
{{$table := .}}
{{range .ColumnsSeq}}{{$col := $table.GetColumn .}}	{{Mapper $col.Name}}	{{Type $col}} {{Tag $table $col}}
{{end}}
}
{{end}}

func syncSchema() {
    DB.StoreEngine("InnoDB").Sync2({{range .Tables}}
    new({{Mapper .Name}}),{{end}}
    )
}