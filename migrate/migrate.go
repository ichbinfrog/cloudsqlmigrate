package migrate

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type Op struct {
	Svc *sqladmin.Service
	Src *sqladmin.DatabaseInstance
	Dst *sqladmin.DatabaseInstance
}

func NewOp(ctx context.Context, srcProject, srcInstance, dstProject, dstInstance string, opts ...option.ClientOption) (*Op, error) {
	svc, err := sqladmin.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	src, err := svc.Instances.Get(srcProject, srcInstance).Do()
	if err != nil {
		return nil, err
	}

	dst, err := svc.Instances.Get(dstProject, dstInstance).Do()
	if err != nil {
		return nil, err
	}

	return &Op{
		Svc: svc,
		Src: src,
		Dst: dst,
	}, nil
}

func (op *Op) Preflight(ctx context.Context, preflight []Preflight) []error {
	errs := []error{}
	for _, p := range preflight {
		log.Printf("running preflight check %s", p.Name())
		if err := p.Check(op); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (op *Op) Prepopulate(ctx context.Context, postflight []PostFlight) []error {
	errs := []error{}
	for _, p := range postflight {
		log.Printf("prepopulating postflight checks: %s", p.Name())
		if err := p.Prepopulate(op); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (op *Op) PostFlight(ctx context.Context, postflight []PostFlight) []error {
	errs := []error{}
	for _, p := range postflight {
		log.Printf("running postflight check %s", p.Name())
		if err := p.Check(op); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (op *Op) Run(ctx context.Context, preflight []Preflight, postflight []PostFlight) []error {
	if errs := op.Preflight(ctx, preflight); len(errs) > 0 {
		return errs
	}

	if errs := op.Prepopulate(ctx, postflight); len(errs) > 0 {
		return errs
	}

	runId, err := op.sourceSnapshot(ctx)
	if err != nil {
		return []error{err}
	}

	if err := op.targetRestore(ctx, runId); err != nil {
		return []error{err}
	}

	if errs := op.PostFlight(ctx, postflight); len(errs) > 0 {
		return errs
	}
	return nil
}

func (op *Op) sourceSnapshot(ctx context.Context) (string, error) {
	id := uuid.NewString()
	task, err := op.Svc.BackupRuns.Insert(op.Src.Project, op.Src.Name, &sqladmin.BackupRun{
		Type:        "ON_DEMAND",
		BackupKind:  "SNAPSHOT",
		Description: id,
		Instance:    op.Src.Name,
		Location:    op.Src.Region,
	}).Do()
	if err != nil {
		return "", err
	}
	if err := WaitSQLAdminOp(ctx, op.Svc, op.Src.Project, task.Name, 10*time.Second); err != nil {
		return "", err
	}
	return id, nil
}

func (op *Op) targetRestore(ctx context.Context, id string) error {
	backups, err := op.Svc.BackupRuns.List(op.Src.Project, op.Src.Name).Do()
	if err != nil {
		return err
	}
	runID := int64(0)
	for _, backups := range backups.Items {
		if backups.Description == id {
			runID = backups.Id
		}
	}

	if runID == 0 {
		return fmt.Errorf("failed to find snapshot with ID %s in %s:%s", id, op.Src.Project, op.Src.Name)
	}

	task, err := op.Svc.Instances.RestoreBackup(op.Dst.Project, op.Dst.Name, &sqladmin.InstancesRestoreBackupRequest{
		RestoreBackupContext: &sqladmin.RestoreBackupContext{
			BackupRunId: runID,
			Project:     op.Src.Project,
			InstanceId:  op.Src.Name,
		},
	}).Do()
	if err != nil {
		return err
	}

	return WaitSQLAdminOp(ctx, op.Svc, op.Dst.Project, task.Name, 10*time.Second)
}
