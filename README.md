# FeIDo Database Server
Demo server mimicking the API of Interpol's I-Checkit service.
This network service enables the FeIDo credential services to query if a given
eID has been reported as stolen/blocked.

## Build and Run Instructions
**Note:** Please fetch this repo as a submodule of the umbrella repo: https://github.com/feido-token/feido.
Otherwise, the symoblic links to the FeIDo protobuf files will not resolve.

0. Update protobuf generated file (if required):
    ```
    protoc --go_out=. ./feido-proto/feido-database.proto
    ```

1. Generate demo database:
    ```
    sqlite3 eids-database.db < create_feido_demo_db.sql
    ```

2. Build service:
    ```
    cd feido-dbsrv/
    go build
    ```

3. Run service:
    ```
    ./feido-dbsrv -db ../eids-database.db 127.0.0.1 4711
    ```

4. Run FeIDo authentication (see other README/s)

## Database Entries
The database entries must be added in hex encoding, i.e., if the MRZ for instance
says:

"ABCDE1234", "E<<", "P<"

you insert as ASCII hex encoding:

x'414243444531323334', x'453c3c', x'503c'

