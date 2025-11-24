json2go
-------

json2go to go provides a library and cli tool for
convening json strings to go struct definitions

    t := json2go. NewTransformer()
    typedef, _ := t.Transform(jsonStr, "TypeName")


cli interface:

    go install olexsmir.xyz/json2go/cmd/json2go@latest

    echo "{...}" | json2go
    json2go "{...}"
