package cassandra

import (
	"context"
	"fmt"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/sirupsen/logrus"
)

// queryObserver implements gocql.QueryObserver to log details about every
// query executed against Cassandra. This is intended for debugging and
// should be disabled in production due to the volume of log output.
type queryObserver struct {
	logger *wrappedLogger
}

func (o *queryObserver) ObserveQuery(_ context.Context, q gocql.ObservedQuery) {
	latency := q.End.Sub(q.Start)

	fields := logrus.Fields{
		"keyspace":  q.Keyspace,
		"statement": sanitizeStatement(q.Statement),
		"latency":   latency.String(),
		"rows":      q.Rows,
		"attempt":   q.Attempt,
	}

	if q.Host != nil {
		fields["host"] = q.Host.ConnectAddress().String()
		fields["host_id"] = q.Host.HostID()
		fields["datacenter"] = q.Host.DataCenter()
	}

	if q.Err != nil {
		if driverLogLevelMap[o.logger.level] > driverLogLevelMap[DriverLogLevelError] {
			return
		}
		fields["error"] = q.Err.Error()
		o.logger.logger.WithFields(fields).Error("gocql-driver: cassandra query failed")
	} else {
		if driverLogLevelMap[o.logger.level] > driverLogLevelMap[DriverLogLevelDebug] {
			return
		}
		o.logger.logger.WithFields(fields).Debug("gocql-driver: cassandra query executed")
	}
}

// sanitizeStatement collapses whitespace in a CQL statement for cleaner log output.
func sanitizeStatement(stmt string) string {
	parts := strings.Fields(stmt)
	return fmt.Sprintf("%s", strings.Join(parts, " "))
}
