%~dp0dbr --driver mysql --source root:88888888@tcp(10.10.30.52:3306)/broker --destination ../../internal/model --template struct.xorm.kwx.tpl --single --file struct_kwx.go
