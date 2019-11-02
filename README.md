## Prereq

- dep - https://golang.github.io/dep/docs/installation.html
- mongo - https://docs.mongodb.com/manual/installation/

## Install

- `dep ensure`

## References

- https://tour.golang.org
- https://golang.github.io/dep/docs/daily-dep.html
- https://github.com/tfogo/mongodb-go-tutorial

### Auth

- https://medium.com/@theShiva5/creating-simple-login-api-using-go-and-mongodb-9b3c1c775d2f
- https://blog.usejournal.com/authentication-in-golang-c0677bcce1a8
- https://github.com/urfave/negroni

### Mongo

- https://github.com/mongodb/mongo-go-driver
- https://docs.mongodb.com/manual/reference/operator/aggregation/lookup/index.html#examples
- https://medium.com/@HasstrupEzekiel/creating-indexes-with-the-new-mongo-go-driver-26310dbc3091
- https://docs.mongodb.com/manual/indexes/
- https://stackoverflow.com/questions/52235070/how-to-run-an-aggregate-query-via-mongo-go-driver-that-has-javascript-in-it
- \*\* https://stackoverflow.com/questions/56948324/how-to-write-bson-form-of-mongo-query-in-golang

#### `$lookup`

```bash
// mongo
> db.habits.insertOne({ "name": "Commit code today" });
{
	"acknowledged" : true,
	"insertedId" : ObjectId("5dbc949c729c5cf9dc3925b2")
}
> db.logs.insertOne({ "entry": "this app tho", habits: [ObjectId("5dbc949c729c5cf9dc3925b2")] });
{
	"acknowledged" : true,
	"insertedId" : ObjectId("5dbc94bb729c5cf9dc3925b3")
}
> db.logs.aggregate([{ $lookup: { from: "habits", localField: "habits", foreignField: "_id", as: "habits_info"  } }]);
{ "_id" : ObjectId("5dbc94bb729c5cf9dc3925b3"), "entry" : "this app tho", "habits" : [ ObjectId("5dbc949c729c5cf9dc3925b2") ], "habits_info" : [ { "_id" : ObjectId("5dbc949c729c5cf9dc3925b2"), "name" : "Commit code today" } ] }
```

### http

- https://golang.org/pkg/net/http
- https://www.alexedwards.net/blog/a-recap-of-request-handling

### Integ tests

- https://www.npmjs.com/package/cucumber-puppeteer
- https://github.com/GoogleChrome/puppeteer
- https://github.com/cucumber/cucumber-js

### Graphql

- https://gqlgen.com/getting-started/

### Deployment

- https://cloud.google.com/kubernetes-engine/docs/tutorials/hello-app
