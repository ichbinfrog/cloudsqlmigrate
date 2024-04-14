package migrate

import (
	"context"
	"log"
	"time"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"k8s.io/apimachinery/pkg/util/wait"
)

func WaitSQLAdminOp(ctx context.Context, sql *sqladmin.Service, project string, operation string, interval time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, interval, 10*time.Hour, true, func(ctx context.Context) (done bool, err error) {
		o, err := sql.Operations.Get(project, operation).Do()
		if err != nil {
			return false, err
		}

		log.Printf("operation(%s:%s:%s) is %s, retrying in %s\n", o.OperationType, project, operation, o.Status, interval)
		if o.Status == "DONE" {
			start, _ := time.Parse(time.RFC3339, o.StartTime)
			end, _ := time.Parse(time.RFC3339, o.EndTime)
			log.Printf("operation(%s:%s:%s) completed in %s\n", o.OperationType, project, operation, end.Sub(start).String())
			return true, nil
		}

		return false, nil
	})
}
