package mailer

import (
	"context"
	"net/smtp"
	"strings"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

func TestSMTP_SendBuildsMessageAndDelivers(t *testing.T) {
	m := NewSMTP("smtp.example.com", "587", "user", "pass", "no-reply@example.com")

	var gotAddr, gotFrom string
	var gotTo []string
	var gotMsg []byte
	m.sendMail = func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
		gotAddr, gotFrom, gotTo, gotMsg = addr, from, to, msg
		return nil
	}

	err := m.Send(context.Background(), domain.Email{
		To:      "user@example.com",
		Subject: "Reset your password",
		Body:    "line one\nline two",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if gotAddr != "smtp.example.com:587" {
		t.Errorf("addr = %q", gotAddr)
	}
	if gotFrom != "no-reply@example.com" || len(gotTo) != 1 || gotTo[0] != "user@example.com" {
		t.Errorf("envelope wrong: from=%q to=%v", gotFrom, gotTo)
	}
	msg := string(gotMsg)
	for _, want := range []string{
		"From: no-reply@example.com\r\n",
		"To: user@example.com\r\n",
		"Subject: Reset your password\r\n",
		"line one\r\nline two", // body normalised to CRLF
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q\n--- got ---\n%s", want, msg)
		}
	}
}

func TestSMTP_NoAuthWhenUsernameEmpty(t *testing.T) {
	if m := NewSMTP("localhost", "25", "", "", "from@example.com"); m.auth != nil {
		t.Error("expected no SMTP auth when username is empty")
	}
}

func TestLogging_SendIsNoError(t *testing.T) {
	if err := NewLogging(nil).Send(context.Background(), domain.Email{To: "a@b.c", Subject: "s", Body: "b"}); err != nil {
		t.Errorf("logging mailer send = %v, want nil", err)
	}
}
