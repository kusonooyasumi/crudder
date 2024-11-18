![image](https://github.com/user-attachments/assets/e1ad5e8e-4c3b-4fca-b701-7d502b53c237)


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