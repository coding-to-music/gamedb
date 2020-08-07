package main

import (
	"os"
)

func main() {

	if len(os.Args) > 1 {

		str := ""

		for _, v := range os.Args[1:] {

			str += "" + "\n" +
				"**/*.go {" + "\n" +
				"    indir: ./cmd/" + v + "\n" +
				"    prep: '" + "\n" +
				"        # Building " + v + "\n" +
				"        go build -ldflags \"-X main.version=`git rev-parse --verify HEAD` -X main.commits=`git rev-list --count master`\"" + "\n" +
				"        '" + "\n" +
				"    daemon: ./" + v + "\n" +
				"}" + "\n" +
				""
		}

		f, err := os.OpenFile("modd.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			panic(err)
		}

		defer func() {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}()

		_, err = f.WriteString(str)
		if err != nil {
			panic(err)
		}
	}
}
