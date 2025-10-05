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

func TestProperty_UserDelete_ExistingUser_RemovesFromSystem(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")
		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
		helpers.AssertNoError(t, err, "Create user")

		initialCount := svc.CountUsers()

		err = svc.DeleteUser(created.ID)
		helpers.AssertNoError(t, err, "Delete user")

		_, err = svc.GetUser(created.ID)
		helpers.AssertErrorIs(t, err, domain.ErrNotFound, "GetUser after delete")

		if finalCount := svc.CountUsers(); finalCount != initialCount-1 {
			t.Fatalf("Count should decrease by 1: expected %d, got %d", initialCount-1, finalCount)
		}
	})
}

func TestProperty_UserDelete_NonExistentUser_ReturnsNotFound(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		nonExistentID := rapid.StringMatching(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`).
			Draw(t, "non_existent_id")

		err := svc.DeleteUser(nonExistentID)

		helpers.AssertErrorIs(t, err, domain.ErrNotFound, "Delete non-existent user")
	})
}

func TestProperty_UserDelete_Idempotence_SecondDeleteFails(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")
		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
		helpers.AssertNoError(t, err, "Create user")

		err = svc.DeleteUser(created.ID)
		helpers.AssertNoError(t, err, "First delete")

		err = svc.DeleteUser(created.ID)
		helpers.AssertErrorIs(t, err, domain.ErrNotFound, "Second delete")
	})
}

func TestProperty_UserDelete_FreesEmail_AllowsReuse(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")
		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
		helpers.AssertNoError(t, err, "Create user")

		savedEmail := created.Email

		err = svc.DeleteUser(created.ID)
		helpers.AssertNoError(t, err, "Delete user")

		newUserData := generators.ValidUserStruct().Draw(t, "new_user")
		newUser, err := svc.CreateUser(newUserData.Name, savedEmail, newUserData.Age)
		helpers.AssertNoError(t, err, "Create user with freed email")

		if newUser.Email != savedEmail {
			t.Fatalf("Email should be reusable: expected %s, got %s", savedEmail, newUser.Email)
		}
	})
}

func TestProperty_UserDelete_MultipleUsers_OnlyDeletesSpecified(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userCount := rapid.IntRange(3, 10).Draw(t, "user_count")
		createdUsers := make([]*domain.User, 0, userCount)

		for i := 0; i < userCount; i++ {
			userData := generators.ValidUserStruct().Draw(t, "user")
			user, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
			helpers.AssertNoError(t, err, "Create user")
			createdUsers = append(createdUsers, user)
		}

		deleteIndex := rapid.IntRange(0, userCount-1).Draw(t, "delete_index")
		userToDelete := createdUsers[deleteIndex]

		err := svc.DeleteUser(userToDelete.ID)
		helpers.AssertNoError(t, err, "Delete user")

		for i, user := range createdUsers {
			if i == deleteIndex {
				// Este debe estar eliminado
				_, err := svc.GetUser(user.ID)
				helpers.AssertErrorIs(t, err, domain.ErrNotFound, "Deleted user not found")
			} else {
				// Estos deben seguir existiendo
				retrieved, err := svc.GetUser(user.ID)
				helpers.AssertNoError(t, err, "Other user still exists")
				helpers.AssertUserEquals(t, user, retrieved, "Other user unchanged")
			}
		}

		if finalCount := svc.CountUsers(); finalCount != userCount-1 {
			t.Fatalf("Count should be %d, got %d", userCount-1, finalCount)
		}
	})
}

func TestProperty_UserDelete_ConcurrentDeletes_OneSucceedsOthersFail(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userData := generators.ValidUserStruct().Draw(t, "user_data")
		created, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
		helpers.AssertNoError(t, err, "Create user")

		attemptCount := rapid.IntRange(5, 15).Draw(t, "attempt_count")
		results := make(chan error, attemptCount)

		for i := 0; i < attemptCount; i++ {
			go func() {
				err := svc.DeleteUser(created.ID)
				results <- err
			}()
		}

		successCount := 0
		notFoundCount := 0

		for i := 0; i < attemptCount; i++ {
			err := <-results
			if err == nil {
				successCount++
			} else if err == domain.ErrNotFound {
				notFoundCount++
			} else {
				t.Fatalf("Unexpected error: %v", err)
			}
		}
		if successCount != 1 {
			t.Fatalf("Exactly one delete should succeed, got %d successes", successCount)
		}
		if notFoundCount != attemptCount-1 {
			t.Fatalf("Expected %d ErrNotFound, got %d", attemptCount-1, notFoundCount)
		}
		_, err = svc.GetUser(created.ID)
		helpers.AssertErrorIs(t, err, domain.ErrNotFound, "User deleted")
	})
}

func TestProperty_UserDelete_DeleteAll_EmptiesSystem(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := service.NewUserService(repo)

		userCount := rapid.IntRange(1, 10).Draw(t, "user_count")
		createdUsers := make([]*domain.User, 0, userCount)

		for i := 0; i < userCount; i++ {
			userData := generators.ValidUserStruct().Draw(t, "user")
			user, err := svc.CreateUser(userData.Name, userData.Email, userData.Age)
			helpers.AssertNoError(t, err, "Create user")
			createdUsers = append(createdUsers, user)
		}

		for _, user := range createdUsers {
			err := svc.DeleteUser(user.ID)
			helpers.AssertNoError(t, err, "Delete user")
		}

		if count := svc.CountUsers(); count != 0 {
			t.Fatalf("System should be empty, found %d users", count)
		}

		allUsers, err := svc.GetAllUsers()
		helpers.AssertNoError(t, err, "GetAllUsers")

		if len(allUsers) != 0 {
			t.Fatalf("GetAllUsers should return empty list, got %d users", len(allUsers))
		}
	})
}
