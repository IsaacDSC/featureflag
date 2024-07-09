# Feature Flag


### Example 2
```
session_id: [01J29MKZ7H4V8A65M536CKQ5HG],
percent: 50%,

QUERO DIZER
50% true -> session_id 01J29MKZ7H4V8A65M536CKQ5HG
50% false -> session_id 01J29MKZ7H4V8A65M536CKQ5HG
```

### Example 3

```
session_id: [01J29MKZ7H4V8A65M536CKQ5HG, 01J29MSCVVPH8CG6R0422NM3ME],
percent: 50%,

QUERO DIZER
50% true -> session_id 01J29MKZ7H4V8A65M536CKQ5HG
50% true -> session_id 01J29MSCVVPH8CG6R0422NM3ME
50% false -> session_id <other>
```

mockgen -source=./internal/domain/interfaces/featureflag_interfaces.go -destination=./internal/mocks/featureflag_interfaces.mock.go -package=mock 