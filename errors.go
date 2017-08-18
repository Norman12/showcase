package main

import "errors"

var (
	ErrDatabase              = errors.New("Database error")
	ErrNoTheme               = errors.New("Cannot render page with no template")
	ErrNoThemes              = errors.New("No templates found")
	ErrInvalidTheme          = errors.New("Template is invalid")
	ErrNoSlug                = errors.New("No slug provided")
	ErrContentExists         = errors.New("Content with this slug already exists")
	ErrProjectExists         = errors.New("Project with this slug already exists")
	ErrMediaNotSupported     = errors.New("This media is not supported yet")
	ErrSetupEmpty            = errors.New("Please fill in required fields")
	ErrConfigurationTimedOut = errors.New("Configuration timed out")
)
