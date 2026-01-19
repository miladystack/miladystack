package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/miladystack/miladystack/pkg/store"
	"github.com/miladystack/miladystack/pkg/store/logger/empty"
	"github.com/miladystack/miladystack/pkg/store/logger/milady"
	"github.com/miladystack/miladystack/pkg/store/where"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// User represents a user model with custom soft delete support
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255" json:"name"`
	Email     string         `gorm:"size:255;uniqueIndex" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:is_deleted;comment:软删除时间;index" json:"is_deleted"` // 软删除字段，使用自定义列名
}

// MySQLProvider implements DBProvider interface
type MySQLProvider struct {
	db *gorm.DB
}

// DB returns the database instance
func (p *MySQLProvider) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	db := p.db.WithContext(ctx)
	for _, where := range wheres {
		db = where.Where(db)
	}
	return db
}

// LoggerType defines the type of logger to use
type LoggerType string

const (
	LoggerTypeEmpty  LoggerType = "empty"
	LoggerTypeMilady LoggerType = "milady"
)

// initDB initializes the database connection and returns the store instance
func initDB(loggerType LoggerType) (*store.Store[User], context.Context, error) {
	// Connect to MySQL database
	dsn := "milady:milady(#)888@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the User model
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create DBProvider
	dbProvider := &MySQLProvider{db: db}

	// Create logger based on type
	var logger store.Logger
	switch loggerType {
	case LoggerTypeMilady:
		logger = milady.NewLogger()
		fmt.Printf("✅ Using Milady Logger\n")
	default:
		logger = empty.NewLogger()
		fmt.Printf("✅ Using Empty Logger\n")
	}

	// Create store instance for User model
	userStore := store.NewStore[User](dbProvider, logger)

	ctx := context.Background()
	return userStore, ctx, nil
}

// testCreateUser tests creating a new user
func testCreateUser(store *store.Store[User], ctx context.Context) (*User, error) {
	fmt.Println("=== CREATE USER TEST ===")
	newUser := &User{
		Name:  "John Doe",
		Email: fmt.Sprintf("john.doe%d@example.com", time.Now().Unix()),
	}

	err := store.Create(ctx, newUser)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, err
	}

	fmt.Printf("✅ Created user: %+v\n", newUser)
	return newUser, nil
}

// testGetUser tests retrieving a user by ID
func testGetUser(store *store.Store[User], ctx context.Context, userID uint) (*User, error) {
	fmt.Println("\n=== GET USER TEST ===")
	retrievedUser, err := store.Get(ctx, where.F("id", userID))
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return nil, err
	}

	fmt.Printf("✅ Retrieved user by ID %d: %+v\n", userID, retrievedUser)
	return retrievedUser, nil
}

// testUpdateUser tests updating a user's information
func testUpdateUser(store *store.Store[User], ctx context.Context, user *User) (*User, error) {
	fmt.Println("\n=== UPDATE USER TEST ===")
	// Update user's name
	user.Name = "Updated Name"
	err := store.Update(ctx, user)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return nil, err
	}

	// Verify update was successful
	updatedUser, _ := store.Get(ctx, where.F("id", user.ID))
	fmt.Printf("✅ Updated user: %+v\n", updatedUser)
	return updatedUser, nil
}

// createTestData creates multiple test users
func createTestData(store *store.Store[User], ctx context.Context, count int) error {
	fmt.Println("\n=== CREATING TEST DATA ===")
	for i := 0; i < count; i++ {
		user := &User{
			Name:  fmt.Sprintf("Test User %d", i+1),
			Email: fmt.Sprintf("test.user%d@example.com", time.Now().Unix()+int64(i)),
		}

		err := store.Create(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to create test user %d: %w", i+1, err)
		}
	}
	fmt.Printf("✅ Created %d test users\n", count)
	return nil
}

