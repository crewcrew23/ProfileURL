## For run needed
- ```./config/<config_name>.yaml```
- ```./storage/<name_db>``` — if use sqlite  
<br> db_name entered on config: ```./config/<config_name>.yaml```  
<br>

- config example  
```
env: (local, dev, prod) // chose one
addr: ":<port>"
storage_path: "<path_to_db>" // if sqlite, you need create dir ./storage and enter ./storage/<bd_name>.db
token_ttl: <time> // format 1s, 1m, 1h — LIFE TIME JWT
secret: <secret_key> // jwt secret
 ```

yaml
Копировать
Редактировать

---

## Run

### local
- ```make``` — exec build and migrations  
- ```./app --config=./config/<config_name>.yaml```

---

### docker
- ```docker build -t go-app .``` — build  
- ```docker run -p <external_port>:<internal_port> go-app``` — run  
- example: ```docker run -p 8080:8080 go-app```

---

## ⚙️ If config file name is changed

If you renamed your config (e.g. `local.yaml` → `dev.yaml`), you must edit **2 lines** in your `Dockerfile`:

### 1. Copy line
```
Was:
COPY ./config/local.yaml ./config/local.yaml

Become:
COPY ./config/dev.yaml ./config/dev.yaml

2. CMD run path

Was:
CMD ["/bin/sh", "-c", "./migrator --storage=./storage/url_profile.db --migration-path=./migrations && ./app --config=./config/local.yaml"]

Become:
CMD ["/bin/sh", "-c", "./migrator --storage=./storage/url_profile.db --migration-path=./migrations && ./app --config=./config/dev.yaml"]
```
🔁 Alternative (copy whole config folder)
Instead of specifying one file, copy the whole config folder:
```
COPY ./config/ ./config/
Then just change the config file path inside the CMD.
```
✅ Example Dockerfile using dev.yaml
```
FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go build -o app ./cmd/app
RUN go build -o migrator ./cmd/migrator

COPY ./config/dev.yaml ./config/dev.yaml

CMD ["/bin/sh", "-c", "./migrator --storage=./storage/url_profile.db --migration-path=./migrations && ./app --config=./config/dev.yaml"]
```
<br>


## DOC
## Рега
POST - ``` api/auth/sign-up ``` <br>
Принемает json : <br>
```
{
    "email":"test@gmail.com",
    "password":"123456"
}

```

Вернут 201 и Header Token с JWT при успегном создании пользователя или ошибку <br>

## Логин
POST - ``` api/auth/login ``` <br>
Принемает json : <br>
```
{
    "email":"test@gmail.com",
    "password":"123456"
}

```

Вернут 200 и Header Token с JWT при успегном логине пользователя или ошибку <br>

## Получение профиля другого пользователя
GET - ``` api/profile/{email} ``` <br>
аутентификация - не требуется <br>
Вернут 200 и пользователя если такой есть или ошибку <br>


## Получение своего профиля
GET - ``` api/profile ``` <br>
аутентификация - требуется (передать jwt)  <br>
Вернут 200 и профиль или ошибку <br>

## Добавление AboutME
POST - ``` api/profile/about ``` <br>
аутентификация - требуется (передать jwt) <br>
Принемает json : <br>
```
{
    "text":"hello i`m vasya"
}

```
Вернут 200 или ошибку <br>

## Добавление ссылки
POST - ``` api/profile/link ``` <br>
аутентификация - требуется (передать jwt) <br>
Принемает json (масив объектов): <br>
```
[
        {
            "link_name":"name",
            "link_color":"#color",
            "link_path":"path_to_link"
        },
        {
            "link_name":"name1",
            "link_color":"#color2",
            "link_path":"path_to_link3"
        }
]

```
Вернут 200 или ошибку <br>

## Обновление ссылки
PUT - ``` api/profile/link ``` <br>
аутентификация - требуется (передать jwt) <br>
Принемает json: <br>
```
{
    "link_id": 1,
    "link_name":"updated_name",
    "link_color":"updated_color",
    "link_path":"updated_path"
  
}


```
Вернут 200 или ошибку <br>


## Удаление ссылки
DELETE - ``` api/profile/link ``` <br>
аутентификация - требуется (передать jwt) <br>
Принемает json: <br>
``` {"id":3} ```
Вернут 200 или ошибку <br>