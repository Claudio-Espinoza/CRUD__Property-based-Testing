package main

import (
	"fmt"
	"log"

	"property-based/internal/repository"
	"property-based/internal/service"
)

func main() {
	repo := repository.NewInMemoryUserRepository()
	svc := service.NewUserService(repo)

	user1, err := svc.CreateUser("John Doe", "john@example.com", 30)
	if err != nil {
		log.Fatalf("Error creating user1: %v", err)
	}
	fmt.Printf("Created user: %+v\n", user1)
	user2, err := svc.CreateUser("Jane Smith", "jane@example.com", 25)
	if err != nil {
		log.Fatalf("Error creating user2: %v", err)
	}
	fmt.Printf("Created user: %+v\n", user2)

	users, err := svc.GetAllUsers()
	if err != nil {
		log.Fatalf("Error getting users: %v", err)
	}

	fmt.Println("All users:")
	for _, u := range users {
		fmt.Printf("  - %s (%s)\n", u.Name, u.Email)
	}

	found, err := svc.GetUserByEmail("john@example.com")
	if err != nil {
		log.Fatalf("Error finding user: %v", err)
	}
	fmt.Printf("Found user by email: %+v\n", found)

	updated, err := svc.UpdateUser(user1.ID, "John Updated", "john.updated@example.com", 31)
	if err != nil {
		log.Fatalf("Error updating user: %v", err)
	}
	fmt.Printf("Updated user: %+v\n", updated)

	if err := svc.DeleteUser(user2.ID); err != nil {
		log.Fatalf("Error deleting user: %v", err)
	}
	fmt.Println("User deleted successfully")

	count := svc.CountUsers()
	fmt.Printf("Final user count: %d\n", count)

}