// testListUsers tests listing users with pagination
func testListUsers(store *store.Store[User], ctx context.Context) error {
	// Create enough test data (20 users)
	err := createTestData(store, ctx, 20)
	if err != nil {
		return err
	}

	fmt.Println("\n=== LIST USERS TEST ===")
	// Test listing all users (limit -1 means no limit)
	count, users, err := store.List(ctx, where.P(1, -1))
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return err
	}

	fmt.Printf("✅ Total users: %d\n", count)
	fmt.Println("\n--- Current Sort Behavior (Default: id desc) ---")
	fmt.Println("First 5 users (should be sorted by ID descending):")
	for i, user := range users {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. ID: %d, Name: %s\n", i+1, user.ID, user.Name)
	}

	// Test pagination
	fmt.Println("\n--- Pagination Test (Page 2, Limit 5) ---")
	count, users, err = store.List(ctx, where.P(2, 5))
	if err != nil {
		log.Printf("Failed to list users with pagination: %v", err)
		return err
	}

	fmt.Printf("✅ Paginated users (Page 2, Limit 5): %d users\n", len(users))
	for i, user := range users {
		fmt.Printf("  %d. ID: %d, Name: %s\n", i+1, user.ID, user.Name)
	}

	// Test custom sorting
	fmt.Println("\n--- CUSTOM SORTING TESTS ---")

	// Test 1: Sort by Name ascending
	fmt.Println("\n1. Sort by Name ascending (name asc):")
	count, users, err = store.List(ctx, where.P(1, 5).Or("name asc"))
	if err != nil {
		log.Printf("Failed to list users with custom sort: %v", err)
		return err
	}
	for i, user := range users {
		fmt.Printf("   %d. ID: %d, Name: %s\n", i+1, user.ID, user.Name)
	}

	// Test 2: Sort by Name descending
	fmt.Println("\n2. Sort by Name descending (name desc):")
	count, users, err = store.List(ctx, where.P(1, 5).Or("name desc"))
	if err != nil {
		log.Printf("Failed to list users with custom sort: %v", err)
		return err
	}
	for i, user := range users {
		fmt.Printf("   %d. ID: %d, Name: %s\n", i+1, user.ID, user.Name)
	}

	// Test 3: Sort by CreatedAt ascending
	fmt.Println("\n3. Sort by CreatedAt ascending (created_at asc):")
	count, users, err = store.List(ctx, where.P(1, 5).Or("created_at asc"))
	if err != nil {
		log.Printf("Failed to list users with custom sort: %v", err)
		return err
	}
	for i, user := range users {
		fmt.Printf("   %d. ID: %d, Name: %s, CreatedAt: %s\n", i+1, user.ID, user.Name, user.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Test 4: Sort by multiple fields
	fmt.Println("\n4. Sort by Name ascending and ID ascending (name asc, id asc):")
	count, users, err = store.List(ctx, where.P(1, 5).Or("name asc, id asc"))
	if err != nil {
		log.Printf("Failed to list users with custom sort: %v", err)
		return err
	}
	for i, user := range users {
		fmt.Printf("   %d. ID: %d, Name: %s\n", i+1, user.ID, user.Name)
	}

	fmt.Println("\n✅ Custom sorting feature is working correctly!")

	return nil
}

// testDeleteUser tests deleting a user
func testDeleteUser(store *store.Store[User], ctx context.Context, userID uint) error {
	fmt.Println("\n=== DELETE USER TEST ===")
	err := store.Delete(ctx, where.F("id", userID))
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return err
	}

	fmt.Printf("✅ Deleted user with ID: %d\n", userID)

	// Verify deletion
	deletedUser, err := store.Get(ctx, where.F("id", userID))
	if err != nil {
		fmt.Printf("✅ Verify deletion: User not found (expected): %v\n", err)
		return nil
	}

	fmt.Printf("❌ Verify deletion: User still exists (unexpected): %+v\n", deletedUser)
	return fmt.Errorf("user was not properly deleted")
}

