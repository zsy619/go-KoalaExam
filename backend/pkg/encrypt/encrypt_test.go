package encrypt_test

import (
	"testing"

	"github.com/your-team/koala-exam-backend/pkg/encrypt"
)

func TestBcryptPassword(t *testing.T) {
	pwd := "koala123"
	hashed, err := encrypt.BcryptPassword(pwd)
	if err != nil { t.Fatal(err) }
	if hashed == "" { t.Fatal("hash empty") }
	if hashed == pwd { t.Fatal("hash equals plain") }
	if !encrypt.BcryptCheck(hashed, pwd) { t.Fatal("check failed") }
	if encrypt.BcryptCheck(hashed, "wrong") { t.Fatal("should not match") }
}

func TestSHA256Hex(t *testing.T) {
	h := encrypt.SHA256Hex("salt", "data")
	if len(h) != 64 { t.Fatalf("expected 64 chars, got %d", len(h)) }
}
