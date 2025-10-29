#!/bin/bash
#
# zone-recorder.sh:
# 
# Interactively generate a firstgo config.yaml file by defining
# clickable zones on a series of images. Suitable for window managers
# such as i3 on Linux; presently untested on other WMs/platforms.

# -- 0. check input

# Usage: ./zone-recorder.sh <output_yaml> <image1> [image2...]
if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <output_yaml> <image1> [image2...]"
    echo ""
    echo "Creates a new config file or appends new pages to an existing one."
    exit 1
fi

# yaml is arg1, the rest are images
YAML_FILE="$1"
IMAGE_FILES=("${@:2:$#-1}")

# Check for required binaries
for tool in qiv slop; do 
    if ! command -v "$tool" &> /dev/null; then
      echo "Error: Required binary '$tool' is not installed or not on your PATH." >&2
      exit 1
    fi
done

# -- 2. create yaml if it doesn't exist

if [ ! -f "$YAML_FILE" ]; then
    echo "Creating new config file: $YAML_FILE"
    
    # write standard firstgo yaml header
    cat > "$YAML_FILE" << EOF
---

# directories
assetsDir: "assets"

# templates within assets/templates directory
pageTemplate: "templates/page.html"
indexTemplate: "templates/index.html"

# list of pages
pages:
EOF
else
    echo "Appending new pages to existing file: $YAML_FILE"
fi

# -- 3. loop through images

IMAGE_COUNT=${#IMAGE_FILES[@]}
CURRENT_INDEX=1

for IMAGE_FILE in "${IMAGE_FILES[@]}"; do
    echo "---"
    echo "Processing image $CURRENT_INDEX of $IMAGE_COUNT: $IMAGE_FILE"

    # -- 3a. generate page stubs
    FILENAME=$(basename -- "$IMAGE_FILE")
    BASENAME="${FILENAME%.*}" # Removes the extension
    IMAGE_FILE_CLEANED=${IMAGE_FILE/assets\//} # remove 'assets' dir
    URL="/$BASENAME"
    # titlecase title
    TITLE="$(tr '[:lower:]' '[:upper:]' <<< ${BASENAME:0:1})${BASENAME:1}"
    {
        echo "  -"
        echo "    URL: \"$URL\""
        echo "    Title: \"$TITLE\""
        echo "    ImagePath: \"$IMAGE_FILE_CLEANED\""
        echo "    Note: \"\""
        echo "    Zones:"
    } >> "$YAML_FILE"

    # -- 3b. display image and capture zones
    # launch qiv in the background
    # -W 100 = 100% (true scale)
    # -e     = set to top left of screen (0,0)
    qiv -W 100 -e "$IMAGE_FILE" &
    QIV_PID=$!

    # origin 0,0 due to 'qiv -e'
    ORIGIN_X=0
    ORIGIN_Y=0
    
    # -- 3c. capture zones, breaking with 'd' key to proceed to next image
    while true; do
        read -p "Enter Target URL (e.g., /about), or 'd' when done with this image: " TARGET_URL
        if [[ "$TARGET_URL" == "d" ]]; then
            break
        fi

        echo "Please select the rectangle for target '$TARGET_URL'..."
        
        # use slop (not AI slop) a program to "query for a [mouse]
        # selection from the user and prints the region to stdout."
        # -o : no hardware acceleration
        # -b : selection border width
        # -c : color of selection
        # -f : format
        SELECTION=$(slop -o -b 2 -c 6.2,4.3,24.8,0.2 -f "%x %y %w %h")
        if [ -z "$SELECTION" ]; then
            echo "Selection cancelled."
            continue # go back to prompt
        fi
        
        read -r SEL_X SEL_Y SEL_W SEL_H <<< "$SELECTION"
        
        LEFT=$((SEL_X - ORIGIN_X))
        TOP=$((SEL_Y - ORIGIN_Y))
        RIGHT=$((LEFT + SEL_W))
        BOTTOM=$((TOP + SEL_H))

        # append zone data to the yaml
        {
            echo "      -"
            echo "        Left:   $LEFT"
            echo "        Top:    $TOP"
            echo "        Right:  $RIGHT"
            echo "        Bottom: $BOTTOM"
            echo "        Target: \"$TARGET_URL\""
        } >> "$YAML_FILE"

        echo "Zone for '$TARGET_URL' saved."
    done

    # -- 4. cleanup
    kill "$QIV_PID"
    ((CURRENT_INDEX++))
done

echo "---"
echo "Processing complete. Configuration saved to '$YAML_FILE'."
