package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsers_Filter(t *testing.T) {
	var users Users
	users = append(users, User{ID: 1})
	users = append(users, User{ID: 2})

	filtered := users.Filter(func(user User) bool {
		return user.ID > 1
	})

	assert.Len(t, filtered, 1)
	assert.Equal(t, int64(2), filtered[0].ID)
}
