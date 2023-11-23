# Simple TODO app

## Contents

- [RPCs](#rpc-end-points)
    - [User related](#rpc-user)
    - [List related](#rpc-list)
    - [Task related](#rpc-task)
    - [Token related](#rpc-token)
- [Stack](#stack)

<a id="api-end-points"></a>
## RPCs

If no response message is specified, [google.protobuf.Empty](github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/empty.proto) is used instead.
```
message Empty{}
```

<a id="rpc-user"></a>
### User related

- **CreateUser / Unary**
    ```
    message CreateUserRequest {
        string username = 1;
        string password = 2;
    }

    message CreateUserResponse {
        int32 user_id = 1;
    }
    ```
- **DeleteUser / Unary**
    ```
    message DeleteUserRequest {
        int32 user_id = 1;
    }
    ```
- **RehashUser / Unary**
    ```
    message RehashUserRequest {
        int32 user_id = 1;
        string new_password = 2;
    }
    ```
- **LoginUser / Unary**
    ```
    message LoginUserRequest {
        string username = 1;
        string password = 2;
    }

    message LoginUserResponse {
        int32 user_id = 1;
        string session_id = 2;
        string access_token = 3;
        string refresh_token = 4;
        google.protobuf.Timestamp access_token_expires_at = 5;
        google.protobuf.Timestamp refresh_token_expires_at = 6;
    }
    ```

<a id="rpc-list"></a>
### List related

- **CreateList / Unary**
    ```
    message CreateListRequest {
        int32 user_id = 1;
        string header = 2;
    }
    ```

- **GetLists / Server-streaming**
    ```
    message GetListsRequest {
        int32 user_id = 1;
    }

    message List {
        int32 id = 1;
        string header = 2;
    }
    ```

- **DeleteList / Unary**
    ```
    message DeleteListRequest {
        int32 list_id = 1;
    }
    ```

<a id="rpc-task"></a>
### Task related

- **CreateTask / Unary**
    ```
    message CreateTaskRequest {
        int32 list_id = 1;
        string task_text = 2;
        optional int32 parent_task_id = 3;
    }

    message CreateTaskResponse {
        int32 task_id = 1;
    }
    ```

- **GetTasks / Server-streaming**
    ```
    message GetTasksRequest {
        int32 list_id = 1;
    }

    message Task {
        int32 task_id = 1;
        int32 list_id = 2;
        string text = 3;
        bool check = 4;
        optional int32 parent_task_id = 5;
    }
    ```

- **DeleteTask / Unary**
    ```
    message DeleteTaskRequest {
        int32 task_id = 1;
    }
    ```

- **UpdateTask / Unary**
    ```
    enum UpdateType {
        UNSET = 0;
        TEXT = 1;
        CHECK = 2;
    }

    message UpdateTaskRequest {
        int32 task_id = 1;
        UpdateType type = 2;
        optional string new_text = 3;
    }
    ```

<a id="rpc-token"></a>
### Token related

- **RefreshAccessToken / Unary**
    ```
    message RefreshAccessTokenRequest {
        string refresh_token = 1;
    }

    message RefreshAccessTokenResponse {
        string access_token = 1;
        google.protobuf.Timestamp access_token_expired_at = 2;
    }
    ```

<a id="stack"></a>
## Stack

### Web server

- [gRPC](https://grpc.io/docs/languages/go/basics)
- [paseto](https://github.com/aidantwoods/go-paseto)

### Logging

- [zerolog](https://github.com/rs/zerolog)

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