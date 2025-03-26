#!/bin/bash

rye sync

echo "Running Go scraper..."
go run main.go

if [ $? -eq 0 ]; then
    echo "Scraping completed successfully. Now converting to Excel..."
    rye run python scripts/convert_to_excel.py
else
    echo "Scraping failed. Please check for errors in the Go script."
fi
