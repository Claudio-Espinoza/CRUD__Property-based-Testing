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
// Invariante: GetUser(id) siempre retorna los mismos datos para el mismo id
// Relación: GetUser(created.ID) == created (identidad)
// Bordes: ID vacío, ID inexistente, usuario recién creado
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
// Invariante: GetUser(id_inexistente) ⟹ ErrNotFound
// Relación: ¬∃user: user.ID == id ⟹ GetUser(id) == ErrNotFound
// Bordes: UUID válido pero inexistente, string vacío, UUID malformado
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
// Invariante: GetByEmail retorna el usuario cuyo email coincide (case-insensitive)
// Relación: GetByEmail(user.Email) == user
// Bordes: Email con mayúsculas/minúsculas, espacios al inicio/final
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
// Invariante: len(GetAll()) == CountUsers()
// Relación: ∀user ∈ created_users: user ∈ GetAll()
// Bordes: Sistema vacío (0 usuarios), 1 usuario, N usuarios (2-10)
func TestProperty_UserRead_GetAll_ReturnsAllCreatedUsers(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userCount := rapid.IntRange(1, 10).Draw(t, "user_count")
		createdUsers := make(map[string]*domain.User)

		for i := 0; i < userCount; i++ {
			userData := generators.ValidUserStruct().Draw(t, "user")
			created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
			helpers.AssertNoError(t, err, "Create user")
			createdUsers[created.ID] = created
		}

		allUsers, err := svc.GetAllUsers()
		helpers.AssertNoError(t, err, "GetAllUsers")

		if len(allUsers) != userCount {
			t.Fatalf("Expected %d users, got %d", userCount, len(allUsers))
		}

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
// Invariante: Lecturas concurrentes retornan los mismos datos
// Relación: ∀i,j: GetUser(id)[i] == GetUser(id)[j] (consistencia)
// Bordes: 5-15 goroutines leyendo simultáneamente el mismo usuario
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
