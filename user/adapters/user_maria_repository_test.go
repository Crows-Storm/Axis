package adapters

import (
	"context"
	"strconv"
	"testing"

	"github.com/Crows-Storm/Axis/common/server/store"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB 创建测试数据库（SQLite 内存数据库）
func setupTestDB(t *testing.T) *store.Store {
	st, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: ":memory:",
	})
	require.NoError(t, err)
	return st
}

// TestUserMariaRepository_Create 测试创建用户
func TestUserMariaRepository_Create(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			LoginId:  "testuser",
			Password: "hashed_password",
			Email:    "test@example.com",
		}

		createdUser, err := repo.Create(ctx, user)

		assert.NoError(t, err)
		assert.NotZero(t, createdUser.Id)
		assert.Equal(t, "testuser", createdUser.LoginId)
		assert.Equal(t, "test@example.com", createdUser.Email)
		assert.Equal(t, int8(1), createdUser.Status)
		assert.Equal(t, int8(0), createdUser.Deleted)
		assert.False(t, createdUser.CreateTime.IsZero())
		assert.False(t, createdUser.UpdateTime.IsZero())
	})

	t.Run("duplicate_loginId", func(t *testing.T) {
		user1 := &domain.User{
			LoginId:  "duplicate",
			Password: "password",
			Email:    "user1@example.com",
		}
		_, err := repo.Create(ctx, user1)
		require.NoError(t, err)

		// 尝试创建相同 LoginId 的用户
		user2 := &domain.User{
			LoginId:  "duplicate",
			Password: "password",
			Email:    "user2@example.com",
		}
		_, err = repo.Create(ctx, user2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("duplicate_email", func(t *testing.T) {
		user1 := &domain.User{
			LoginId:  "user1",
			Password: "password",
			Email:    "duplicate@example.com",
		}
		_, err := repo.Create(ctx, user1)
		require.NoError(t, err)

		// 尝试创建相同 Email 的用户
		user2 := &domain.User{
			LoginId:  "user2",
			Password: "password",
			Email:    "duplicate@example.com",
		}
		_, err = repo.Create(ctx, user2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

// TestUserMariaRepository_GetInfo 测试查询用户
func TestUserMariaRepository_GetInfo(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// 创建用户
		user := &domain.User{
			LoginId:  "getuser",
			Password: "password",
			Email:    "get@example.com",
		}
		created, err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 查询用户
		found, err := repo.GetInfo(created.Id)

		assert.NoError(t, err)
		assert.Equal(t, created.Id, found.Id)
		assert.Equal(t, "getuser", found.LoginId)
		assert.Equal(t, "get@example.com", found.Email)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.GetInfo(99999)

		assert.Error(t, err)
		_, ok := err.(domain.NotFoundError)
		assert.True(t, ok)
	})

	t.Run("soft_deleted_user_not_found", func(t *testing.T) {
		// 创建并软删除用户
		user := &domain.User{
			LoginId:  "deleteduser",
			Password: "password",
			Email:    "deleted@example.com",
		}
		created, err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.SoftDelete(ctx, created.Id)
		require.NoError(t, err)

		// 查询已删除的用户
		_, err = repo.GetInfo(created.Id)

		assert.Error(t, err)
		_, ok := err.(domain.NotFoundError)
		assert.True(t, ok)
	})
}

// TestUserMariaRepository_GetByLoginId 测试根据 LoginId 查询
func TestUserMariaRepository_GetByLoginId(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			LoginId:  "loginuser",
			Password: "password",
			Email:    "login@example.com",
		}
		_, err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.GetByLoginId(ctx, "loginuser")

		assert.NoError(t, err)
		assert.Equal(t, "loginuser", found.LoginId)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.GetByLoginId(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestUserMariaRepository_Update 测试更新用户
func TestUserMariaRepository_Update(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// 创建用户
		user := &domain.User{
			LoginId:  "updateuser",
			Password: "password",
			Email:    "update@example.com",
		}
		created, err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 更新用户
		err = repo.Update(ctx, &domain.User{Id: created.Id}, func(ctx context.Context, u *domain.User) (*domain.User, error) {
			u.Email = "newemail@example.com"
			return u, nil
		})

		assert.NoError(t, err)

		// 验证更新
		updated, err := repo.GetInfo(created.Id)
		require.NoError(t, err)
		assert.Equal(t, "newemail@example.com", updated.Email)
	})

	t.Run("user_not_found", func(t *testing.T) {
		err := repo.Update(ctx, &domain.User{Id: 99999}, func(ctx context.Context, u *domain.User) (*domain.User, error) {
			return u, nil
		})

		assert.Error(t, err)
		_, ok := err.(domain.NotFoundError)
		assert.True(t, ok)
	})

	t.Run("update_function_error", func(t *testing.T) {
		user := &domain.User{
			LoginId:  "erroruser",
			Password: "password",
			Email:    "error@example.com",
		}
		created, err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.Update(ctx, &domain.User{Id: created.Id}, func(ctx context.Context, u *domain.User) (*domain.User, error) {
			return nil, assert.AnError
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update function failed")
	})
}

// TestUserMariaRepository_UpdateStatus 测试更新状态
func TestUserMariaRepository_UpdateStatus(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			LoginId:  "statususer",
			Password: "password",
			Email:    "status@example.com",
		}
		created, err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 更新状态为 0（停用）
		err = repo.UpdateStatus(ctx, created.Id, 0)
		assert.NoError(t, err)

		// 验证状态
		updated, err := repo.GetInfo(created.Id)
		require.NoError(t, err)
		assert.Equal(t, int8(0), updated.Status)
	})

	t.Run("user_not_found", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, 99999, 1)

		assert.Error(t, err)
		_, ok := err.(domain.NotFoundError)
		assert.True(t, ok)
	})
}

