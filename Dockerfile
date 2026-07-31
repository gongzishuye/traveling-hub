FROM golang:1.25

ENV PATH=/usr/local/go/bin:${PATH}
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /usr/local/bin/travelinghub ./cmd/server
EXPOSE 8080
CMD ["/usr/local/bin/travelinghub"]
