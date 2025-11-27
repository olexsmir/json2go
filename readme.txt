json2go
-------

json2go to go provides a library and cli tool for
converting json strings to go struct definitions

    t := json2go. NewTransformer()
    typedef, err := t.Transform(`{"json": true}`, "TypeName")


cli interface:

    go install olexsmir.xyz/json2go/cmd/json2go@latest

    echo '{"id": 1, "name": "Alice"}' | json2go
    json2go '{"id": 1, "name": "Alice"}'
