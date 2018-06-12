package main

import "fmt"

var CmdDriver = &Command{
	UsageLine: "driver",
	Short:     "list all supported drivers",
	Long: `
list all supported drivers
`,
}

func init() {
	CmdDriver.Run = runDriver
	CmdDriver.Flags = map[string]bool{}
}

func runDriver(cmd *Command, args []string) {
	for n, d := range supportedDrivers {
		fmt.Println(n, "\t", d)
	}
}
