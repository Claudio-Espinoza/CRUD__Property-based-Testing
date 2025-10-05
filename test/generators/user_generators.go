package generators

import (
	"fmt"
	"math/rand"
	"time"

	"pgregory.net/rapid"
)

type ValidUserData struct {
	Name  string
	Email string
	Age   int
}

type InvalidUserData struct {
	Name     string
	Email    string
	Age      int
	CaseType int
}

// ValidUserStruct genera datos VÁLIDOS para un usuario
func ValidUserStruct() *rapid.Generator[ValidUserData] {
	return rapid.Custom(func(t *rapid.T) ValidUserData {
		return ValidUserData{
			Name:  ValidName().Draw(t, "name"),
			Email: ValidEmail().Draw(t, "email"),
			Age:   ValidAge().Draw(t, "age"),
		}
	})
}

// InvalidUserStruct genera datos INVÁLIDOS para un usuario
func InvalidUserStruct() *rapid.Generator[InvalidUserData] {
	return rapid.Custom(func(t *rapid.T) InvalidUserData {
		caseType := rapid.IntRange(0, 2).Draw(t, "case_type")

		var name, email string
		var age int

		switch caseType {
		case 0: // Nombre inválido
			name = InvalidName().Draw(t, "invalid_name")
			email = ValidEmail().Draw(t, "email")
			age = ValidAge().Draw(t, "age")

		case 1: // Email inválido
			name = ValidName().Draw(t, "name")
			email = InvalidEmail().Draw(t, "invalid_email")
			age = ValidAge().Draw(t, "age")

		case 2: // Edad inválida
			name = ValidName().Draw(t, "name")
			email = ValidEmail().Draw(t, "email")
			age = InvalidAge().Draw(t, "invalid_age")
		}

		return InvalidUserData{
			Name:     name,
			Email:    email,
			Age:      age,
			CaseType: caseType,
		}
	})
}

// ==================== GENERATORS ATÓMICOS ====================

// ValidName genera nombres válidos (2-50 caracteres, sin espacios al inicio/final)
func ValidName() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{
		// Nombres de 2 caracteres (casos borde)
		"AB", "CD", "Jo", "Al", "Bo", "Ed", "Li", "Ty", "Ma", "Lu",

		// Nombres simples comunes
		"John", "Jane", "Alice", "Bob", "Carlos", "Maria",
		"Michael", "Jennifer", "Daniel", "Jessica", "David", "Sarah",
		"Alexander", "Elizabeth", "Christopher", "Amanda",

		// Nombres compuestos
		"John Doe", "Jane Smith", "Alice Johnson", "Bob Wilson",
		"Mary Jane", "Anna Maria", "John Paul", "Sarah Connor",
		"James Bond", "Peter Parker", "Bruce Wayne", "Clark Kent",

		// Nombres largos (cerca del límite de 50)
		"Christopher Alexander Montgomery Wellington",
	})
}

// InvalidName genera nombres REALMENTE inválidos (que fallan DESPUÉS del trim)
func InvalidName() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{
		// Vacío
		"",

		// Solo espacios (después del trim = vacío)
		"   ",
		"     ",

		// Un solo carácter DESPUÉS del trim
		"A",
		"Z",

		// Muy largo (> 50 caracteres)
		"ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"Christopher Alexander Montgomery Wellington Johnson Smith",

		// Caracteres no permitidos (números, símbolos)
		"John123",
		"Jane456",
		"John@Doe",
		"Jane#Smith",
		"John_Doe",
		"Jane-Smith",
		"123John",
		"@Alice",
	})
}

// ValidEmail genera emails válidos con UNICIDAD garantizada
func ValidEmail() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// Generar un sufijo único usando timestamp + random
		timestamp := time.Now().UnixNano()
		randomSuffix := rand.Intn(1000000)

		// Elegir un prefijo base
		prefixes := []string{
			"user", "test", "john", "jane", "alice", "bob",
			"admin", "contact", "info", "demo", "sample",
		}
		prefix := rapid.SampledFrom(prefixes).Draw(t, "email_prefix")

		// Elegir un dominio
		domains := []string{
			"example.com", "test.org", "demo.net", "sample.edu",
			"service.io", "company.co",
		}
		domain := rapid.SampledFrom(domains).Draw(t, "email_domain")

		// Construir email único: prefix+timestamp+random@domain
		return fmt.Sprintf("%s%d%d@%s", prefix, timestamp, randomSuffix, domain)
	})
}

// InvalidEmail genera emails inválidos
func InvalidEmail() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{
		// Sin @
		"invalidemail",
		"userexample.com",

		// Sin dominio
		"user@",
		"test@",

		// Sin usuario
		"@domain.com",
		"@example.org",

		// @ doble
		"user@@domain.com",
		"test@@example.org",

		// Sin TLD
		"user@domain",
		"test@example",

		// Vacío
		"",

		// Espacios
		"user @domain.com",
		"user@ domain.com",
		"user@domain .com",

		// Sin dominio después de @
		"user@.com",
		"test@.org",

		// TLD vacío
		"user@domain.",
		"test@example.",

		// Caracteres inválidos
		"user name@domain.com",
		"user@domain com",
	})
}

// ValidAge genera edades válidas (1-150)
func ValidAge() *rapid.Generator[int] {
	return rapid.IntRange(1, 150)
}

// InvalidAge genera edades inválidas
func InvalidAge() *rapid.Generator[int] {
	return rapid.OneOf(
		rapid.Just(0),            // Cero
		rapid.IntRange(-100, -1), // Negativo
		rapid.IntRange(151, 500), // Muy alto
	)
}
