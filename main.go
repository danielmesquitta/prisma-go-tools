package main

import (
	"time"

	"github.com/danielmesquitta/prisma-to-go/cmd"
)

func main() {
	time.Local = time.UTC
	cmd.Execute()
}
