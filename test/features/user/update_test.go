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
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ user existente, ∀ datos válidos,
//	  UpdateUser(id, datos) → usuario actualizado ∧
//	  GetUser(id) = usuario actualizado
//
// INVARIANTES:
//  1. ID permanece inmutable
//  2. CreatedAt permanece inmutable
//  3. UpdatedAt >= CreatedAt
//  4. UpdatedAt nuevo >= UpdatedAt anterior
func TestProperty_UserUpdate_ValidData_SucceedsAndPersists(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		// Crear usuario inicial
		initialData := generators.ValidUserStruct().Draw(t, "initial_data")
		created, err := svc.CreateUser(initialData.Name, initialData.Email, initialData.Age)
		helpers.AssertNoError(t, err, "Create user")

		// Generar nuevos datos válidos
		updateData := generators.ValidUserStruct().Draw(t, "update_data")

		// Actualizar
		updated, err := svc.UpdateUser(created.ID, updateData.Name, updateData.Email, updateData.Age)
		helpers.AssertNoError(t, err, "Update user")

		// Verificar invariantes
		if updated.ID != created.ID {
			t.Fatalf("ID should be immutable: expected %s, got %s", created.ID, updated.ID)
		}

		if !updated.CreatedAt.Equal(created.CreatedAt) {
			t.Fatal("CreatedAt should be immutable")
		}

		if updated.UpdatedAt.Before(created.UpdatedAt) {
			t.Fatal("UpdatedAt should increase")
		}

		// Verificar nuevos datos
		if updated.Name != updateData.Name {
			t.Fatalf("Name not updated: expected %s, got %s", updateData.Name, updated.Name)
		}
		if updated.Email != updateData.Email {
			t.Fatalf("Email not updated: expected %s, got %s", updateData.Email, updated.Email)
		}
		if updated.Age != updateData.Age {
			t.Fatalf("Age not updated: expected %d, got %d", updateData.Age, updated.Age)
		}

		// Verificar persistencia
		retrieved, err := svc.GetUser(updated.ID)
		helpers.AssertNoError(t, err, "GetUser after update")
		helpers.AssertUserEquals(t, updated, retrieved, "Updated user persisted")
	})
}

// TestProperty_UserUpdate_InvalidData_FailsWithoutModifying
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ user existente, ∀ datos inválidos,
//	  UpdateUser(id, datos) → (nil, error) ∧
//	  GetUser(id) = user_original (sin cambios)
func TestProperty_UserUpdate_InvalidData_FailsWithoutModifying(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		// Crear usuario válido
		validData := generators.ValidUserStruct().Draw(t, "valid_data")
		created, err := svc.CreateUser(validData.Name, validData.Email, validData.Age)
		helpers.AssertNoError(t, err, "Create user")

		// Intentar actualizar con datos inválidos
		invalidData := generators.InvalidUserStruct().Draw(t, "invalid_data")

		updated, err := svc.UpdateUser(created.ID, invalidData.Name, invalidData.Email, invalidData.Age)

		helpers.AssertError(t, err, "Update with invalid data")

		if updated != nil {
			t.Fatalf("UpdateUser should return nil on error, got: %+v", updated)
		}

		// Verificar que el usuario original no cambió
		retrieved, err := svc.GetUser(created.ID)
		helpers.AssertNoError(t, err, "GetUser after failed update")
		helpers.AssertUserEquals(t, created, retrieved, "User unchanged after failed update")
	})
}

// TestProperty_UserUpdate_NonExistentUser_ReturnsNotFound
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ id ∉ Sistema, ∀ datos,
//	  UpdateUser(id, datos) → (nil, ErrNotFound)
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
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ user1, user2 (user1.ID ≠ user2.ID),
//	  UpdateUser(user1.ID, ..., user2.Email, ...) → (nil, ErrAlreadyExists)
func TestProperty_UserUpdate_DuplicateEmail_Fails(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		// Crear dos usuarios
		user1Data := generators.ValidUserStruct().Draw(t, "user1")
		user1, err := svc.CreateUser(user1Data.Name, user1Data.Email, user1Data.Age)
		helpers.AssertNoError(t, err, "Create user1")

		user2Data := generators.ValidUserStruct().Draw(t, "user2")
		user2, err := svc.CreateUser(user2Data.Name, user2Data.Email, user2Data.Age)
		helpers.AssertNoError(t, err, "Create user2")

		// Intentar actualizar user1 con el email de user2
		updated, err := svc.UpdateUser(user1.ID, user1.Name, user2.Email, user1.Age)

		helpers.AssertErrorIs(t, err, domain.ErrAlreadyExists, "Update with duplicate email")

		if updated != nil {
			t.Fatal("Update should fail with duplicate email")
		}

		// Verificar que user1 no cambió
		retrieved, err := svc.GetUser(user1.ID)
		helpers.AssertNoError(t, err, "GetUser after failed update")
		helpers.AssertUserEquals(t, user1, retrieved, "User1 unchanged")
	})
}

// TestProperty_UserUpdate_MultipleSequentialUpdates_EachPersists
//
// PROPIEDAD MATEMÁTICA:
//
//	∀ user, ∀ secuencia de updates [u1, u2, ..., un],
//	  Update(u1) → Update(u2) → ... → Update(un) ∧
//	  GetUser() = un
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
