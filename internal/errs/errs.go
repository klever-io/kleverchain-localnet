package errs

import "errors"

var (
	ErrPrereq                 = errors.New("prerequisite not met")
	ErrStateConflict          = errors.New("state conflict")
	ErrInvalidInput           = errors.New("invalid input")
	ErrConsensusGroupTooLarge = errors.New("consensus group size exceeds validator count")
	ErrComposeMissing         = errors.New("docker-compose file not found")
	ErrDockerFailure          = errors.New("docker command failed")
	ErrKeysIncomplete         = errors.New("keys directory is in an incomplete state")
)
