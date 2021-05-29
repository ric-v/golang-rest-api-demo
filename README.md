# golang-rest-api-demo

This is an example application for REST APIs in golang with gorilla mux router and http package from standard library.
The storage used is MySQL.

## API documentation 

> ___
>
> - ANY /
>   - homepage with this info
>
> - GET /articles
>   - retrives all articles from DB
>   - query params : id (last ID from previous GET call for pagination), limit (max entry per page)
>   - response : list of articles
>
> - POST /article
>   - Add new article to DB
>   - payload :
>
> ```json
>           {
>               "Title"     (string)
>               "desc"      (string)
>               "content"   (string)
>           }
> ```
>
> - PUT /article/{id}
>   - Update an existing article DB
>   - query param : id (article id from GET API)
>   - payload :
>
> ```json
>           {
>               "Title"     (string)
>               "desc"      (string)
>               "content"   (string)
>           }
> ```
>
> - DELETE /article/{id}
>   - Deletes an entry from DB
>   - query param : id (article id from GET API)
>
> - GET /article/{id}
>   - Retrieves article data from DB for a given ID
>   - query param : id (article id from GET API)
>
> ___

## Installation

installing and running mysql
> docker run --name mysql-instance -p 3306:3036 -e MYSQL_ROOT_PASSWORD=my-secret-pw -d mysql
>
> docker exec -it mysql-instance -p 5432:5432 mysql -u myuser -p

installing running psql
> docker run --name postgres-instance -e POSTGRES_PASSWORD=mysecretpassword -d postgres
>
> docker exec -it some-postgres psql -U postgres

create new table

> ```sql
> CREATE TABLE articles(
>   id INTEGER PRIMARY KEY,
>   title TEXT,
>   descr TEXT,
>   content TEXT
> )
> ```

run the main.go file from the cloned repo to get access to REST APIs.
