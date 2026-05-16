#!/bin/sh
set -e

# Replace environment variables in config.template.js and write to config.js
# This allows for runtime configuration in a containerized environment
envsubst '${API_URL} ${APP_ENV} ${ENABLE_DEBUG} ${VERSION}' < /usr/share/nginx/html/assets/config.template.js > /usr/share/nginx/html/assets/config.js

echo "Generated runtime config.js:"
cat /usr/share/nginx/html/assets/config.js

# Start Nginx
exec nginx -g 'daemon off;'
