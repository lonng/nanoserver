package main

import (
	"fmt"
	"strings"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

var CmdShell = &Command{
	UsageLine: "shell driverName datasourceName",
	Short:     "a general shell to operate all kinds of database",
	Long: `
general database's shell for sqlite3, mysql, postgres.

    driverName        Database driver name, now supported four: mysql mymysql sqlite3 postgres
    datasourceName    Database connection uri, for detail infomation please visit driver's project page
`,
}

func init() {
	CmdShell.Run = runShell
	CmdShell.Flags = map[string]bool{}
}

var engine *xorm.Engine

func shellHelp() {
	fmt.Println(`
        show tables                    show all tables
        columns <table_name>         show table's column info
        indexes <table_name>        show table's index info
        exit                         exit shell
        source <sql_file>            exec sql file to current database
        dump [-nodata] <sql_file>    dump structs or records to sql file
        help                        show this document
        <statement>                    SQL statement
    `)
}

func runShell(cmd *Command, args []string) {
	if len(args) != 2 {
		fmt.Println("params error, please see xorm help shell")
		return
	}

	var err error
	engine, err = xorm.NewEngine(args[0], args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	engine.ShowSQL(false)
	engine.Logger().SetLevel(core.LOG_UNKNOWN)

	err = engine.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}

	var scmd string
	fmt.Print("xorm$ ")
	for {
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if strings.ToLower(input) == "help" {
			shellHelp()
			scmd = ""
			fmt.Print("xorm$ ")
			continue
		}

		if strings.ToLower(input) == "exit" {
			fmt.Println("bye")
			return
		}

		if !strings.HasSuffix(input, ";") {
			scmd = scmd + " " + input
			continue
		}

		scmd = scmd + " " + input
		lcmd := strings.TrimSpace(strings.ToLower(scmd))
		lcmd = strings.TrimRight(lcmd, ";")
		if strings.HasPrefix(lcmd, "select") {
			res, err := engine.Query(scmd + "\n")
			if err != nil {
				fmt.Println(err)
			} else {
				if len(res) <= 0 {
					fmt.Println("no records")
				} else {
					columns := make(map[string]int)
					for k, _ := range res[0] {
						columns[k] = len(k)
					}

					for _, m := range res {
						for k, s := range m {
							l := len(string(s))
							if l > columns[k] {
								columns[k] = l
							}
						}
					}

					var maxlen = 0
					for _, l := range columns {
						maxlen = maxlen + l + 3
					}
					maxlen = maxlen + 1

					fmt.Println(strings.Repeat("-", maxlen))
					fmt.Print("|")
					slice := make([]string, 0)
					for k, l := range columns {
						fmt.Print(" " + k + " ")
						fmt.Print(strings.Repeat(" ", l-len(k)))
						fmt.Print("|")
						slice = append(slice, k)
					}
					fmt.Print("\n")
					for _, r := range res {
						fmt.Print("|")
						for _, k := range slice {
							fmt.Print(" " + string(r[k]) + " ")
							fmt.Print(strings.Repeat(" ", columns[k]-len(string(r[k]))))
							fmt.Print("|")
						}
						fmt.Print("\n")
					}
					fmt.Println(strings.Repeat("-", maxlen))
					//fmt.Println(res)
				}
			}
		} else if lcmd == "show tables" {
			tables, err := engine.Dialect().GetTables()
			if err != nil {
				fmt.Println(err)
			} else {
				var maxlen int
				for _, table := range tables {
					if len(table.Name) > maxlen {
						maxlen = len(table.Name)
					}
				}
				head := "Table Name"
				if maxlen < len(head) {
					maxlen = len(head)
				}

				maxlen = maxlen + 2

				fmt.Println(strings.Repeat("-", maxlen+3))
				fmt.Print("| " + head + strings.Repeat(" ", maxlen-len(head)))
				fmt.Println("|")
				fmt.Println(strings.Repeat("-", maxlen+3))
				for _, table := range tables {
					fmt.Print("|")
					fmt.Print(" " + table.Name)
					fmt.Print(strings.Repeat(" ", maxlen-len(table.Name)))
					fmt.Println("|")
				}
				fmt.Println(strings.Repeat("-", maxlen+3))
			}
		} else if strings.HasPrefix(lcmd, "dump") {
			fields := strings.Fields(strings.TrimRight(scmd, ";"))
			if len(fields) == 2 {
				err = engine.DumpAllToFile(fields[1])
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("dump successfully!")
				}
			} else {
				fmt.Println("param error")
			}
		} else if strings.HasPrefix(lcmd, "source") {
			fields := strings.Fields(strings.TrimRight(scmd, ";"))
			if len(fields) == 2 {
				_, err = engine.ImportFile(fields[1])
				if err.Error() == "not an error" {
					err = nil
				}
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("param error")
			}
		} else if strings.HasPrefix(lcmd, "columns") {
			fields := strings.Fields(strings.TrimRight(scmd, ";"))
			if len(fields) == 2 {
				_, columns, err := engine.Dialect().GetColumns(fields[1])
				if err != nil {
					fmt.Println(err)
				} else {
					if len(columns) == 0 {
						fmt.Println("no column in", fields[1])
					} else {
						var maxlen int
						for name, _ := range columns {
							if len(name) > maxlen {
								maxlen = len(name)
							}
						}
						head := "Column Name"
						if maxlen < len(head) {
							maxlen = len(head)
						}

						maxlen = maxlen + 2

						fmt.Println(strings.Repeat("-", maxlen+3))
						fmt.Print("| " + head + strings.Repeat(" ", maxlen-len(head)))
						fmt.Println("|")
						fmt.Println(strings.Repeat("-", maxlen+3))
						for name, _ := range columns {
							fmt.Print("|")
							fmt.Print(" " + name)
							fmt.Print(strings.Repeat(" ", maxlen-len(name)))
							fmt.Println("|")
						}
						fmt.Println(strings.Repeat("-", maxlen+3))
					}
				}
			} else {
				fmt.Println("param error")
			}
		} else if strings.HasPrefix(lcmd, "indexes") {
			fields := strings.Fields(strings.TrimRight(scmd, ";"))
			if len(fields) == 2 {
				indexes, err := engine.Dialect().GetIndexes(fields[1])
				if err != nil {
					fmt.Println(err)
				} else {
					if len(indexes) == 0 {
						fmt.Println("no index in", fields[1])
					} else {
						var maxlen int
						for name, _ := range indexes {
							if len(name) > maxlen {
								maxlen = len(name)
							}
						}
						head := "Index Name"
						if maxlen < len(head) {
							maxlen = len(head)
						}

						maxlen = maxlen + 2
						fmt.Println(strings.Repeat("-", maxlen+3))
						fmt.Print("| " + head + strings.Repeat(" ", maxlen-len(head)))
						fmt.Println("|")
						fmt.Println(strings.Repeat("-", maxlen+3))
						for name, _ := range indexes {
							fmt.Print("|")
							fmt.Print(" " + name)
							fmt.Print(strings.Repeat(" ", maxlen-len(name)))
							fmt.Println("|")
						}
						fmt.Println(strings.Repeat("-", maxlen+3))
					}
				}
			} else {
				fmt.Println("param error")
			}
		} else {
			res, err := engine.Exec(scmd)
			if err != nil {
				fmt.Println(err)
			} else {
				cnt, err := res.RowsAffected()
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Printf("%d records changed.\n", cnt)
				}
			}
		}
		scmd = ""
		fmt.Print("xorm$ ")
	}
}
