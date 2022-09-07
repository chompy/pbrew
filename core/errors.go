package core

import "errors"

var (
	ErrBrewNotInstalled        = errors.New("homebrew not installed")
	ErrServiceNotFound         = errors.New("service not found")
	ErrServiceNotInstalled     = errors.New("service not installed")
	ErrInvalidService          = errors.New("invalid service")
	ErrInvalidDef              = errors.New("invalid definition")
	ErrTemplateNotFound        = errors.New("service config template not found")
	ErrServiceAlreadyRunning   = errors.New("service already running")
	ErrServiceNotRunning       = errors.New("service not running")
	ErrServiceNoName           = errors.New("service has no name")
	ErrServiceReloadNotDefined = errors.New("service reload command not defined")
	ErrServiceNotMySQL         = errors.New("service must be based on mysql")
	ErrServiceNotSolr          = errors.New("service must be based on solr")
	ErrServiceDefNotDefined    = errors.New("service definition not defined")
	ErrPHPExtNotFound          = errors.New("php extension not found")
	ErrProjectNotFound         = errors.New("project not found")
)
