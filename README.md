# golang-rest-api-demo

This is an example application for REST APIs in golang with gorilla mux router and http package from standard library.
The storage used is MySQL.

### API documentation 

> ___
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
>       <pre>{
>         Title     (string)
>         desc      (string)
>         content   (string)
>     }</pre>
> 
> - PUT /article/{id}
>   - Update an existing article DB
>   - query param : id (article id from GET API)
>   - payload :
>       <pre>{
>         Title     (string)
>         desc      (string)
>         content   (string)
>     }</pre>
> 
> - DELETE /article/{id}
>   - Deletes an entry from DB
>   - query param : id (article id from GET API)
> 
> - GET /article/{id}
>   - Retrieves article data from DB for a given ID
>   - query param : id (article id from GET API)
> ___
> 