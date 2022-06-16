package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-spatial/geom/encoding/gpkg"
	"github.com/urfave/cli/v2"
)

const SOURCE string = `source`
const TARGET string = `target`
const RESOLUTION string = `resolution`
const PAGESIZE string = `pagesize`

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
	}

	app.Action = func(c *cli.Context) error {
		sourcePath := c.String(SOURCE)
		targetPath := c.String(TARGET)
		pageSize := c.Int(PAGESIZE)
		resolution := c.Float64(RESOLUTION)

		return RunSieve(sourcePath, targetPath, pageSize, resolution)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func RunSieve(sourcePath string, targetPath string, pageSize int, resolution float64) error {
	_, err := os.Stat(sourcePath)
	if os.IsNotExist(err) {
		log.Fatalf("error opening source GeoPackage: %s", err)
	}
	srcHandle, err := gpkg.Open(sourcePath)
	if err != nil {
		log.Fatalf("error opening source GeoPackage: %s", err)
	}
	defer srcHandle.Close()

	trgHandle, err := gpkg.Open(targetPath)
	if err != nil {
		log.Fatalf("error opening target GeoPackage: %s", err)
	}
	defer trgHandle.Close()

	tables := getSourceTableInfo(srcHandle)

	err = initTargetGeopackage(trgHandle, tables)
	if err != nil {
		log.Fatalf("error initialization the target GeoPackage: %s", err)
	}

	log.Println("=== start sieving ===")

	// Process the tables sequential
	for _, table := range tables {
		log.Printf("  sieving %s", table.name)
		preSieve := make(chan feature)
		postSieve := make(chan feature)
		kill := make(chan bool)

		go writeFeatures(postSieve, kill, trgHandle, table, pageSize)
		go sieveFeatures(preSieve, postSieve, resolution)
		go readFeatures(srcHandle, preSieve, table)

		for {
			if <-kill {
				break
			}
		}
		close(kill)
		log.Println(fmt.Sprintf(`  finished %s`, table.name))
		log.Println("")
	}

	log.Println("=== done sieving ===")
	return nil
}
