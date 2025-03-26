#!/bin/bash

rye sync

if [ "$#" -gt 1 ]; then
    echo "Invalid number of arguments. Please provide only one argument."
    echo "Usage: ./run_scraper.sh [-test | -t]"
    exit 1
fi

if [ "$#" -eq 1 ]; then
    if [ "$1" == "-test" ] || [ "$1" == '-t' ]; then
        echo "Running Go scraper..."
        go run main.go -test
    else
        echo "Error: Invalid argument '$1'."
        echo "Usage: ./run_scraper.sh [-test | -t] for test mode or no arguments for full scrape."
        exit 1
    fi
else
    echo "Running Go scraper..."
    go run main.go
fi

if [ $? -eq 0 ]; then
    echo "Scraping completed successfully. Now inserting into database..."
    rye run python scripts/daily_comparison.py
else
    echo "Scraping failed. Please check for errors in the Go script."
fi
