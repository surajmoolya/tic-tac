FROM heroiclabs/nakama-pluginbuilder:3.21.1 AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /backend
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -mod=mod --trimpath --buildmode=plugin -o ./backend.so

FROM heroiclabs/nakama:3.21.1

COPY --from=builder /backend/backend.so /nakama/data/modules/
COPY local.yml /nakama/data/
