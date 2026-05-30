#!/bin/bash
set -e

APP_DIR=/opt/backpack
BINARY=$APP_DIR/backpack-app
SERVICE=backpack
HEALTH_URL=https://backpack.adediiji.uk/login
MAX_RETRIES=10
RETRY_INTERVAL=3

if [ ! -f $APP_DIR/backpack.db ]; then
    echo "Running initial migration..."
    $BINARY -db_path=$APP_DIR/backpack.db -migrate
fi

echo "Stopping service..."
sudo systemctl stop $SERVICE

echo "Setting permissions..."
chmod +x $BINARY

echo "Starting service..."
sudo systemctl start $SERVICE

echo "Waiting for app to come back up..."
for i in $(seq 1 $MAX_RETRIES); do
    if curl -sf $HEALTH_URL > /dev/null 2>&1; then
        echo "App is up after $i attempt(s)"
        exit 0
    fi
    echo "Attempt $i/$MAX_RETRIES — retrying in ${RETRY_INTERVAL}s..."
    sleep $RETRY_INTERVAL
done

echo "App failed to come back up — rolling back..."
sudo systemctl stop $SERVICE
exit 1