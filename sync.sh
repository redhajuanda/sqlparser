#!/bin/bash

BASE_DIR="/Users/redhajuanda/dev/redhajuanda"

# Source directory (folder from which files will be copied)
SOURCE_DIR="/vitess/go/vt/sqlparser/"
DEST_DIR="/sqlparser"
VERSION="v21.0.2"

# go to source directory and git checkout the version
cd $BASE_DIR$SOURCE_DIR
git checkout $VERSION

# Copy all files from the source directory to the destination directory, if the files already exist, overwrite them
cp -R $BASE_DIR$SOURCE_DIR $BASE_DIR$DEST_DIR

echo "Copy operation completed."


# find dependencies
cd $BASE_DIR$DEST_DIR

#!/bin/bash

# Prefix to search for
PREFIX="vitess.io/vitess/go/"

# Declare an array to store distinct imports
declare -a IMPORTS=()

# Find all .go files recursively
while IFS= read -r -d '' file; do
    # Process multiline imports
    while IFS= read -r line; do
        # Skip commented lines
        [[ $line =~ ^[[:space:]]*// ]] && continue
        
        # Clean the line: remove quotes and leading/trailing whitespace
        cleaned=$(echo "$line" | sed 's/"//g' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        
        # Skip empty lines
        [[ -z $cleaned ]] && continue
        
        # If line has an alias, extract the import path after the alias
        if [[ $cleaned =~ ^[[:alnum:]_]+[[:space:]]+(.*) ]]; then
            cleaned="${BASH_REMATCH[1]}"
        fi
        
        # Check if line starts with our prefix
        [[ $cleaned == $PREFIX* ]] && IMPORTS+=("$cleaned")
        
    done < <(awk '/^import[[:space:]]*\($/{p=1;next} /^\)/{p=0} p{print}' "$file")
    
    # Process single-line imports
    while IFS= read -r line; do
        # Skip commented lines
        [[ $line =~ ^[[:space:]]*// ]] && continue
        
        # Handle both aliased and non-aliased single-line imports
        if [[ $line =~ import[[:space:]]+([[:alnum:]_]+[[:space:]]+)?\"([^\"]+)\" ]]; then
            import="${BASH_REMATCH[2]}"
            
            # Add if it matches our prefix
            [[ $import == $PREFIX* ]] && IMPORTS+=("$import")
        fi
    done < <(grep -E '^[[:space:]]*import[[:space:]]' "$file")
    
done < <(find . -type f -name "*.go" -print0)

# Remove duplicates and sort
IMPORTS=($(printf "%s\n" "${IMPORTS[@]}" | sort -u))

# Output the distinct imports
# echo "Distinct imports found from ${PREFIX}:"
for import in "${IMPORTS[@]}"; do
    echo "$import"
done


cd $BASE_DIR"/vitess/go"

for import in "${IMPORTS[@]}"; do
    #hapus prefix "vitess.io"
    import=${import#"vitess.io/vitess/go/"}
    # echo "Import: $import"

    # echo $BASE_DIR$import $BASE_DIR$DEST_DIR"/dependencies"

    rsync -av --relative $import $BASE_DIR$DEST_DIR"/dependencies"
done


