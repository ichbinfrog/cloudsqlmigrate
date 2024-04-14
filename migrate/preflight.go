package migrate

import (
	"errors"
	"fmt"
)

var (
	ErrPreflight                = errors.New("preflight_error")
	ErrPreflightVersionMismatch = fmt.Errorf("%w - mismatch database version", ErrPreflight)
	ErrPreflightStatus          = fmt.Errorf("%w - instance is not running", ErrPreflight)
)

type Preflight interface {
	Name() string
	Check(op *Op) error
}

type PreflightVersion struct{}

func (p PreflightVersion) Name() string {
	return "check_database_version"
}

func (p PreflightVersion) Check(op *Op) error {
	if op.Src.DatabaseVersion != "" && op.Src.DatabaseVersion == op.Dst.DatabaseVersion {
		return nil
	}
	return fmt.Errorf("%w: source_version=%s, target_version=%s", ErrPreflightVersionMismatch, op.Src.DatabaseVersion, op.Dst.DatabaseVersion)
}

type PreflightStatus struct{}

func (p PreflightStatus) Name() string {
	return "check_database_state"
}

func (p PreflightStatus) Check(op *Op) error {
	if op.Src.State == "RUNNABLE" && op.Dst.State == op.Src.State {
		return nil
	}
	return fmt.Errorf("%w: source_state=(%s:%v), target_state=(%s:%v)", ErrPreflightStatus, op.Src.State, op.Src.SuspensionReason, op.Dst.State, op.Dst.SuspensionReason)
}
