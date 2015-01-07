#!/bin/bash

DEBUG_BUILD="$1"


appendData() {
	echo "" >> $2
	echo "" >> $2
	echo "" >> $2
	echo "$1" >> $2
}

# Check If the required commands exists
if ! which scss >/dev/null; then
    echo "SASS compiler does not exists!"
    exit 1
elif ! which uglifyjs >/dev/null; then
    echo "uglifyjs compiler does not exists!"
    exit 1
fi

## Build and compress the SASS files
echo "#> Compiling bulldozer.scss "
scss --unix-newlines --no-cache --sourcemap=none -t compressed ./sass/bulldozer.scss ./resources/css/bulldozer.css

##
## Build the bulldozer javascript library
##

echo "#> Building and compressing bulldozer.js"

# Add the bulldozer javascript parts
cat ./javascript/bulldozer.js > ./resources/js/bulldozer.js
appendData "$(cat ./javascript/utils.js)" ./resources/js/bulldozer.js
appendData "$(cat ./javascript/loadingindicator.js)" ./resources/js/bulldozer.js
appendData "$(cat ./javascript/connectionlost.js)" ./resources/js/bulldozer.js
appendData "$(cat ./javascript/websocket.js)" ./resources/js/bulldozer.js
appendData "$(cat ./javascript/ajaxsocket.js)" ./resources/js/bulldozer.js
appendData "$(cat ./javascript/socket.js)" ./resources/js/bulldozer.js
appendData "$(cat ./javascript/core.js)" ./resources/js/bulldozer.js
#appendData "$(cat ./javascript/auth.js)" ./resources/js/bulldozer.js
appendData "$(cat ./javascript/render.js)" ./resources/js/bulldozer.js

if [ "$DEBUG_BUILD" == "debug" ]; then
	# Just copy the uncompressed bulldozer file
	mv ./resources/js/bulldozer.js ./resources/js/bulldozer.min.js
else
	# Compress the bulldozer javascript file
	uglifyjs ./resources/js/bulldozer.js -c -m > ./resources/js/bulldozer.min.js
	rm ./resources/js/bulldozer.js
fi

