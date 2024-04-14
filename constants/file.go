package constants

import (
	"fmt"
	"os"
)

var FILE_BASE_PATH string

func LoadFileBasePath() {
	if len(os.Args) == 1 {
		FILE_BASE_PATH = ""
	} else {
		FILE_BASE_PATH = fmt.Sprintf("%s/", os.Args[1])
	}
}
