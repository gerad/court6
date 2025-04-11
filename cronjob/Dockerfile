FROM alpine:latest

# Install curl and dcron
RUN apk add --no-cache curl dcron

# Copy crontab file
COPY crontab /etc/crontabs/root

# Start cron in foreground
CMD ["crond", "-f"] 
