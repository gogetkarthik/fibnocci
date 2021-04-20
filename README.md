# Fibonacci service with postgres

## How to start the service
Assuming docker is installed in the local machine. 

Build local docker image
```bash 
make docker-build-local
```

Bring the postgres and service 
```bash
make service-up
```

Run the DDL preset in the ./sqls dir
```
./sqls --> run DDL and insert command
```

verify the fib sequence by hitting the url in the browser or using curl

 [verify_link](http://localhost:9001/fib/10)

## Unit test and vet other commands are available in Makefile. 
```bash
make test
```

```bash
make vet
```