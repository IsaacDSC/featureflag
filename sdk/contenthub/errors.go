package contenthub

import "errors"

var ErrInvalidStrategy = errors.New("contenthub with strategy required sessionID")
var ErrNotFoundContenthub = errors.New("not found contenthub")
