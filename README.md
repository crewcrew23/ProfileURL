## For run needed
 - ```./config/<config_name>.yaml```
 - ```./storage/<name_db>``` -- if use sqlite
 <br> db_name entered on config ```./config/<config_name>.yaml ```
<br>
- config example <br>

```
env: (local, dev, prod) // chose one
addr: ":<port>"
storage_path: "<path_to_db>" //if sqlite you need create dir ./storage  and entered ./storage/<bd_name>
token_ttl: <time> //format 1s, 1m, 1h LIFE TIME JWT
secret: <secret_key> // jwt secret

```

## Run
- ``` make ``` exec build and migrations
- ``` ./app --config=<path_to_config>```