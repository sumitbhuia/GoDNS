#!/bin/sh
set -e  # Exit immediately if a command exits with a non-zero status.

cd "$(dirname "$0")"  # Ensure compile steps are run within the repository directory
go build -o ./GoDNS *.go
./GoDNS  # Actually run the server after building