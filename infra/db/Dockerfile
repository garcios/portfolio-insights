FROM alpine:3.18

# Install make, curl, and postgresql-client
RUN apk add --no-cache make curl postgresql-client

# Install golang-migrate
ENV MIGRATE_VERSION=v4.17.0
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

WORKDIR /migrations

# Copy migration files and Makefile
COPY . .

# Default command
CMD ["sh", "-c", "until pg_isready -d $DB_URL; do echo 'Waiting for database...'; sleep 2; done; make migrate-up"]
