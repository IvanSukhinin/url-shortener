# URL SHORTENER

## sso grpc api for this app
https://github.com/IvanSukhinin/sso-grpc

## proto
https://github.com/IvanSukhinin/sso-proto

## local deploy
1) Containers build and up
```shell
docker-compose up -d
```
2) Migrations up
```shell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

```shell
cd ./migrations && \
goose postgres "host=localhost port=54321 user=postgres password=postgres database=url_shortener sslmode=disable" up
```
