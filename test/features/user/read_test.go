package user_test

import (
	"testing"

	"pgregory.net/rapid"

	"property-based/internal/domain"
	"property-based/internal/repository"
	"property-based/internal/service"
	"property-based/test/generators"
	"property-based/test/helpers"
)

// TestProperty_UserRead_ExistingUser_ReturnsCorrectData
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ user creado, GetUser(user.ID) = user
//
// INVARIANTES:
//  1. Datos devueltos coinciden exactamente con los creados
//  2. La operación es idempotente: múltiples lecturas devuelven lo mismo
func TestProperty_UserRead_ExistingUser_ReturnsCorrectData(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")

		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
		helpers.AssertNoError(t, err, "Create user")

		// Primera lectura
		retrieved1, err := svc.GetUser(created.ID)
		helpers.AssertNoError(t, err, "First GetUser")
		helpers.AssertUserEquals(t, created, retrieved1, "First read")

		// Segunda lectura (idempotencia)
		retrieved2, err := svc.GetUser(created.ID)
		helpers.AssertNoError(t, err, "Second GetUser")
		helpers.AssertUserEquals(t, created, retrieved2, "Second read")
		helpers.AssertUserEquals(t, retrieved1, retrieved2, "Reads are idempotent")
	})
}

// TestProperty_UserRead_NonExistentUser_ReturnsNotFound
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ id ∉ Sistema, GetUser(id) = (nil, ErrNotFound)
func TestProperty_UserRead_NonExistentUser_ReturnsNotFound(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		// Generar ID aleatorio que no existe
		nonExistentID := rapid.StringMatching(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`).
			Draw(t, "non_existent_id")

		user, err := svc.GetUser(nonExistentID)

		helpers.AssertErrorIs(t, err, domain.ErrNotFound, "GetUser with non-existent ID")

		if user != nil {
			t.Fatalf("Expected nil user, got: %+v", user)
		}
	})
}

// TestProperty_UserRead_ByEmail_FindsCorrectUser
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ user creado, GetUserByEmail(user.Email) = user
func TestProperty_UserRead_ByEmail_FindsCorrectUser(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")

		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
		helpers.AssertNoError(t, err, "Create user")

		retrieved, err := svc.GetUserByEmail(created.Email)
		helpers.AssertNoError(t, err, "GetUserByEmail")
		helpers.AssertUserEquals(t, created, retrieved, "Retrieved by email")
	})
}

// TestProperty_UserRead_GetAll_ReturnsAllCreatedUsers
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ conjunto de users U creados,
//	  |GetAllUsers()| = |U| ∧ ∀u ∈ U, u ∈ GetAllUsers()
func TestProperty_UserRead_GetAll_ReturnsAllCreatedUsers(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userCount := rapid.IntRange(1, 10).Draw(t, "user_count")
		createdUsers := make(map[string]*domain.User)

		// Crear múltiples usuarios
		for i := 0; i < userCount; i++ {
			userData := generators.ValidUserStruct().Draw(t, "user")
			created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
			helpers.AssertNoError(t, err, "Create user")
			createdUsers[created.ID] = created
		}

		// Obtener todos
		allUsers, err := svc.GetAllUsers()
		helpers.AssertNoError(t, err, "GetAllUsers")

		// Verificar cantidad
		if len(allUsers) != userCount {
			t.Fatalf("Expected %d users, got %d", userCount, len(allUsers))
		}

		// Verificar que todos los creados están en la lista
		for _, user := range allUsers {
			originalUser, exists := createdUsers[user.ID]
			if !exists {
				t.Fatalf("User %s not in created users", user.ID)
			}
			helpers.AssertUserEquals(t, originalUser, user, "User in GetAll")
		}
	})
}

// TestProperty_UserRead_ConcurrentReads_AreConsistent
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ user, ∀ lecturas concurrentes,
//	  Todas las lecturas devuelven datos idénticos
func TestProperty_UserRead_ConcurrentReads_AreConsistent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")
		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
		helpers.AssertNoError(t, err, "Create user")

		readCount := rapid.IntRange(5, 20).Draw(t, "read_count")
		results := make(chan *domain.User, readCount)

		// Múltiples lecturas concurrentes
		for i := 0; i < readCount; i++ {
			go func() {
				user, err := svc.GetUser(created.ID)
				if err == nil {
					results <- user
				}
			}()
		}

		// Verificar que todas las lecturas son consistentes
		for i := 0; i < readCount; i++ {
			retrieved := <-results
			helpers.AssertUserEquals(t, created, retrieved, "Concurrent read")
		}
	})
}
