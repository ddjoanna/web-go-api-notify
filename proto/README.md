## Documentation

- [Protobuf 規格文件](doc/README.md)

## Prerequisites

- Install [protoc v25](https://grpc.io/docs/protoc-installation/)

  ```shell
  # 確認 protoc 版本
  protoc --version
  brew install protobuf
  # 確認 protoc-gen-doc 是否可用
  which protoc-gen-doc
  go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
  ```

- Install [make](https://www.gnu.org/software/make/)

  ```shell
  brew install make
  ```

- Install [golang 1.22](https://golang.org/dl/)

  ```shell
  brew install golang@1.22
  ```

- Install dependencies

  ```shell
  make install
  ```

## Code Generation

After modifying the protobuf files, you need to regenerate the code:

```shell
make gen
```
