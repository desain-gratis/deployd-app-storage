FROM ubuntu:24.04

# Allow hit https endpoint from inside 
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Install chDB library
RUN curl -sL https://lib.chdb.io | bash

# important for cleanly closing connection
STOPSIGNAL SIGINT

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY ./dist/storage .

# Command to run the application
CMD ["./storage"]
