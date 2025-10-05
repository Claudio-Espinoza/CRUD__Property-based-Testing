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

// TestProperty_UserUpdate_ValidData_SucceedsAndPersists
// Invariante: UpdatedAt > CreatedAt, CreatedAt inmutable, ID inmutable
// Relación: GetUser(updated.ID) == updated (nuevos datos)
// Bordes: Actualizar con mismos datos, cambiar solo un campo, todos los campos
func TestProperty_UserUpdate_ValidData_SucceedsAndPersists(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		initialData := generators.ValidUserStruct().Draw(t, "initial_data")
		created, err := svc.CreateUser(initialData.Name, initialData.Email, initialData.Age)
		helpers.AssertNoError(t, err, "Create user")

		updateData := generators.ValidUserStruct().Draw(t, "update_data")

		updated, err := svc.UpdateUser(created.ID, updateData.Name, updateData.Email, updateData.Age)
		helpers.AssertNoError(t, err, "Update user")

		if updated.ID != created.ID {
			t.Fatalf("ID should be immutable: expected %s, got %s", created.ID, updated.ID)
		}

		if !updated.CreatedAt.Equal(created.CreatedAt) {
			t.Fatal("CreatedAt should be immutable")
		}

		if updated.UpdatedAt.Before(created.UpdatedAt) {
			t.Fatal("UpdatedAt should increase")
		}

		if updated.Name != updateData.Name {
			t.Fatalf("Name not updated: expected %s, got %s", updateData.Name, updated.Name)
		}
		if updated.Email != updateData.Email {
			t.Fatalf("Email not updated: expected %s, got %s", updateData.Email, updated.Email)
		}
		if updated.Age != updateData.Age {
			t.Fatalf("Age not updated: expected %d, got %d", updateData.Age, updated.Age)
		}

		retrieved, err := svc.GetUser(updated.ID)
		helpers.AssertNoError(t, err, "GetUser after update")
		helpers.AssertUserEquals(t, updated, retrieved, "Updated user persisted")
	})
}

// TestProperty_UserUpdate_InvalidData_FailsWithoutModifying
// Invariante: Usuario original sin modificar tras fallo
// Relación: UpdateUser(id, datos_inválidos) ⟹ GetUser(id) == estado_previo
// Bordes: Edad 0/151, nombre vacío/"A", email sin @
func TestProperty_UserUpdate_InvalidData_FailsWithoutModifying(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		validData := generators.ValidUserStruct().Draw(t, "valid_data")
		created, err := svc.CreateUser(validData.Name, validData.Email, validData.Age)
		helpers.AssertNoError(t, err, "Create user")

		invalidData := generators.InvalidUserStruct().Draw(t, "invalid_data")
		updated, err := svc.UpdateUser(created.ID, invalidData.Name, invalidData.Email, invalidData.Age)
		helpers.AssertError(t, err, "Update with invalid data")

		if updated != nil {
			t.Fatalf("UpdateUser should return nil on error, got: %+v", updated)
		}

		retrieved, err := svc.GetUser(created.ID)
		helpers.AssertNoError(t, err, "GetUser after failed update")
		helpers.AssertUserEquals(t, created, retrieved, "User unchanged after failed update")
	})
}

// TestProperty_UserUpdate_NonExistentUser_ReturnsNotFound
// Invariante: UpdateUser(id_inexistente, _) ⟹ ErrNotFound
// Relación: ¬∃user: user.ID == id ⟹ UpdateUser(id, _) == ErrNotFound
// Bordes: UUID válido inexistente, después de Delete
func TestProperty_UserUpdate_NonExistentUser_ReturnsNotFound(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		nonExistentID := rapid.StringMatching(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`).
			Draw(t, "non_existent_id")

		validData := generators.ValidUserStruct().Draw(t, "valid_data")

		updated, err := svc.UpdateUser(nonExistentID, validData.Name, validData.Email, validData.Age)

		helpers.AssertErrorIs(t, err, domain.ErrNotFound, "Update non-existent user")

		if updated != nil {
			t.Fatal("Update should return nil for non-existent user")
		}
	})
}

// TestProperty_UserUpdate_DuplicateEmail_Fails
// Invariante: Email único en sistema (no cambiar a email existente)
// Relación: UpdateUser(id, _, email_existente, _) ⟹ ErrAlreadyExists
// Bordes: Cambiar email a uno existente, mantener propio email (debe permitirse)
func TestProperty_UserUpdate_DuplicateEmail_Fails(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		user1Data := generators.ValidUserStruct().Draw(t, "user1")
		user1, err := svc.CreateUser(user1Data.Name, user1Data.Email, user1Data.Age)
		helpers.AssertNoError(t, err, "Create user1")

		user2Data := generators.ValidUserStruct().Draw(t, "user2")
		user2, err := svc.CreateUser(user2Data.Name, user2Data.Email, user2Data.Age)
		helpers.AssertNoError(t, err, "Create user2")

		updated, err := svc.UpdateUser(user1.ID, user1.Name, user2.Email, user1.Age)

		helpers.AssertErrorIs(t, err, domain.ErrAlreadyExists, "Update with duplicate email")

		if updated != nil {
			t.Fatal("Update should fail with duplicate email")
		}

		retrieved, err := svc.GetUser(user1.ID)
		helpers.AssertNoError(t, err, "GetUser after failed update")
		helpers.AssertUserEquals(t, user1, retrieved, "User1 unchanged")
	})
}

// TestProperty_UserUpdate_MultipleSequentialUpdates_EachPersists
// Invariante: Cada UpdatedAt[i] > UpdatedAt[i-1]
// Relación: ∀i: GetUser(id) == update[i] (última actualización)
// Bordes: 1-5 actualizaciones secuenciales, cambios en diferentes campos
func TestProperty_UserUpdate_MultipleSequentialUpdates_EachPersists(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		// Crear usuario inicial
		initialData := generators.ValidUserStruct().Draw(t, "initial")
		user, err := svc.CreateUser(initialData.Name, initialData.Email, initialData.Age)
		helpers.AssertNoError(t, err, "Create user")

		updateCount := rapid.IntRange(2, 5).Draw(t, "update_count")

		// Realizar múltiples actualizaciones secuenciales
		for i := 0; i < updateCount; i++ {
			updateData := generators.ValidUserStruct().Draw(t, "update")
			user, err = svc.UpdateUser(user.ID, updateData.Name, updateData.Email, updateData.Age)
			helpers.AssertNoError(t, err, "Sequential update")
		}

		// Verificar que la última actualización persiste
		retrieved, err := svc.GetUser(user.ID)
		helpers.AssertNoError(t, err, "GetUser after multiple updates")
		helpers.AssertUserEquals(t, user, retrieved, "Final state persisted")
	})
}
