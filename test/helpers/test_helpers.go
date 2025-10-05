package helpers

import (
	"property-based/internal/domain"
)

// Interfaz que implementan tanto *testing.T como *rapid.T
type TestingT interface {
	Helper()
	Fatalf(format string, args ...interface{})
}

// AssertUserEquals verifica que dos usuarios sean iguales (excepto timestamps)
func AssertUserEquals(t TestingT, expected, actual *domain.User, context string) {
	t.Helper()

	if actual.ID != expected.ID {
		t.Fatalf("%s: ID mismatch - expected %s, got %s", context, expected.ID, actual.ID)
	}
	if actual.Name != expected.Name {
		t.Fatalf("%s: Name mismatch - expected %s, got %s", context, expected.Name, actual.Name)
	}
	if actual.Email != expected.Email {
		t.Fatalf("%s: Email mismatch - expected %s, got %s", context, expected.Email, actual.Email)
	}
	if actual.Age != expected.Age {
		t.Fatalf("%s: Age mismatch - expected %d, got %d", context, expected.Age, actual.Age)
	}
}

// AssertNoError verifica que no haya error
func AssertNoError(t TestingT, err error, context string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: expected no error, got: %v", context, err)
	}
}

// AssertError verifica que haya un error
func AssertError(t TestingT, err error, context string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got nil", context)
	}
}

// AssertErrorIs verifica que el error sea de un tipo específico
func AssertErrorIs(t TestingT, err, expectedErr error, context string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error %v, got nil", context, expectedErr)
	}
	if err != expectedErr {
		t.Fatalf("%s: expected error %v, got %v", context, expectedErr, err)
	}
}
