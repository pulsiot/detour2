############################################
# STEP 1 build docker image for detour2-proxy
############################################
FROM golang:1.20.5-bookworm AS builder
ENV GITHUB_ORG=pulsiot \
    GITHUB_REPO=detour2 \
    DETOUR_CONF_DIR=/build/etc/detour

WORKDIR $GOPATH/src/${GITHUB_ORG}/${GITHUB_REPO}/
COPY . .
RUN go mod tidy
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/detour2-proxy
# Create the detour config directory
RUN mkdir -vp ${DETOUR_CONF_DIR}
COPY detour2.yaml ${DETOUR_CONF_DIR}
COPY server.crt ${DETOUR_CONF_DIR}
COPY server.key ${DETOUR_CONF_DIR}
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /bin/detour2-proxy /app/detour2-proxy
COPY --from=builder /build/etc /etc
WORKDIR /app
# Run the hello binary.
ENTRYPOINT ["/app/detour2-proxy"]
