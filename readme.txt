json2go
-------

json2go provides a library and cli tool for converting json strings to go struct definitions.

library Usage:

    // with json tags
    code, err := json2go.Transform("User", `{"name": "Alice"}`, true)

    // without json tags
    code, err := json2go.Transform("User", `{"name": "Alice"}`, false)


cli interface:

    go install olexsmir.xyz/json2go/cmd/json2go@latest

    echo '{"id": 1, "name": "Alice"}' | json2go
    json2go '{"id": 1, "name": "Alice"}'
    json2go --help
