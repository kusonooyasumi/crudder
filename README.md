![image](https://github.com/user-attachments/assets/bcad1800-b056-4fee-9743-c88c21602253)


# API Auth Testing Tool written in GO

crudder is a simple tool written in go for the sole purpose of testing API endpoints for authentication as well as authorization.

## Usage

`./crudder cru -s api.example.com  -e endpoints.txt -r 10`

`cru` - specifies the request methods to be sent (c:POST,r:GET,u:PUT,d:DELETE)

`-s` - api domain/subdomain in a comma separated list (`-sf` can be used to input list of subdomains)

`-e` - list of endpoints to test

`-r` - number of concurrent requests

optional: 

`-o`  - specify file for output
