package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	id := NewUserID("user-id-1")
	username := "testuser"
	email := "test@example.com"
	passwordHash := []byte("hashedpassword")
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	user := NewUser(id, username, email, passwordHash, createdAt, updatedAt)

	assert.NotNil(t, user, "NewUser は nil を返すべきではない")
	assert.Equal(t, id, user.id, "NewUser は正しい ID を設定するべき")
	assert.Equal(t, username, user.username, "NewUser は正しい username を設定するべき")
	assert.Equal(t, email, user.email, "NewUser は正しい email を設定するべき")
	assert.Equal(t, passwordHash, user.passwordHash, "NewUser は正しい passwordHash を設定するべき")
	assert.Equal(t, createdAt, user.createdAt, "NewUser は正しい createdAt を設定するべき")
	assert.Equal(t, updatedAt, user.updatedAt, "NewUser は正しい updatedAt を設定するべき")
}

func TestUser_Getters(t *testing.T) {
	id := NewUserID("user-id-2")
	username := "anotheruser"
	email := "another@example.com"
	passwordHash := []byte("anotherhashedpassword")
	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := time.Now().Add(-24 * time.Hour)

	user := NewUser(id, username, email, passwordHash, createdAt, updatedAt)

	assert.Equal(t, id, user.ID(), "ID() は正しい ID を返すべき")
	assert.Equal(t, username, user.Username(), "Username() は正しい username を返すべき")
	assert.Equal(t, email, user.Email(), "Email() は正しい email を返すべき")
	assert.Equal(t, passwordHash, user.PasswordHash(), "PasswordHash() は正しい passwordHash を返すべき")
	assert.Equal(t, createdAt, user.CreatedAt(), "CreatedAt() は正しい createdAt を返すべき")
	assert.Equal(t, updatedAt, user.UpdatedAt(), "UpdatedAt() は正しい updatedAt を返すべき")
}

func TestUser_Update(t *testing.T) {
	id := NewUserID("user-id-3")
	originalUsername := "originaluser"
	originalEmail := "original@example.com"
	originalPasswordHash := []byte("originalhashedpassword")
	originalCreatedAt := time.Now().Add(-72 * time.Hour)
	originalUpdatedAt := time.Now().Add(-48 * time.Hour)

	user := NewUser(id, originalUsername, originalEmail, originalPasswordHash, originalCreatedAt, originalUpdatedAt)

	newUsername := "updateduser"
	newEmail := "updated@example.com"
	newPasswordHash := []byte("newhashedpassword")
	newUpdatedAt := time.Now()

	updatedUser := user.Update(newUsername, newEmail, newPasswordHash, newUpdatedAt)

	assert.NotNil(t, updatedUser, "Update は新しい User インスタンスを返すべき")
	assert.Equal(t, id, updatedUser.ID(), "Update は元の ID を保持すべき")
	assert.Equal(t, newUsername, updatedUser.Username(), "Update は新しい username を設定すべき")
	assert.Equal(t, newEmail, updatedUser.Email(), "Update は新しい email を設定すべき")
	assert.Equal(t, newPasswordHash, updatedUser.PasswordHash(), "Update は新しい passwordHash を設定すべき")
	assert.Equal(t, originalCreatedAt, updatedUser.CreatedAt(), "Update は元の createdAt を保持すべき")
	assert.Equal(t, newUpdatedAt, updatedUser.UpdatedAt(), "Update は新しい updatedAt を設定すべき")

	// 元の user が変更されていないことを確認
	assert.Equal(t, originalUsername, user.Username(), "元の User の username は変更されてはいけない")
	assert.Equal(t, originalEmail, user.Email(), "元の User の email は変更されてはいけない")
	assert.Equal(t, originalPasswordHash, user.PasswordHash(), "元の User の passwordHash は変更されてはいけない")
	assert.Equal(t, originalUpdatedAt, user.UpdatedAt(), "元の User の updatedAt は変更されてはいけない")
}

func TestUser_Equals(t *testing.T) {
	id1 := NewUserID("user-id-4")
	id2 := NewUserID("user-id-5")
	now := time.Now()

	user1 := NewUser(id1, "userA", "a@example.com", []byte("hashA"), now, now)
	user2 := NewUser(id1, "userA", "a@example.com", []byte("hashA"), now, now) // user1 と同じ ID
	user3 := NewUser(id2, "userB", "b@example.com", []byte("hashB"), now, now) // user1 と異なる ID

	assert.True(t, user1.Equals(user2), "同じ ID を持つ 2 つの User は等しいと判定されるべき")
	assert.False(t, user1.Equals(user3), "異なる ID を持つ 2 つの User は等しくないと判定されるべき")
	assert.False(t, user1.Equals(nil), "User は nil と等しいと判定されるべきではない")
}
