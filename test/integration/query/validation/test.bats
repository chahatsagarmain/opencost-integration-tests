setup() {
    : # nothing to set up
}

teardown() {
    : # nothing to tear down
}

@test "query api : validation of api response" {
    go test -v ./test/integration/query/validation/validation_test.go
}