// TestUserMariaRepository_SoftDelete 测试软删除
func TestUserMariaRepository_SoftDelete(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{
			LoginId:  "deleteuser",
			Password: "password",
			Email:    "delete@example.com",
		}
		created, err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 软删除
		err = repo.SoftDelete(ctx, created.Id)
		assert.NoError(t, err)

		// 验证无法查询到
		_, err = repo.GetInfo(created.Id)
		assert.Error(t, err)
	})

	t.Run("user_not_found", func(t *testing.T) {
		err := repo.SoftDelete(ctx, 99999)

		assert.Error(t, err)
		_, ok := err.(domain.NotFoundError)
		assert.True(t, ok)
	})
}

// TestUserMariaRepository_List 测试分页查询
func TestUserMariaRepository_List(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	// 创建测试数据
	for i := 1; i <= 25; i++ {
		user := &domain.User{
			LoginId:  "listuser" + string(strconv.Itoa(i)),
			Password: "password",
			Email:    "list" + string(strconv.Itoa(i)) + "@example.com",
		}
		if i%2 == 0 {
			user.Status = 0 // 偶数用户停用
		}
		_, err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("first_page", func(t *testing.T) {
		users, total, err := repo.List(ctx, 1, 10, nil)

		assert.NoError(t, err)
		assert.Equal(t, int64(25), total)
		assert.Len(t, users, 10)
	})

	t.Run("second_page", func(t *testing.T) {
		users, total, err := repo.List(ctx, 2, 10, nil)

		assert.NoError(t, err)
		assert.Equal(t, int64(25), total)
		assert.Len(t, users, 10)
	})

	t.Run("last_page", func(t *testing.T) {
		users, total, err := repo.List(ctx, 3, 10, nil)

		assert.NoError(t, err)
		assert.Equal(t, int64(25), total)
		assert.Len(t, users, 5)
	})

	t.Run("filter_by_status", func(t *testing.T) {
		filters := map[string]interface{}{
			"status": 1, // 只查询活跃用户
		}
		users, total, err := repo.List(ctx, 1, 100, filters)

		assert.NoError(t, err)
		assert.Equal(t, int64(13), total) // 25个用户中，13个是活跃的（奇数）
		assert.Len(t, users, 13)
	})

	t.Run("filter_by_keyword", func(t *testing.T) {
		filters := map[string]interface{}{
			"keyword": "listuser1",
		}
		_, total, err := repo.List(ctx, 1, 100, filters)

		assert.NoError(t, err)
		assert.Greater(t, total, int64(0))
	})
}

// TestUserMariaRepository_CreateBatch 测试批量创建
func TestUserMariaRepository_CreateBatch(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		users := []*domain.User{
			{LoginId: "batch1", Password: "pass", Email: "batch1@example.com"},
			{LoginId: "batch2", Password: "pass", Email: "batch2@example.com"},
			{LoginId: "batch3", Password: "pass", Email: "batch3@example.com"},
		}

		err := repo.CreateBatch(ctx, users)
		assert.NoError(t, err)

		// 验证创建成功
		found, err := repo.GetByLoginId(ctx, "batch1")
		assert.NoError(t, err)
		assert.Equal(t, "batch1", found.LoginId)
	})

	t.Run("empty_slice", func(t *testing.T) {
		err := repo.CreateBatch(ctx, []*domain.User{})
		assert.NoError(t, err)
	})
}

