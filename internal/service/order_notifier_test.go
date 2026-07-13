package service_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// channelMailer records sends on a channel so tests can await the notifier's
// fire-and-forget goroutines deterministically.
type channelMailer struct {
	sent chan domain.Email
	mu   sync.Mutex
	err  error
}

func newChannelMailer() *channelMailer {
	return &channelMailer{sent: make(chan domain.Email, 16)}
}

func (m *channelMailer) fail(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func (m *channelMailer) Send(_ context.Context, msg domain.Email) error {
	m.mu.Lock()
	err := m.err
	m.mu.Unlock()
	if err != nil {
		return err
	}
	m.sent <- msg
	return nil
}

// wait collects n emails or fails the test after a timeout.
func (m *channelMailer) wait(t *testing.T, n int) []domain.Email {
	t.Helper()
	out := make([]domain.Email, 0, n)
	for len(out) < n {
		select {
		case msg := <-m.sent:
			out = append(out, msg)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for email %d of %d", len(out)+1, n)
		}
	}
	return out
}

// assertQuiet fails if any email arrives within a short window.
func (m *channelMailer) assertQuiet(t *testing.T) {
	t.Helper()
	select {
	case msg := <-m.sent:
		t.Fatalf("unexpected email: %+v", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

// notifierFixture wires an OrderService with a notifier over fakes: users 1
// and 2 are chefs (with emails), user 100 is the customer.
func notifierFixture(t *testing.T) (*service.OrderService, *channelMailer, *fakeMenuItemRepo) {
	t.Helper()
	ctx := context.Background()

	users := newFakeUserRepo()
	for _, u := range []*domain.User{
		{Email: "chef1@test.dev", Username: "chef1", Role: domain.RoleChef},
		{Email: "chef2@test.dev", Username: "chef2", Role: domain.RoleChef},
	} {
		if err := users.Create(ctx, u); err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	customer := &domain.User{Email: "cust@test.dev", Username: "cust", Role: domain.RoleCustomer}
	if err := users.Create(ctx, customer); err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	chefs := newFakeChefRepo()
	for uid, name := range map[int]string{1: "Kitchen One", 2: "Kitchen Two"} {
		if err := chefs.Create(ctx, &domain.Chef{UserID: uid, BusinessName: name, IsActive: true}); err != nil {
			t.Fatalf("seed chef: %v", err)
		}
	}

	items := newFakeMenuItemRepo()
	mailer := newChannelMailer()
	notifier := service.NewOrderNotifier(mailer, users, chefs)
	svc := service.NewOrderService(newFakeOrderRepo(), items, chefs, nil, notifier)
	return svc, mailer, items
}

func placeNotified(t *testing.T, svc *service.OrderService, items *fakeMenuItemRepo) *domain.Order {
	t.Helper()
	a := seedItem(t, items, 1, 5, 10)
	b := seedItem(t, items, 2, 3, 10)
	order, err := svc.PlaceOrder(context.Background(), 3, service.PlaceOrderInput{
		DeliveryAddress: "x", PaymentMethod: domain.PaymentMethodCash,
		Lines: []service.OrderLineInput{
			{MenuItemID: a.ID, Quantity: 1},
			{MenuItemID: b.ID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	return order
}

// Every participating chef gets a "new order" email with only their items.
func TestOrderNotifier_NewOrderEmailsEachChef(t *testing.T) {
	svc, mailer, items := notifierFixture(t)
	order := placeNotified(t, svc, items)

	emails := mailer.wait(t, 2)
	byTo := map[string]domain.Email{}
	for _, e := range emails {
		byTo[e.To] = e
	}
	one, ok1 := byTo["chef1@test.dev"]
	two, ok2 := byTo["chef2@test.dev"]
	if !ok1 || !ok2 {
		t.Fatalf("wrong recipients: %v", byTo)
	}
	if !strings.Contains(one.Subject, order.OrderCode) {
		t.Errorf("subject %q missing order code", one.Subject)
	}
	// Each chef sees their own slice only.
	if !strings.Contains(one.Body, "Kitchen One") || strings.Contains(one.Body, "Kitchen Two") {
		t.Errorf("chef1 body leaks other chef: %q", one.Body)
	}
	if !strings.Contains(one.Body, "$5.00") || strings.Contains(one.Body, "$6.00") {
		t.Errorf("chef1 body has wrong subtotals: %q", one.Body)
	}
	if !strings.Contains(two.Body, "$6.00") {
		t.Errorf("chef2 body missing their subtotal: %q", two.Body)
	}
}

// The customer is emailed on meaningful transitions only: confirm, delivering,
// delivered, decline — not preparing/ready.
func TestOrderNotifier_StatusChangeEmailsCustomer(t *testing.T) {
	svc, mailer, items := notifierFixture(t)
	order := placeNotified(t, svc, items)
	mailer.wait(t, 2) // drain the placement emails
	ctx := context.Background()

	steps := []struct {
		action string
		mails  bool
	}{
		{"confirm", true},
		{"preparing", false},
		{"ready", false},
		{"delivering", true},
		{"delivered", true},
	}
	for _, step := range steps {
		if _, err := svc.AdvanceForChef(ctx, 1, order.ID, step.action); err != nil {
			t.Fatalf("%s: %v", step.action, err)
		}
		if step.mails {
			got := mailer.wait(t, 1)[0]
			if got.To != "cust@test.dev" {
				t.Errorf("%s email to %q, want customer", step.action, got.To)
			}
			if !strings.Contains(got.Subject, order.OrderCode) {
				t.Errorf("%s subject %q missing order code", step.action, got.Subject)
			}
		} else {
			mailer.assertQuiet(t)
		}
	}

	// Decline (chef 2) notifies too.
	if _, err := svc.AdvanceForChef(ctx, 2, order.ID, "decline"); err != nil {
		t.Fatalf("decline: %v", err)
	}
	got := mailer.wait(t, 1)[0]
	if !strings.Contains(got.Subject, "declined") {
		t.Errorf("decline subject = %q, want declined", got.Subject)
	}
}

// Mail failures are logged, never surfaced: the order still places and
// advances when the mailer is down.
func TestOrderNotifier_FailuresNeverBlockOrders(t *testing.T) {
	svc, mailer, items := notifierFixture(t)
	mailer.fail(errors.New("smtp down"))

	order := placeNotified(t, svc, items)
	if order.ID == 0 {
		t.Fatal("order not placed")
	}
	if _, err := svc.AdvanceForChef(context.Background(), 1, order.ID, "confirm"); err != nil {
		t.Fatalf("advance with failing mailer: %v", err)
	}
	mailer.assertQuiet(t)
}
