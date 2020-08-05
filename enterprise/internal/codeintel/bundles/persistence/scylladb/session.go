package scylladb

import "github.com/gocql/gocql"

var session *gocql.Session

func init() {
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "lsif"
	cluster.Consistency = gocql.One

	var err error
	if session, err = cluster.CreateSession(); err != nil {
		panic(err.Error())
	}
}
