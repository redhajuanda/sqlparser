#!/bin/bash

# Remove the dependency directory if it exists
if [ -d "dependencies" ]; then
  echo "Removing existing dependency directory..."
  rm -rf dependencies
fi

# Create the dependency directory
echo "Creating dependency directory..."
mkdir dependencies

# Loop 10 times and call sync_loop.sh
for i in {1..5}
do
  echo "Calling sync_loop.sh - iteration $i"
  ./sync.sh
done



# run go mod init github.com/redhajuanda/sqlparser
go mod init github.com/redhajuanda/sqlparser


#!/bin/bash

# Configuration
OLD_PREFIX="vitess.io/vitess/go/"
NEW_PREFIX="github.com/redhajuanda/sqlparser/dependencies/"

# Temporary file for sed operations on macOS/BSD
TEMP_FILE=$(mktemp)
trap 'rm -f "$TEMP_FILE"' EXIT

# Find all .go files recursively
while IFS= read -r -d '' file; do
    echo "Processing file: $file"
    
    # Create a temporary copy of the file
    cp "$file" "$TEMP_FILE"
    
    # Process multiline imports
    in_import_block=false
    while IFS= read -r line; do
        # Detect start of import block
        if [[ $line =~ ^[[:space:]]*import[[:space:]]*\( ]]; then
            in_import_block=true
            continue
        fi
        
        # Detect end of import block
        if [[ $line =~ ^\) ]] && [ "$in_import_block" = true ]; then
            in_import_block=false
            continue
        fi
        
        # Skip commented lines
        [[ $line =~ ^[[:space:]]*// ]] && continue
        
        if [ "$in_import_block" = true ]; then
            # Handle imports within block
            if [[ $line =~ \"$OLD_PREFIX ]]; then
                # Extract the import path and any alias
                if [[ $line =~ ^[[:space:]]*([[:alnum:]_]+[[:space:]]+)?\"($OLD_PREFIX[^\"]+)\" ]]; then
                    alias="${BASH_REMATCH[1]}"
                    old_import="${BASH_REMATCH[2]}"
                    new_import="${old_import/$OLD_PREFIX/$NEW_PREFIX}"
                    
                    # Construct the replacement pattern
                    if [ -n "$alias" ]; then
                        sed -i.bak "s|$alias\"$old_import\"|$alias\"$new_import\"|g" "$TEMP_FILE"
                    else
                        sed -i.bak "s|\"$old_import\"|\"$new_import\"|g" "$TEMP_FILE"
                    fi
                fi
            fi
        fi
    done < "$file"
    
    # Process single-line imports
    while IFS= read -r line; do
        # Skip commented lines
        [[ $line =~ ^[[:space:]]*// ]] && continue
        
        # Handle single-line imports
        if [[ $line =~ import[[:space:]]+([[:alnum:]_]+[[:space:]]+)?\"($OLD_PREFIX[^\"]+)\" ]]; then
            alias="${BASH_REMATCH[1]}"
            old_import="${BASH_REMATCH[2]}"
            new_import="${old_import/$OLD_PREFIX/$NEW_PREFIX}"
            
            if [ -n "$alias" ]; then
                sed -i.bak "s|$alias\"$old_import\"|$alias\"$new_import\"|g" "$TEMP_FILE"
            else
                sed -i.bak "s|\"$old_import\"|\"$new_import\"|g" "$TEMP_FILE"
            fi
        fi
    done < "$file"
    
    # Check if any changes were made
    if ! cmp -s "$file" "$TEMP_FILE"; then
        mv "$TEMP_FILE" "$file"
        echo "Updated imports in: $file"
    fi
    
    # Clean up backup files
    find . -name "*.bak" -type f -delete
    
done < <(find . -type f -name "*.go" -print0)

echo "Finished updating imports."

# Print summary of changes
echo -e "\nModified imports from '$OLD_PREFIX' to '$NEW_PREFIX'"
echo "Review the changes in your Git diff to verify the modifications."

# Run go mod tidy to clean up the go.mod file
go mod tidy