package main

import (
	"time"

	"github.com/danielmesquitta/prisma-go-tools/cmd"
)

func main() {
	time.Local = time.UTC
	cmd.Execute()
}
