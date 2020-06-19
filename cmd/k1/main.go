package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"gitlab.node-3.net/nadams/gpr/cmd/k1/list"
)

type CLI struct {
	In string `arg:"" name:"in" help:"List of proteins to check against K1"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)
	ctx.FatalIfErrorf(ctx.Validate())

	if err := do(cli.In); err != nil {
		log.Fatalln(err)
	}
}

func do(in string) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}

	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		if err := scan.Err(); err != nil {
			if err == io.EOF {
				break
			}
		}

		str := strings.TrimSpace(scan.Text())

		if _, ok := list.K1[str]; ok {
			fmt.Println(str)
		}
	}

	return nil
}
