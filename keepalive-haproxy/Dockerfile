FROM haproxy:2.6-alpine

# Install keepalived and other necessary packages
USER root
RUN apk update && \
    apk add --no-cache keepalived ipvsadm bash curl procps

# Copy startup script
COPY start.sh /start.sh
RUN chmod +x /start.sh

# Set entrypoint to our startup script
ENTRYPOINT ["/start.sh"]