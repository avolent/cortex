#!/bin/sh

# Define the folder containing the markdown files
OBSIDIAN_PATH="/Users/williamchrisp/Documents/Notes/Cortex"

cp -R "${OBSIDIAN_PATH}" ./tmp

# Loop through all markdown files in the folder
find "tmp" -type f -name "*.md" | while read -r file; do
    # Use sed to replace the image links
    sed -i.backup 's/!\[\([^]]*\)\](cortex\/\([^)]*\))/![\1](\/\2)/g' "$file"
    # Use sed to replace the page links
    sed -i.backup 's/\[\([^]]*\)\](cortex\/pages\/\([^)]*\))/![\1](\/\2)/g' "$file"
    # Use sed to replace the headlinks
    sed -i.backup 's/\(\[[^]]*\]\)(#\([[:alnum:]]*\)%20\([[:alnum:]]*\))/\1(#\2-\3)/g' "$file"
done
find "tmp" -name "*.backup" -type f -delete
echo "Markdown updated..."

cp -R tmp/images public
cp -R tmp/pages src
echo "Updated Wiki files"

rm -R tmp
echo "Removed tmp folder..."
