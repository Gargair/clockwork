package service

import "errors"

var ErrNoActiveTimer = errors.New("service: no active timer")
var ErrCategoryCycle = errors.New("service: category cycle detected")
var ErrCrossProjectParent = errors.New("service: parent category belongs to a different project")
var ErrInvalidParent = errors.New("service: invalid parent category")


