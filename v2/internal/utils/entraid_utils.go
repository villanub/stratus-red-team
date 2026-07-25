package utils

import (
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// RetryOnGraphEventualConsistency retries fn against Microsoft Graph's
// well-documented eventual consistency for freshly-created directory
// objects: a Graph call that references an object by ID moments after
// creating it can fail with "Resource does not exist or one of its
// queried reference-property objects are not present" until the object
// has propagated across Graph's replicas. fn should return that exact
// error class for retryable failures; any other error is returned
// immediately without retrying. Mirrors WaitForAndAssumeAWSRole's
// eventual-consistency retry pattern in aws_utils.go, same dependency
// (github.com/cenkalti/backoff/v4), already vendored for that use.
func RetryOnGraphEventualConsistency(fn func() error) error {
	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = 2 * time.Second
	backoffStrategy.Multiplier = 2
	backoffStrategy.MaxInterval = 15 * time.Second
	backoffStrategy.MaxElapsedTime = 2 * time.Minute
	return backoff.Retry(func() error {
		err := fn()
		if err == nil {
			return nil
		}
		if isGraphEventualConsistencyError(err) {
			return err
		}
		return backoff.Permanent(err)
	}, backoffStrategy)
}

func isGraphEventualConsistencyError(err error) bool {
	return strings.Contains(err.Error(), "does not exist or one of its queried reference-property objects are not present")
}
