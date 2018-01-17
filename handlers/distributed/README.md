## Distributed queries


post to `/distributed/add` to create a new distributed query for a node.  Multiple nodes may
be specified in the same post request

```bash
curl https://osquerydev.phishinghole.io:8888/distributed/add -d '{"nodes":[{"node_key":"ls": ["select * from users;", "select * from logged_in_users;"]}]}'
```

/distributed/add


post data in the following format
```json
{
  "nodes":
  [
    {
      "node_key": "llngzieyoh43e0133ols",
      "queries": [
        "select * from users;",
        "select * from etc_hosts"
      ]
    },
    {
      "node_key": "llngzieyoh43e0133ldf",
      "queries": [
        "select * from users",
        "select * from dns_resolvers"
      ]
    }
  ]
}
```