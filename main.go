package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/heshanpadmasiri/central-search/cmd"
	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"github.com/heshanpadmasiri/central-search/internal/central"
	"github.com/heshanpadmasiri/central-search/internal/centraldocs"
	"github.com/heshanpadmasiri/central-search/internal/search"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	centralClient, err := central.NewClient(central.DefaultBaseURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: configure Central client: %v\n", err)
		os.Exit(1)
	}
	searchService := search.NewService(centralClient)
	documentationClient, err := centraldocs.NewClient(centraldocs.DefaultBaseURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: configure Central documentation client: %v\n", err)
		os.Exit(1)
	}
	documentationService := catalog.NewService(searchService, centralClient, documentationClient)
	command := cmd.NewRootCommand(
		searchService,
		documentationService,
		cmd.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	)
	if err := command.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
