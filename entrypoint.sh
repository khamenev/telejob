#!/bin/bash

# Ensure that the script stops on errors
set -e

# Write out the current crontab
crontab -l > mycron

# Echo new cron into cron file
echo "0 12 * * * /root/main" >> mycron
echo "0 18 * * * /root/main" >> mycron

# Install new cron file
crontab mycron

# Remove the temporary cron file
rm mycron

# Start cron
crond -f -d 8
