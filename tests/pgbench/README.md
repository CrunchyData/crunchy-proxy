# Testing with pgbench

Pgbench can be used to run general benchmark tests against the proxy.

Currently, the following tests are provided as part of this test suite:

* **simple-load-tests.sql** - These tests are meant to simply perform
  repetitive read only operations against the proxy. This test has two simple
  queries one that is annotated with the 'read' annotation and one that is not.
  The purpose of this test is to send the same number of queries to both backend
  nodes per pgbench transaction. 

* **concurrency-tests.sql** - These tests perform a series of simple insert, 
  update and read operations against the proxy.

## Running crunchy-proxy

First make sure that both the 'master' and 'replica' PostgreSQL nodes are
running. Then it is necessary to initialize the the database for use with
pgbench.

These tests assume that these nodes are available at the following domains:

* master.crunchy.lab
* replica.crunchy.lab

Initialize database:

```
$> cd tests/pgbench
$> ./init_tests.sh
```

After initializing the test database, start up the proxy using the provided
configuration. Running the proxy from source requires doing so from the
project's root directory.

Run the proxy:

```
$> go run main.go start --config=./tests/pgbench/config.yaml
```

## Running the tests

Tests must be run from the `tests/pgbench` directory.

For load testing:

```
$> cd tests/pgbench
$> ./run-load-tests.sh
```

For concurrency testing:

```
$> cd tests/pgbench
$> ./run-concurrency-tests.sh
```
