package apns

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	t.Run("retry=succcess", func(t *testing.T) {
		var attempts int
		err := retry(func() error {
			attempts++
			if attempts < 3 {
				return connError("error")
			}
			return nil
		}, 4)
		assert.NoError(t, err)
		assert.Equal(t, attempts, 3)
	})

	t.Run("retry=false", func(t *testing.T) {
		err := retry(func() error {
			return errors.New("error")
		}, 4)

		assert.Error(t, err)
	})

	t.Run("retry=maxAttempts", func(t *testing.T) {
		err := retry(func() error {
			return connError("error")
		}, 1)
		assert.Error(t, err)
	})
}
