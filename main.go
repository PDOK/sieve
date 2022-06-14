package main

import (
	"github.com/go-spatial/geom/encoding/gpkg"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	SOURCE      string = `source`
	TARGET      string = `target`
	RESOLUTION  string = `resolution`
	PAGESIZE    string = `pageSize`
	MEMORYLIMIT string = `memoryLimit`
)

func main() {
	app := cli.NewApp()
	app.Name = "GOSieve"
	app.Usage = "A Golang Polygon Sieve application"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     SOURCE,
			Aliases:  []string{"s"},
			Usage:    "Source GPKG",
			Required: true,
			EnvVars:  []string{"SOURCE_GPKG"},
		},
		&cli.StringFlag{
			Name:     TARGET,
			Aliases:  []string{"t"},
			Usage:    "Target GPKG",
			Required: true,
			EnvVars:  []string{"TARGET_GPKG"},
		},
		&cli.Float64Flag{
			Name:     RESOLUTION,
			Aliases:  []string{"r"},
			Usage:    "Resolution, the threshold area to determine if a feature is sieved or not",
			Value:    0.0,
			Required: false,
			EnvVars:  []string{"SIEVE_RESOLUTION"},
		},
		&cli.IntFlag{
			Name:     PAGESIZE,
			Aliases:  []string{"p"},
			Usage:    "Page Size, how many features are written per transaction to the target GPKG",
			Value:    1000,
			Required: false,
			EnvVars:  []string{"SIEVE_PAGESIZE"},
		},
		&cli.IntFlag{
			Name:     MEMORYLIMIT,
			Aliases:  []string{"p"},
			Usage:    "Memory Limit for the application in bytes", // TODO make it megabytes
			Value:    0,
			Required: false,
			EnvVars:  []string{"SIEVE_MEMORY_LIMIT"},
		},
	}

	app.Action = func(c *cli.Context) error {
		source := c.String(SOURCE)
		target := c.String(TARGET)
		pageSize := c.Int(PAGESIZE)
		resolution := c.Float64(RESOLUTION)
		memoryLimit := c.Int(MEMORYLIMIT)
		return RunSieve(source, target, pageSize, resolution, memoryLimit)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func RunSieve(sourcePath string, targetPath string, pageSize int, resolution float64, memoryLimit int) error {
	_, err := os.Stat(sourcePath)
	if os.IsNotExist(err) {
		log.Fatalf("error opening source GeoPackage: %s", err)
	}

	source := SourceGeopackage{}
	source.Init(sourcePath)
	defer func(handle *gpkg.Handle) {
		err := handle.Close()
		if err != nil {
			log.Fatalf("error closing the source GeoPackage: %s", err)
		}
	}(source.handle)

	target := TargetGeopackage{}
	target.Init(targetPath, pageSize, memoryLimit)
	defer func(handle *gpkg.Handle) {
		err := handle.Close()
		if err != nil {
			log.Fatalf("error closing the target GeoPackage: %s", err)
		}
	}(target.handle)

	tables := source.GetTableInfo()

	err = target.CreateTables(tables)
	if err != nil {
		log.Fatalf("error initialization the target GeoPackage: %s", err)
	}

	log.Println("=== start sieving ===")

	// Process the tables sequential
	for _, table := range tables {
		log.Printf("  sieving %s", table.name)
		source.table = table
		target.table = table
		Sieve(source, target, resolution)
		log.Printf("  finised %s", table.name)
	}

	log.Println("=== done sieving ===")
	return nil
}
