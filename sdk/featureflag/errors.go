package featureflag

import "errors"

var ErrInvalidStrategy = errors.New("featureflag with strategy required sessionID")
var ErrNotFoundFeatureFlag = errors.New("not found featureflag")
