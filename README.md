# depster

A command-line interface to operator dependency resolution.

# Usage

```sh
$ docker run -d -p 50051:50051 quay.io/operatorhubio/catalog:latest
$ make
$ ./depster resolve samples/*.yaml
```
