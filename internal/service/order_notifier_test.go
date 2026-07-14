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
// and 2 are chefs (with emails), user 3 is the customer. All start opted in;
// tests flip the flag through the user repo.
func notifierFixture(t *testing.T) (*service.OrderService, *channelMailer, *fakeMenuItemRepo, *fakeUserRepo) {
	t.Helper()
	ctx := context.Background()

	users := newFakeUserRepo()
	for _, u := range []*domain.User{
		{Email: "chef1@test.dev", Username: "chef1", Role: domain.RoleChef, EmailNotifications: true},
		{Email: "chef2@test.dev", Username: "chef2", Role: domain.RoleChef, EmailNotifications: true},
	} {
		if err := users.Create(ctx, u); err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	customer := &domain.User{Email: "cust@test.dev", Username: "cust", Role: domain.RoleCustomer, EmailNotifications: true}
	if err := users.Create(ctx, customer); err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	chefs := newFakeChefRepo()
	// Seed in a fixed order: the fake assigns chef IDs sequentially, and the
	// assertions rely on user 1 owning chef 1 ("Kitchen One") — a map here
	// would randomise the pairing and flake.
	for i, name := range []string{"Kitchen One", "Kitchen Two"} {
		if err := chefs.Create(ctx, &domain.Chef{UserID: i + 1, BusinessName: name, IsActive: true}); err != nil {
			t.Fatalf("seed chef: %v", err)
		}
	}

	items := newFakeMenuItemRepo()
	mailer := newChannelMailer()
	notifier := service.NewOrderNotifier(mailer, users, chefs)
	svc := service.NewOrderService(newFakeOrderRepo(), items, chefs, nil, nil, nil, nil, notifier)
	return svc, mailer, items, users
}

// setOptOut flips a user's email-notification preference through the repo.
func setOptOut(t *testing.T, users *fakeUserRepo, userID int, optedIn bool) {
	t.Helper()
	u, err := users.FindByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("find user %d: %v", userID, err)
	}
	u.EmailNotifications = optedIn
	if err := users.UpdateProfile(context.Background(), u); err != nil {
		t.Fatalf("set opt-out: %v", err)
	}
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
	svc, mailer, items, _ := notifierFixture(t)
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
	svc, mailer, items, _ := notifierFixture(t)
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
	svc, mailer, items, _ := notifierFixture(t)
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

// Opted-out recipients get no order emails: an opted-out chef stays silent on
// placement (the other chef still gets theirs), and an opted-out customer
// stays silent on status changes.
func TestOrderNotifier_RespectsOptOut(t *testing.T) {
	svc, mailer, items, users := notifierFixture(t)
	setOptOut(t, users, 1, false) // chef 1 (user 1) opts out
	setOptOut(t, users, 3, false) // the customer (user 3) opts out

	order := placeNotified(t, svc, items)

	// Only chef 2's placement email arrives.
	got := mailer.wait(t, 1)[0]
	if got.To != "chef2@test.dev" {
		t.Errorf("placement email to %q, want only chef2 (chef1 opted out)", got.To)
	}
	mailer.assertQuiet(t)

	// A notify-worthy transition stays silent for the opted-out customer.
	if _, err := svc.AdvanceForChef(context.Background(), 1, order.ID, "confirm"); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	mailer.assertQuiet(t)

	// Opting back in resumes delivery.
	setOptOut(t, users, 3, true)
	if _, err := svc.AdvanceForChef(context.Background(), 1, order.ID, "preparing"); err != nil {
		t.Fatalf("preparing: %v", err)
	}
	if _, err := svc.AdvanceForChef(context.Background(), 1, order.ID, "ready"); err != nil {
		t.Fatalf("ready: %v", err)
	}
	if _, err := svc.AdvanceForChef(context.Background(), 1, order.ID, "delivering"); err != nil {
		t.Fatalf("delivering: %v", err)
	}
	back := mailer.wait(t, 1)[0]
	if back.To != "cust@test.dev" {
		t.Errorf("post-opt-in email to %q, want the customer", back.To)
	}
}
