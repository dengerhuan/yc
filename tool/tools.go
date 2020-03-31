package tool

import (
	"fmt"
	"os"
	)

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		//os.Exit(1)
	}
}
