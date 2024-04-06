# InHouseAd_assignment

Simple API for work with products, catalogs and user

# Table of contents

- [Startup](#startup)
    - [Configuration](#configuration)
    - [Preparing environment variables](#preparing-environment-variables)
    - [Direct startup](#direct-startup)
    - [Docker startup](#docker-startup)
- [Http request examples](#http-request-examples)
    - [User handlers](#user-handlers)
      - [Register](#register)
      - [Login](#login)
      - [Refresh](#refresh)
    - [Category handlers](#ategory-handlers)
      - [Category create](#category-create)
      - [Category read](#category-read)
      - [Category update](#category-update)
      - [Category delete](#category-delete)
    - [Product handlers](#product-handlers)
      - [Product create](#product-create)
      - [Product read](#product-read)
      - [Product update](#product-update)
      - [Product delete](#product-delete)

## Startup

This section describes the configuration of the application and how to run it

### Configuration

The config has the following fields:

```yaml
log_level: debug
http:
  port: 9099
  host: 0.0.0.0
  ping_timeout: 2s
postgres:
  db_con_format: postgres
  db_host: postgres
  db_port: 5432
  db_user: user
  db_pass: password
  db_name: InHouseAd_assignment
  db_tbl_user: test-user
  db_tbl_category: test-category
  db_tbl_product: test-product
  db_tbl_product_category: test-product_category
data_collect_time: 1h
data_collect_link: https://emojihub.yurace.pro/api/all
refresh_ttl: 1h
token_ttl: 240h
secret_key: test-key
```

- `log_level` - level reports the minimum record level that will be logged.
- `http` - settings for http server.
    - `ping_timeout` - timeout for healthcheck.
- `postgres` - setting for connection and name of tabbles that will be used.
- `data_collect_time` - interval for auto collecting data (products and categories) from source.
- `data_collect_link` - the link of source from which data will be collected.
- `refresh_ttl` & `token_ttl` - time to live for access and refresh tokens
- `secret_key` - a key to sign jwt

Also, the following path `storage/init/init.sh` contains a script for creating a database.

### Preparing environment variables

You can use script to convert .yaml to .env file

`go run config_to_env.go <path_to_config>`

**! The order will be disrupted !**

### Direct startup

You can use build command to get bin file :</br>
`CGO_ENABLED=0 GOOS=linux go build -o <output path> <path to main.go>`

Than you need to start file via command:</br>
`file -config=<path_to_config>` - if you want to use config file.
Or just run file without flag to use env variables.

Also you can run service by using `go run`: `go run ./cmd/server/main.go -config=./configs/config.yaml` where you also can use config file or env variables.

### Docker startup

You can start only service by launching Dockerfile or start service with the database by launchig docker-compose file: `docker-compose up`

## Http request examples

### User handlers

#### Register

Request:
```
curl --location --request POST 'localhost:9999/api/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "newuser@email.com",
    "password":"t"
}'
```
Response
```json
HTTP/1.1 201 Created
Content-Type: application/json
Date: Sat, 06 Apr 2024 09:51:00 GMT
Content-Length: 19
 
{"registered":true}
```

#### Login
Request:
```
curl --location --request POST 'localhost:9999/api/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "newuser@email.com",
    "password":"t"
}'
```
Response
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 06 Apr 2024 09:52:56 GMT
Content-Length: 303
 
{
    "token_pair": {
        "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMTc1fQ.W_p3bxqo3pC8F3izno9PiHW1WQgcDXtGjg0xcnPnHPMQ5VEfh0GlRZq7JKP_d8Bp_uNzyZFlzZDzjcUs9RDRLQ",
        "refresh_token": "6959f1438acbfe99170fe738d585703c6350b3981f7962f5e79941faaad51d40"
    }
}
```
#### Refresh
Request:
```
curl --location --request POST 'localhost:9999/api/refresh' \
--header 'Content-Type: application/json' \
--data '{
    "token_pair": {
        "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMTc1fQ.W_p3bxqo3pC8F3izno9PiHW1WQgcDXtGjg0xcnPnHPMQ5VEfh0GlRZq7JKP_d8Bp_uNzyZFlzZDzjcUs9RDRLQ",
        "refresh_token": "6959f1438acbfe99170fe738d585703c6350b3981f7962f5e79941faaad51d40"
    }
}'
```
Response
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 06 Apr 2024 09:53:41 GMT
Content-Length: 303

{
    "token_pair": {
        "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMjIxfQ.puFqUPFfCEQVoynBhLDkwtflAVAuNXqiDSP09tCEmanZDEYxm2f0jSlFM17RtA9jIRmfJGHqp4SqTSxzY1zixQ",
        "refresh_token": "2b5a9fb537cd48341bb299af9e32428cd2acb2ca2f005e380c2b1a8fa8f122de"
    }
}
```

### Category handlers

#### Category create
Request:
```
curl --location --request POST 'localhost:9999/api/category/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMjIxfQ.puFqUPFfCEQVoynBhLDkwtflAVAuNXqiDSP09tCEmanZDEYxm2f0jSlFM17RtA9jIRmfJGHqp4SqTSxzY1zixQ' \
--data '{
    "category": {
        "name":"New test category",
        "code":"new_test_category",
        "description":"Description of test category"
    }
}'
```
Response
```json
HTTP/1.1 201 Created
Content-Type: application/json
Date: Sat, 06 Apr 2024 10:00:03 GMT
Content-Length: 14
 
{
  "added":true
}
```
#### Category read
Request:
```
curl --location --request GET 'localhost:9999/api/category/new_test_category'
```
Response
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 06 Apr 2024 10:01:22 GMT
Content-Length: 113

{
    "category": {
        "name": "New test category",
        "code": "new_test_category",
        "description": "Description of test category"
    }
}
```
#### Category update
Request:
```
curl --location --request PATCH 'localhost:9999/api/category/new_test_category/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMjIxfQ.puFqUPFfCEQVoynBhLDkwtflAVAuNXqiDSP09tCEmanZDEYxm2f0jSlFM17RtA9jIRmfJGHqp4SqTSxzY1zixQ' \
--data '{
    "category_new_data": {
        "name": "Updated name",
		"code": "new_code_for_category"
    }	
}'
```
Response
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 06 Apr 2024 10:02:59 GMT
Content-Length: 15

{
    "edited": true
}
```
#### Category delete
Request:
```
curl --location --request DELETE 'localhost:9999/api/category/new_code_for_category/delete' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMjIxfQ.puFqUPFfCEQVoynBhLDkwtflAVAuNXqiDSP09tCEmanZDEYxm2f0jSlFM17RtA9jIRmfJGHqp4SqTSxzY1zixQ'
```
Response
```json
HTTP/1.1 200 OK
Date: Sat, 06 Apr 2024 10:04:15 GMT
Content-Length: 0
```

### Product handlers

#### Product create
Request:
```
curl --location --request POST 'localhost:9999/api/product/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMjIxfQ.puFqUPFfCEQVoynBhLDkwtflAVAuNXqiDSP09tCEmanZDEYxm2f0jSlFM17RtA9jIRmfJGHqp4SqTSxzY1zixQ' \
--data '{
    "product": {
        "name": "New test product",
        "description": "",
        "category_codes": ["test_category_one"]
    }
}'
```
Response:
```json
HTTP/1.1 201 Created
Content-Type: application/json
Date: Sat, 06 Apr 2024 12:11:27 GMT
Content-Length: 21

{
    "product_id": "5351"
}
```
#### Product read

Get by Id request:
```
curl --location --request GET 'localhost:9999/api/product/5351'
```
Get by Id response:
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 06 Apr 2024 12:12:47 GMT
Content-Length: 105

{
    "product": {
        "id": 5351,
        "name": "New test product",
        "description": "",
        "category_codes": [
            "test_category_one"
        ]
    }
}
```

Get by category request:
```
curl --location --request GET 'localhost:9999/api/products/test_category_one'
```
Get by category response:
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 06 Apr 2024 12:18:15 GMT
Content-Length: 108

{
    "products": [
        {
            "id": 4,
            "name": "Test product Testtt",
            "description": "",
            "category_codes": [
                "test_category_one"
            ]
        }
    ]
}
```
#### Product update
Request:
```
curl --location --request PATCH 'localhost:9999/api/product/5351/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMjIxfQ.puFqUPFfCEQVoynBhLDkwtflAVAuNXqiDSP09tCEmanZDEYxm2f0jSlFM17RtA9jIRmfJGHqp4SqTSxzY1zixQ' \
--data '{
    "product_new_data": {
        "category_codes": ["test_category_two"]
    }
}'
```
Response:
```json
HTTP/1.1 200 OK
Date: Sat, 06 Apr 2024 12:14:21 GMT
Content-Length: 0
```
#### Product delete

Request:
```
curl --location --request DELETE 'localhost:9999/api/product/5351/delete' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZW1haWwuY29tIiwiZXhwIjoxNzEzMjYxMjIxfQ.puFqUPFfCEQVoynBhLDkwtflAVAuNXqiDSP09tCEmanZDEYxm2f0jSlFM17RtA9jIRmfJGHqp4SqTSxzY1zixQ'
```

Response:
```json
HTTP/1.1 200 OK
Date: Sat, 06 Apr 2024 12:15:41 GMT
Content-Length: 0
```