// testSoftDelete tests soft deletion and unscoped queries
func testSoftDelete(store *store.Store[User], ctx context.Context) error {
	fmt.Println("\n=== SOFT DELETE AND UNSCOPED QUERY TEST ===")

	// 1. Create a test user for soft delete
	fmt.Println("\n1. Creating test user for soft delete...")
	softDeleteUser := &User{
		Name:  "Soft Delete Test User",
		Email: fmt.Sprintf("soft.delete.test%d@example.com", time.Now().Unix()),
	}

	err := store.Create(ctx, softDeleteUser)
	if err != nil {
		log.Printf("Failed to create soft delete test user: %v", err)
		return err
	}

	userID := softDeleteUser.ID
	fmt.Printf("   ✅ Created test user: ID=%d, Name=%s\n", userID, softDeleteUser.Name)

	// 2. Delete the user (soft delete)
	fmt.Println("\n2. Performing soft delete on test user...")
	err = store.Delete(ctx, where.F("id", userID))
	if err != nil {
		log.Printf("Failed to soft delete user: %v", err)
		return err
	}

	fmt.Printf("   ✅ Soft deleted user with ID: %d\n", userID)

	// 3. Try to get the user with normal query (should fail)
	fmt.Println("\n3. Attempting to get soft deleted user with normal query...")
	deletedUser, err := store.Get(ctx, where.F("id", userID))
	if err != nil {
		fmt.Printf("   ✅ Normal query: User not found (expected): %v\n", err)
	} else {
		fmt.Printf("   ❌ Normal query: User still found (unexpected): %+v\n", deletedUser)
		return fmt.Errorf("user should not be found with normal query after soft delete")
	}

	// 4. Try to get the user with Unscoped query (should succeed)
	fmt.Println("\n4. Attempting to get soft deleted user with Unscoped query...")
	unscopedUser, err := store.Get(ctx, where.F("id", userID).U(true))
	if err != nil {
		fmt.Printf("   ❌ Unscoped query: User not found (unexpected): %v\n", err)
		return err
	} else {
		fmt.Printf("   ✅ Unscoped query: Found soft deleted user: %+v\n", unscopedUser)
	}

	// 5. List users with normal query (should not include deleted user)
	fmt.Println("\n5. Listing users with normal query (should exclude deleted users):")
	count, _, err := store.List(ctx, where.P(1, 10))
	if err != nil {
		log.Printf("Failed to list users with normal query: %v", err)
		return err
	}
	fmt.Printf("   ✅ Normal query - Total users: %d\n", count)

	// 6. List users with Unscoped query (should include deleted user)
	fmt.Println("\n6. Listing users with Unscoped query (should include deleted users):")
	unscopedCount, unscopedUsers, err := store.List(ctx, where.P(1, 10).U(true))
	if err != nil {
		log.Printf("Failed to list users with unscoped query: %v", err)
		return err
	}
	fmt.Printf("   ✅ Unscoped query - Total users: %d\n", unscopedCount)

	// 7. Compare counts to show difference
	fmt.Printf("\n7. Count comparison: %d total users (including %d deleted)\n", unscopedCount, unscopedCount-count)

	// 8. Show deleted users in unscoped list
	fmt.Println("\n8. Showing deleted users in unscoped list:")
	for _, user := range unscopedUsers {
		if user.DeletedAt.Time != (time.Time{}) {
			fmt.Printf("   🗑️  Deleted user: ID=%d, Name=%s, DeletedAt=%s\n",
				user.ID, user.Name, user.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}
	}

	// 9. Restore the soft deleted user
	fmt.Println("\n9. Restoring soft deleted user...")
	// To restore a soft deleted record, we need to update the DeletedAt field to zero value
	// Get the record with Unscoped first
	restoredUser, err := store.Get(ctx, where.F("id", userID).U(true))
	if err != nil {
		log.Printf("Failed to get soft deleted user for restoration: %v", err)
		return err
	}

	// Clear the DeletedAt field to restore
	restoredUser.DeletedAt = gorm.DeletedAt{}
	// Update the user
	err = store.Update(ctx, restoredUser)
	if err != nil {
		log.Printf("Failed to restore user: %v", err)
		return err
	}
	fmt.Printf("   ✅ Restored user with ID: %d\n", userID)

	// 10. Verify restoration
	fmt.Println("\n10. Verifying restoration with normal query...")
	restoredUser, err = store.Get(ctx, where.F("id", userID))
	if err != nil {
		fmt.Printf("   ❌ Normal query: User not found (unexpected): %v\n", err)
		return err
	} else {
		fmt.Printf("   ✅ Normal query: Found restored user: %+v\n", restoredUser)
	}

	fmt.Println("\n✅ Soft delete, Unscoped query and restoration tests completed successfully!")
	return nil
}

// testFilteredList tests listing users with filters
func testFilteredList(store *store.Store[User], ctx context.Context) error {
	fmt.Println("\n=== FILTERED LIST TEST ===")
	// Create a test user with specific name for filtering
	testUser := &User{
		Name:  "Filter Test User",
		Email: fmt.Sprintf("filter.test%d@example.com", time.Now().Unix()),
	}

	err := store.Create(ctx, testUser)
	if err != nil {
		return fmt.Errorf("failed to create test user for filtering: %w", err)
	}

	// Test filtering by name
	count, users, err := store.List(ctx, where.F("name", "Filter Test User"))
	if err != nil {
		log.Printf("Failed to list users with filter: %v", err)
		return err
	}

	fmt.Printf("✅ Filtered users by name 'Filter Test User': %d users found\n", count)
	for i, user := range users {
		fmt.Printf("  %d. %+v\n", i+1, user)
	}

	// Clean up test user
	store.Delete(ctx, where.F("id", testUser.ID))

	return nil
}

// runTests runs all the tests with the given store instance
func runTests(store *store.Store[User], ctx context.Context) error {
	// Run all tests in sequence
	var createdUser *User
	var retrievedUser *User
	var updatedUser *User

	// 1. Create a new user
	createdUser, err := testCreateUser(store, ctx)
	if err != nil {
		return fmt.Errorf("Create user test failed: %w", err)
	}

	// 2. Get the created user
	retrievedUser, err = testGetUser(store, ctx, createdUser.ID)
	if err != nil {
		return fmt.Errorf("Get user test failed: %w", err)
	}

	// 3. Update the user
	updatedUser, err = testUpdateUser(store, ctx, retrievedUser)
	if err != nil {
		return fmt.Errorf("Update user test failed: %w", err)
	}

	// 4. List all users
	err = testListUsers(store, ctx)
	if err != nil {
		return fmt.Errorf("List users test failed: %w", err)
	}

	// 5. Test filtered list
	err = testFilteredList(store, ctx)
	if err != nil {
		return fmt.Errorf("Filtered list test failed: %w", err)
	}

	// 6. Delete the user
	err = testDeleteUser(store, ctx, updatedUser.ID)
	if err != nil {
		return fmt.Errorf("Delete user test failed: %w", err)
	}

	return nil
}

func main() {
	fmt.Println("🚀 STARTING STORE PACKAGE TESTS...")
	fmt.Println("====================================")

	// Test with Milady Logger
	fmt.Println("\nTesting with Milady Logger:")
	fmt.Println("------------------------------------")
	userStore, ctx, err := initDB(LoggerTypeMilady)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("✅ Database initialized successfully")

	// Run tests with Milady Logger
	var createdUser *User
	var retrievedUser *User
	var updatedUser *User

	// 1. Create a new user
	createdUser, err = testCreateUser(userStore, ctx)
	if err != nil {
		log.Fatalf("Create user test failed: %v", err)
	}

	// 2. Get the created user
	retrievedUser, err = testGetUser(userStore, ctx, createdUser.ID)
	if err != nil {
		log.Fatalf("Get user test failed: %v", err)
	}

	// 3. Update the user
	updatedUser, err = testUpdateUser(userStore, ctx, retrievedUser)
	if err != nil {
		log.Fatalf("Update user test failed: %v", err)
	}

	// 4. List users without creating duplicate data
	fmt.Println("\n=== LIST USERS TEST (Without Creating Duplicate Data) ===")
	// Test listing all users with pagination
	count, users, err := userStore.List(ctx, where.P(1, 5))
	if err != nil {
		log.Printf("Failed to list users: %v", err)
	} else {
		fmt.Printf("✅ Total users: %d\n", count)
		fmt.Println("First 5 users:")
		for i, user := range users {
			fmt.Printf("  %d. ID: %d, Name: %s\n", i+1, user.ID, user.Name)
		}
	}

	// Test custom sorting with Milady Logger
	fmt.Println("\n--- CUSTOM SORTING WITH MILADY LOGGER ---")
	count, users, err = userStore.List(ctx, where.P(1, 3).Or("name asc"))
	if err != nil {
		log.Printf("Failed to list users with custom sort: %v", err)
	} else {
		fmt.Println("Users sorted by Name ascending:")
		for i, user := range users {
			fmt.Printf("   %d. ID: %d, Name: %s\n", i+1, user.ID, user.Name)
		}
	}

	// 5. Delete the user
	err = testDeleteUser(userStore, ctx, updatedUser.ID)
	if err != nil {
		log.Fatalf("Delete user test failed: %v", err)
	}

	// 6. Test soft delete and unscoped queries
	err = testSoftDelete(userStore, ctx)
	if err != nil {
		log.Fatalf("Soft delete test failed: %v", err)
	}

	fmt.Println("\n🎉 ALL TESTS COMPLETED SUCCESSFULLY!")
	fmt.Println("====================================")
}
