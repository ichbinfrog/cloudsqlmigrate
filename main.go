package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/ichbinfrog/cloudsqlmigrate/migrate"
)

func main() {
	srcProject := flag.String("src-project", "", "GCP project where the source CloudSQL Instance is provisioned")
	srcInstance := flag.String("src-instance", "", "Name of the source CloudSQL Instance")
	dstProject := flag.String("dst-project", "", "GCP project where the target CloudSQL Instance is provisioned")
	dstInstance := flag.String("dst-instance", "", "Name of the target CloudSQL Instance")
	flag.Parse()

	if *srcProject == "" {
		log.Fatalf("missing argument: src-project is required")
	}
	if *srcInstance == "" {
		log.Fatalf("missing argument: src-instance is required")
	}
	if *dstProject == "" {
		log.Fatalf("missing argument: dst-project is required")
	}
	if *dstInstance == "" {
		log.Fatalf("missing argument: dst-instance is required")
	}

	ctx := context.Background()

	start := time.Now()
	log.Printf("starting the migration of %s:%s to %s:%s at %s", *srcProject, *srcInstance, *dstProject, *dstInstance, start.String())
	op, err := migrate.NewOp(ctx, *srcProject, *srcInstance, *dstProject, *dstInstance)
	if err != nil {
		log.Fatal(err)
	}

	if errs := op.Run(ctx,
		[]migrate.Preflight{
			migrate.PreflightVersion{},
			migrate.PreflightStatus{},
		},
		[]migrate.PostFlight{
			&migrate.PostFlightSQLAdmin{},
		},
	); len(errs) > 0 {
		log.Fatalf("checks failed: %v", errs)
	}

	log.Printf("operation completed successfully in %s", time.Since(start).String())
}
