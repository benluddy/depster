# Depster

A command-line interface to operator dependency resolution.

# Usage

```sh
$ docker run -d -p 50051:50051 quay.io/operatorhubio/catalog:latest
$ make
$ ./depster resolve samples/*.yaml
```

# Releases

Signed binaries are available as [Github Releases](https://github.com/benluddy/depster/releases). The public key for signature verification is [https://github.com/benluddy.gpg](https://github.com/benluddy.gpg).
