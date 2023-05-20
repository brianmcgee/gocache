package nats

import "github.com/juju/errors"

const (
	ErrTagsNotSupported             = errors.ConstError("tags are not supported in a nats KV store")
	ErrPerKeyExpirationNotSupported = errors.ConstError("expiration can only be configured store wide, not on a per key basis")
)