// TestUserMariaRepository_GetStats 测试统计信息
func TestUserMariaRepository_GetStats(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	// 创建测试数据
	for i := 1; i <= 10; i++ {
		user := &domain.User{
			LoginId:  "statsuser" + string(strconv.Itoa(i)),
			Password: "password",
			Email:    "stats" + string(strconv.Itoa(i)) + "@example.com",
		}
		created, err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 删除部分用户
		if i <= 3 {
			err = repo.SoftDelete(ctx, created.Id)
			require.NoError(t, err)
		}
	}

	stats, err := repo.GetStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), stats["total_users"])
	assert.Equal(t, int64(7), stats["active_users"])
	assert.Equal(t, int64(3), stats["deleted_users"])
}

// TestUserMariaRepository_ConcurrentCreate 测试并发创建
func TestUserMariaRepository_ConcurrentCreate(t *testing.T) {
	st := setupTestDB(t)
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	done := make(chan bool, 10)
	errors := make(chan error, 10)

	// 并发创建 10 个用户
	for i := 0; i < 10; i++ {
		go func(index int) {
			user := &domain.User{
				LoginId:  "concurrent" + string(strconv.Itoa(index)),
				Password: "password",
				Email:    "concurrent" + string(strconv.Itoa(index)) + "@example.com",
			}
			_, err := repo.Create(ctx, user)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 不应该有错误
	select {
	case err := <-errors:
		t.Fatalf("Unexpected error: %v", err)
	default:
	}

	// 验证所有用户都创建成功
	users, total, err := repo.List(ctx, 1, 100, map[string]interface{}{
		"keyword": "concurrent",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, users, 10)
}

// BenchmarkUserMariaRepository_Create 基准测试：创建用户
func BenchmarkUserMariaRepository_Create(b *testing.B) {
	st, _ := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: ":memory:",
	})
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &domain.User{
			LoginId:  "benchuser" + string(strconv.Itoa(i)),
			Password: "password",
			Email:    "bench" + string(strconv.Itoa(i)) + "@example.com",
		}
		_, _ = repo.Create(ctx, user)
	}
}

// BenchmarkUserMariaRepository_GetInfo 基准测试：查询用户
func BenchmarkUserMariaRepository_GetInfo(b *testing.B) {
	st, _ := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: ":memory:",
	})
	defer st.Close()

	repo := NewUserMariaRepository(st)
	ctx := context.Background()

	// 创建测试用户
	user := &domain.User{
		LoginId:  "benchgetuser",
		Password: "password",
		Email:    "benchget@example.com",
	}
	created, _ := repo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetInfo(created.Id)
	}
}
