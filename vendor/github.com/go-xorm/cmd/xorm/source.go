package main

import (
	"fmt"
	"os"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

var CmdSource = &Command{
	UsageLine: "source driverName datasourceName",
	Short:     "source execute std in to datasourceName",
	Long: `
source from standard std in for sqlite3, mysql, postgres.

    driverName        Database driver name, now supported four: mysql mymysql sqlite3 postgres
    datasourceName    Database connection uri, for detail infomation please visit driver's project page
`,
}

func init() {
	CmdSource.Run = runSource
	CmdSource.Flags = map[string]bool{}
}

func runSource(cmd *Command, args []string) {
	if len(args) != 2 {
		fmt.Println("params error, please see xorm help source")
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

	_, err = engine.Import(os.Stdin)
	if err.Error() == "not an error" {
		err = nil
	}
	if err != nil {
		fmt.Println(err)
		return
	}
}
