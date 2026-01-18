# Migration Notes

API specifications for migrating ncobase from ncore v0.1.x to v0.2.0.

## API Changes

v0.2.0 replaces direct internal field access with unified public APIs:

| Resource          | v0.1.x                   | v0.2.0                                     |
|-------------------|--------------------------|--------------------------------------------|
| **SQL Database**  | `d.Conn.DBM.Master()`    | `d.GetMasterDB()`                          |
|                   | `d.Conn.DBM.Slave()`     | `d.GetSlaveDB()`                           |
| **Redis**         | `d.Conn.RC`              | `d.GetRedis()`                             |
| **MongoDB**       | `d.Conn.MGM.Master()`    | `d.GetMongoDatabase(name, readOnly)`       |
|                   | `d.Conn.MGM.Slave()`     | `d.GetMongoCollection(db, coll, readOnly)` |
| **Elasticsearch** | `d.Conn.ES`              | `d.GetElasticsearch()`                     |
| **OpenSearch**    | `d.Conn.OS`              | `d.GetOpenSearch()`                        |
| **Meilisearch**   | `d.Conn.MS`              | `d.GetMeilisearch()`                       |
| **Search**        | `r.data.Search()`        | `r.sc.Search()`                            |
| **Index**         | `r.data.IndexDocument()` | `r.sc.Index()`                             |

## Data Layer Patterns

### SQL Database (Ent ORM)

```go
type Data struct {
    *data.Data
    EC     *ent.Client // master
    ECRead *ent.Client // slave
}

func New(conf *config.Data, env ...string) (*Data, func(), error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, nil, err
    }

    masterDB := d.GetMasterDB()
    entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate, env...)
    if err != nil {
        return nil, cleanup, err
    }

    var entClientRead *ent.Client
    if readDB, err := d.GetSlaveDB(); err == nil && readDB != masterDB {
        entClientRead, _ = newEntClient(readDB, conf.Database.Master, false, env...)
    }
    if entClientRead == nil {
        entClientRead = entClient
    }

    return &Data{Data: d, EC: entClient, ECRead: entClientRead}, cleanup, nil
}
```

### Redis

```go
// Type assertion required
redisClient := d.GetRedis().(*redis.Client)
```

### MongoDB

```go
// Get database/collection
db, err := d.GetMongoDatabase("mydb", false)      // master
dbRead, err := d.GetMongoDatabase("mydb", true)   // slave
coll, err := d.GetMongoCollection("mydb", "users", false)

// Transaction
err = d.WithMongoTransaction(ctx, func(sessCtx any) error {
    mongoCtx := sessCtx.(mongo.SessionContext)
    return nil
})

// Cache client (optional)
mongoManager := d.GetMongoManager()
if mgm, ok := mongoManager.(interface{ Master() *mongo.Client }); ok {
    mongoMaster = mgm.Master()
}
```

### Search Engine

```go
sc := nd.NewSearchClient(d.Data)

// Check nil before use
if sc != nil {
    engines := sc.GetAvailableEngines()
    if len(engines) > 0 {
        sc.Index(ctx, &search.IndexRequest{Index: "users", Document: user})
        sc.Search(ctx, &search.SearchRequest{Index: "users", Query: "John"})
    }
}
```

## Repository Pattern

```go
type userRepository struct {
    data         *data.Data
    sc *search.Client
    redisClient  *redis.Client
}

func NewUserRepository(d *data.Data) UserRepositoryInterface {
    return &userRepository{
        data:         d,
        sc: nd.NewSearchClient(d.Data),
        redisClient:  d.GetRedis().(*redis.Client),
    }
}

func (r *userRepository) Create(ctx context.Context, body *structs.CreateUserBody) (*ent.User, error) {
    user, err := r.data.GetMasterEntClient().User.Create().SetUsername(body.Username).Save(ctx)
    if err != nil {
        return nil, err
    }
    if r.sc != nil {
        r.sc.Index(ctx, &search.IndexRequest{Index: "users", Document: user})
    }
    return user, nil
}

func (r *userRepository) Get(ctx context.Context, id string) (*ent.User, error) {
    return r.data.GetSlaveEntClient().User.Query().Where(userEnt.IDEQ(id)).Only(ctx)
}
```

## Important Notes

1. **Redis Type Assertion**: `GetRedis()` returns `any`, must assert to `(*redis.Client)` before use
2. **SearchClient Nil Check**: Search engine is optional, must check nil before use
3. **Read/Write Separation**: Write operations use Master, read operations use Slave (auto-fallback to Master)

## Migration Checklist

- [ ] Replace `d.Conn.*` with corresponding `d.Get*()` methods
- [ ] Add type assertion `(*redis.Client)` for Redis
- [ ] Add `sc` field to repositories and initialize
- [ ] Add nil check before SearchClient calls
- [ ] Update import path from `data/databases/cache` to `data/cache`
