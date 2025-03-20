#!/bin/bash

# Script to replace import paths in Go files
# Replace "notifuse/server/" with "github.com/Notifuse/notifuse/"

# Set the directory to search (default: current directory)
SEARCH_DIR="${1:-.}"

# Display what will be changed first
echo "The following files will be modified:"
grep -r --include="*.go" "notifuse/server/" "$SEARCH_DIR" | cut -d: -f1 | sort | uniq

# Ask for confirmation
read -p "Do you want to proceed with the changes? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    echo "Operation cancelled."
    exit 1
fi

# Perform the replacement
echo "Updating import paths..."
find "$SEARCH_DIR" -name "*.go" -type f -exec sed -i '' 's|"notifuse/server/|"github.com/Notifuse/notifuse/|g' {} \;

# Also update the go.mod file if it exists
if [ -f "$SEARCH_DIR/server/go.mod" ]; then
    echo "Updating server/go.mod..."
    sed -i '' 's|module notifuse/server|module github.com/Notifuse/notifuse|' "$SEARCH_DIR/server/go.mod"
fi

echo "Done! Please check your code and test it thoroughly."
echo "Note: You may need to run 'go mod tidy' after these changes." 