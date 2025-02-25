#!/bin/bash
# Description: This script is used to run the migration scripts for the airline service
# read the environment variables from the .env file
# print present working directory


# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ENV_FILE="$SCRIPT_DIR/../.env"
DB_MIGRATIONS_DIR="$SCRIPT_DIR/../db/migrations"

# print error if the ENV_FILE does not exist
if [[ ! -f "$ENV_FILE" ]]; then
    echo "File not found: $ENV_FILE"
    exit 1
fi

# load the environment variables
source $ENV_FILE

# run migrations using the migrate tool
migrate -database "mysql://root:$DB_PASSWORD@tcp(127.0.0.1:3306)/airline" -path $DB_MIGRATIONS_DIR up