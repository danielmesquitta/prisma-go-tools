# Prisma to Go

Parser to convert schema.prisma files into Golang structs

I appreciate Prisma’s migration management, but I prefer writing raw SQL queries over using ORMs. To combine the best of both worlds, I created this parser. It allows me to use Prisma's migration management without relying on its ORM.

## Installation

```bash
go install github.com/danielmesquitta/prisma-to-go@latest
```

## Usage

```bash
prisma-to-go --schema ./path/to/schema.prisma --output ./path/to/output/dir
```
