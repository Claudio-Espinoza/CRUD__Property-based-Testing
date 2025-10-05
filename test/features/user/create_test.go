package user_test

import (
	"sync"
	"testing"

	"pgregory.net/rapid"

	"property-based/internal/domain"
	"property-based/internal/repository"
	"property-based/internal/service"
	"property-based/test/generators"
	"property-based/test/helpers"
)

func TestProperty_UserCreate_ValidData_SucceedsAndIsRetrievable(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")

		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)

		helpers.AssertNoError(t, err, "CreateUser with valid data")

		if created.ID == "" {
			t.Fatal("Created user must have non-empty ID")
		}

		if created.Name != userData.Name {
			t.Fatalf("Name mismatch: expected %q, got %q", userData.Name, created.Name)
		}
		if created.Email != userData.Email {
			t.Fatalf("Email mismatch: expected %q, got %q", userData.Email, created.Email)
		}
		if created.Age != userData.Age {
			t.Fatalf("Age mismatch: expected %d, got %d", userData.Age, created.Age)
		}

		if created.UpdatedAt.Before(created.CreatedAt) {
			t.Fatal("UpdatedAt must be >= CreatedAt")
		}

		retrieved, err := svc.GetUser(created.ID)
		helpers.AssertNoError(t, err, "GetUser after create")
		helpers.AssertUserEquals(t, created, retrieved, "Retrieved user")

		if err := created.Validate(); err != nil {
			t.Fatalf("Created user fails domain validation: %v", err)
		}
	})
}

func TestProperty_UserCreate_InvalidData_FailsWithoutPersisting(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		invalidData := generators.InvalidUserStruct().Draw(t, "invalid_data")

		created, err := svc.CreateUser(invalidData.Name, invalidData.Email, invalidData.Age)

		helpers.AssertError(t, err, "CreateUser with invalid data")

		if created != nil {
			t.Fatalf("CreateUser should return nil user on error, got: %+v", created)
		}

		if count := svc.CountUsers(); count != 0 {
			t.Fatalf("Invalid user should not be persisted, found %d users", count)
		}

		allUsers, _ := svc.GetAllUsers()
		if len(allUsers) != 0 {
			t.Fatalf("System should have 0 users, found %d", len(allUsers))
		}
	})
}

func TestProperty_UserCreate_DuplicateEmail_Fails(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		firstUser := generators.ValidUserStruct().Draw(t, "first_user")
		created1, err := svc.CreateUser(firstUser.Name, firstUser.Email, firstUser.Age)
		helpers.AssertNoError(t, err, "Create first user")

		secondUser := generators.ValidUserStruct().Draw(t, "second_user")

		created2, err := svc.CreateUser(secondUser.Name, created1.Email, secondUser.Age)

		helpers.AssertErrorIs(t, err, domain.ErrAlreadyExists, "Duplicate email")

		if created2 != nil {
			t.Fatal("Second user with duplicate email should not be created")
		}

		if count := svc.CountUsers(); count != 1 {
			t.Fatalf("Expected 1 user, found %d", count)
		}
	})
}

// TestProperty_UserCreate_ConcurrentCreates_AllSucceedWithUniqueIDs
//
// IMPORTANTE: Los datos se generan ANTES de las goroutines
func TestProperty_UserCreate_ConcurrentCreates_AllSucceedWithUniqueIDs(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userCount := rapid.IntRange(2, 10).Draw(rt, "user_count")

		// ✅ GENERAR TODOS LOS DATOS **ANTES** DE LAS GOROUTINES
		usersData := make([]generators.ValidUserData, userCount)
		for i := 0; i < userCount; i++ {
			usersData[i] = generators.ValidUserStruct().Draw(rt, "user")
		}

		type result struct {
			user *domain.User
			err  error
		}
		results := make(chan result, userCount)
		var wg sync.WaitGroup

		// Crear usuarios concurrentemente usando los datos ya generados
		for i := 0; i < userCount; i++ {
			wg.Add(1)
			go func(userData generators.ValidUserData) {
				defer wg.Done()
				user, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
				results <- result{user: user, err: err}
			}(usersData[i]) // Pasar los datos como parámetro
		}

		// Esperar a que todas las goroutines terminen
		go func() {
			wg.Wait()
			close(results)
		}()

		// Verificar resultados
		createdIDs := make(map[string]bool)
		successCount := 0

		for res := range results {
			helpers.AssertNoError(rt, res.err, "Concurrent create")

			if createdIDs[res.user.ID] {
				rt.Fatalf("Duplicate ID generated: %s", res.user.ID)
			}
			createdIDs[res.user.ID] = true
			successCount++
		}

		if successCount != userCount {
			rt.Fatalf("Expected %d successful creates, got %d", userCount, successCount)
		}

		if count := svc.CountUsers(); count != userCount {
			rt.Fatalf("Expected %d users in repository, found %d", userCount, count)
		}
	})
}
