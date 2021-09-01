package core

import "errors"

var (
	ErrBrewNotInstalled        = errors.New("homebrew not installed")
	ErrServiceNotFound         = errors.New("service not found")
	ErrServiceNotInstalled     = errors.New("service not installed")
	ErrInvalidDef              = errors.New("invalid definition")
	ErrNginxTemplateNotFound   = errors.New("nginx template not found")
	ErrServiceAlreadyRunning   = errors.New("service already running")
	ErrServiceNotRunning       = errors.New("service not running")
	ErrServiceReloadNotDefined = errors.New("service reload command not defined")
	ErrServiceNotMySQL         = errors.New("service must be based on mysql")
)
