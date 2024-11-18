# API Auth Testing Tool written in GO

crudder is a simple tool written in go for the sole purpose of testing API endpoints for authentication as well as authorization.

## Usage

`./crudder cru -s api.example.com endpoints.txt 10`

`cru` - specifies the request methods to be sent (c:POST,r:GET,u:PUT,d:DELETE)

`-s` - api domain/subdomain in a comma separated list (`-sf` can be used to input list of subdomains)

`endpoints.txt` - list of endpoints to test

`10` - number of concurrent requests

optional: 

`output.txt`  - specify file for output
