package passwordbcrypt

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("123456")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !CheckPassword("123456", hash) {
		t.Fatal("CheckPassword() = false, want true")
	}

	if CheckPassword("wrong-password", hash) {
		t.Fatal("CheckPassword() = true for wrong password, want false")
	}
}
