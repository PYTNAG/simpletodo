# Simple TODO API

### There is also [grpc](https://github.com/PYTNAG/simpletodo/tree/grpc) branch

## Contents

- [API End-points](#api-end-points)
    - [User related](#api-user)
    - [List related](#api-list)
    - [Task related](#api-task)
    - [Token related](#api-token)
- [Stack](#stack)

<a id="api-end-points"></a>
## API End-points

Optional fields marked with "optional" comment

<a id="api-user"></a>
### User related

- **POST /users**
    ```yaml
    # POST /users
    
    # Request body
    {
        "username": <string>,   
        "password": <string>    # printable ascii ; minimal length is 8
    }

    # Response body
    {
        "user_id": <int32>
    }
    ```
- **POST /users/login**
    ```yaml
    # POST /users/login

    # Request body
    {
        "username": <string>,   
        "password": <string>
    }

    # Response body
    {
        "session_id": <uuid>,
        "access_token": <string>,
        "access_token_expires_at": <time>, # RFC3339 with maximum 9 digits in fractional seconds, without trailing zeros in fractional seconds
        "refresh_token": <string>,
        "refresh_token_expires_at": <time>, # same as access_token_expires_at
        "user_id": <int32>
    }
    ```
- **PUT /users/\<int32\>**
    ```yaml
    # PUT /users/<int32>
    # Require header "authorization : bearer <access_token>"

    # Request body
    {
        "old_password": <string>,
        "new_password": <string>
    }

    # Without resposne body
    ```
- **DELETE /users/\<int32\>**
    ```yaml
    # DELETE /users/<int32>
    # Require header "authorization : bearer <access_token>"

    # Without request body

    # Without response body
    ```

<a id="api-list"></a>
### List related

- **GET /users/\<int32\>/lists**
    ```yaml
    # GET /users/<int32>/lists
    # Require header "authorization : bearer <access_token>"

    # Without request body

    # Resposne body
    {
        "lists": [
            {
                "id": <int32>,
                "header": <string>
            }...
        ]
    }
    ```

- **POST /users/\<int32\>/lists**
    ```yaml
    # POST /users/<int32>/lists
    # Require header "authorization : bearer <access_token>"

    # Request body
    {
        "header": <string>
    }

    # Without response body
    ```

- **DELETE /users/\<int32\>/lists/\<int32\>**
    ```yaml
    # POST /users/<int32>/lists/<int32>
    # Require header "authorization : bearer <access_token>"

    # Without request body

    # Without response body
    ```

<a id="api-task"></a>
### Task related

- **GET /users/\<int32\>/lists/\<int32\>/tasks**
    ```yaml
    # POST /users/<int32>/lists/<int32>/tasks
    # Require header "authorization : bearer <access_token>"

    # Without request body

    # Response body
    {
        "tasks": [
            {
                "id": <int32>,
                "list_id": <int32>,
                "parent_task": <int32>, # optional ; min = 1
                "task": <string>,
                "complete": <bool>
            }...
        ]
    }
    ```

- **POST /users/\<int32\>/lists/\<int32\>/tasks**
    ```yaml
    # POST /users/<int32>/lists/<int32>/tasks
    # Require header "authorization : bearer <access_token>"

    # Request body
    {
        "parent_task": <int32>, # optional ; min = 1
        "task": <string>
    }

    # Response body
    {
        "created_task_id": <int32>
    }
    ```

- **PUT /users/\<int32\>/lists/\<int32\>/tasks/\<int32\>**
    ```yaml
    # POST /users/<int32>/lists/<int32>/tasks/<int32>
    # Require header "authorization : bearer <access_token>"

    # Request body
    {
        "type": <string>, # "TEXT" or "CHECK"
        "text": <string>, # required if type == TEXT
        "check": <bool> # required if type == CHECK
    }

    # Without response body
    ```

- **DELETE /users/\<int32\>/lists/\<int32\>/tasks/\<int32\>**
    ```yaml
    # DELETE /users/<int32>/lists/<int32>/tasks/<int32>
    # Require header "authorization : bearer <access_token>"

    # Without request body

    # Without response body
    ```

<a id="api-token"></a>
### Token related

- **POST /tokens/refresh_access**
    ```yaml
    # POST /tokens/refresh_access

    # Request body
    {
        "refresh_token": <string>
    }

    # Response body
    {
        "access_token": <string>,
        "access_token_expires_at": <time> # RFC3339 with maximum 9 digits in fractional seconds, without trailing zeros in fractional seconds
    }
    ```

<a id="stack"></a>
## Stack

### Web server

- [gin](https://github.com/gin-gonic/gin)
- [paseto](https://github.com/aidantwoods/go-paseto)

### Data Base (PostgreSQL)

- [lib/pq](https://github.com/lib/pq)
- [migrate](https://github.com/golang-migrate/migrate)
- [sqlc](https://github.com/sqlc-dev/sqlc)
- [mock](https://github.com/uber-go/mock)

### Testing

- [mock](https://github.com/uber-go/mock)
- [testify](https://github.com/stretchr/testify)

### Deploy

- Docker
- [viper](https://github.com/spf13/viper)

### Types

- [google/uuid](github.com/google/uuid)
