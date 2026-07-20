package service_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/payment"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakePaymentSessionRepo is an in-memory domain.PaymentSessionRepository.
type fakePaymentSessionRepo struct {
	mu       sync.Mutex
	sessions map[int]*domain.PaymentSession
	nextID   int
}

func newFakePaymentSessionRepo() *fakePaymentSessionRepo {
	return &fakePaymentSessionRepo{sessions: map[int]*domain.PaymentSession{}, nextID: 1}
}

func (f *fakePaymentSessionRepo) Create(_ context.Context, s *domain.PaymentSession) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s.ID = f.nextID
	f.nextID++
	if s.Status == "" {
		s.Status = domain.PaymentSessionInitiated
	}
	s.CreatedAt = time.Now()
	cp := *s
	f.sessions[s.ID] = &cp
	return nil
}
func (f *fakePaymentSessionRepo) FindByToken(_ context.Context, token string) (*domain.PaymentSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.sessions {
		if s.Token == token {
			cp := *s
			return &cp, nil
		}
	}
	return nil, domain.ErrPaymentSessionNotFound
}
func (f *fakePaymentSessionRepo) FindPaidByOrder(_ context.Context, orderID int) (*domain.PaymentSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.sessions {
		if s.OrderID == orderID && s.Status == domain.PaymentSessionPaid {
			cp := *s
			return &cp, nil
		}
	}
	return nil, domain.ErrPaymentSessionNotFound
}
func (f *fakePaymentSessionRepo) UpdateStatus(_ context.Context, id int, status string, paymentID *string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[id]
	if !ok {
		return domain.ErrPaymentSessionNotFound
	}
	s.Status = status
	if paymentID != nil {
		s.PaymentID = paymentID
	}
	return nil
}

// fakePaymentMethodRepo is an in-memory domain.PaymentMethodRepository.
type fakePaymentMethodRepo struct {
	mu     sync.Mutex
	cards  []*domain.SavedCard
	nextID int
}

func newFakePaymentMethodRepo() *fakePaymentMethodRepo {
	return &fakePaymentMethodRepo{nextID: 1}
}

func (f *fakePaymentMethodRepo) Add(_ context.Context, c *domain.SavedCard) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.cards {
		if existing.UserID == c.UserID && existing.CardToken == c.CardToken {
			c.ID = existing.ID
			c.CreatedAt = existing.CreatedAt
			return nil // idempotent
		}
	}
	c.ID = f.nextID
	f.nextID++
	c.CreatedAt = time.Now()
	cp := *c
	f.cards = append(f.cards, &cp)
	return nil
}
func (f *fakePaymentMethodRepo) ListByUser(_ context.Context, userID int) ([]*domain.SavedCard, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*domain.SavedCard
	for i := len(f.cards) - 1; i >= 0; i-- {
		if f.cards[i].UserID == userID {
			cp := *f.cards[i]
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakePaymentMethodRepo) FindByToken(_ context.Context, userID int, token string) (*domain.SavedCard, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.cards {
		if c.UserID == userID && c.CardToken == token {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrCardNotFound
}
func (f *fakePaymentMethodRepo) CardUserKey(_ context.Context, userID int) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.cards) - 1; i >= 0; i-- {
		if f.cards[i].UserID == userID {
			return f.cards[i].CardUserKey, nil
		}
	}
	return "", nil
}
func (f *fakePaymentMethodRepo) Delete(_ context.Context, userID int, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, c := range f.cards {
		if c.UserID == userID && c.CardToken == token {
			f.cards = append(f.cards[:i], f.cards[i+1:]...)
			return nil
		}
	}
	return domain.ErrCardNotFound
}

// paymentFixture wires a PaymentService over fakes + the mock gateway, seeds a
// user (buyer) and returns the order repo for placing orders.
func paymentFixture(t *testing.T) (*service.PaymentService, *fakeOrderRepo, *fakePaymentSessionRepo) {
	t.Helper()
	users := newFakeUserRepo()
	if err := users.Create(context.Background(), &domain.User{Username: "cust", Email: "c@e.com", Role: domain.RoleCustomer, IsActive: true}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	orders := newFakeOrderRepo()
	sessions := newFakePaymentSessionRepo()
	svc := service.NewPaymentService(sessions, orders, users, payment.NewMock("http://app.test"), "http://app.test")
	return svc, orders, sessions
}

func seedCardOrder(t *testing.T, orders *fakeOrderRepo, userID int) *domain.Order {
	t.Helper()
	method := domain.PaymentMethodCard
	o := domain.NewOrder(userID, "1 St")
	o.PaymentMethod = &method
	o.TotalPrice = 10
	o.Subtotal = 10
	o.Items = []*domain.OrderItem{{MenuItemID: 1, ChefID: 1, ItemName: "Dish", Quantity: 1, UnitPrice: 10, Subtotal: 10}}
	if err := orders.Create(context.Background(), o); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	return o
}

func TestPaymentService_StartCheckout(t *testing.T) {
	svc, orders, _ := paymentFixture(t)
	ctx := context.Background()
	order := seedCardOrder(t, orders, 1)

	url, err := svc.StartCheckout(ctx, 1, order.ID, false)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if !strings.HasPrefix(url, "http://app.test/mock-pay?token=") {
		t.Errorf("payment page = %q", url)
	}

	// Guards: wrong owner, cash order, already paid.
	if _, err := svc.StartCheckout(ctx, 99, order.ID, false); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("foreign order = %v, want ErrForbidden", err)
	}
	cash := domain.NewOrder(1, "1 St")
	m := domain.PaymentMethodCash
	cash.PaymentMethod = &m
	_ = orders.Create(ctx, cash)
	if _, err := svc.StartCheckout(ctx, 1, cash.ID, false); !errors.Is(err, domain.ErrOrderNotPayable) {
		t.Errorf("cash order = %v, want ErrOrderNotPayable", err)
	}
}

func TestPaymentService_CompleteCheckout(t *testing.T) {
	svc, orders, sessions := paymentFixture(t)
	ctx := context.Background()
	order := seedCardOrder(t, orders, 1)

	pageURL, err := svc.StartCheckout(ctx, 1, order.ID, false)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	token := strings.TrimPrefix(pageURL, "http://app.test/mock-pay?token=")

	orderID, paid, err := svc.CompleteCheckout(ctx, token)
	if err != nil || !paid || orderID != order.ID {
		t.Fatalf("complete = (%d,%v,%v), want (%d,true,nil)", orderID, paid, err, order.ID)
	}
	got, _ := orders.FindByID(ctx, order.ID)
	if got.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("order payment = %q, want paid", got.PaymentStatus)
	}
	session, _ := sessions.FindByToken(ctx, token)
	if session.Status != domain.PaymentSessionPaid || session.PaymentID == nil {
		t.Errorf("session = %+v, want paid with payment id", session)
	}

	// Replay is idempotent.
	if _, paid, err := svc.CompleteCheckout(ctx, token); err != nil || !paid {
		t.Errorf("replay = (%v,%v), want (true,nil)", paid, err)
	}

	// Failure outcome marks the session failed and leaves the order pending.
	order2 := seedCardOrder(t, orders, 1)
	page2, _ := svc.StartCheckout(ctx, 1, order2.ID, false)
	token2 := strings.TrimPrefix(page2, "http://app.test/mock-pay?token=")
	_, paid, err = svc.CompleteCheckout(ctx, token2+":fail")
	if err != nil || paid {
		t.Fatalf("failed checkout = (%v,%v), want (false,nil)", paid, err)
	}
	got2, _ := orders.FindByID(ctx, order2.ID)
	if got2.PaymentStatus != domain.PaymentStatusPending {
		t.Errorf("failed order payment = %q, want pending", got2.PaymentStatus)
	}

	// Unknown token surfaces not-found.
	if _, _, err := svc.CompleteCheckout(ctx, "ghost"); !errors.Is(err, domain.ErrPaymentSessionNotFound) {
		t.Errorf("unknown token = %v, want ErrPaymentSessionNotFound", err)
	}
}

func TestPaymentService_SavedCards(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserRepo()
	_ = users.Create(ctx, &domain.User{Username: "cust", Email: "c@e.com", Role: domain.RoleCustomer, IsActive: true})
	orders := newFakeOrderRepo()
	sessions := newFakePaymentSessionRepo()
	cards := newFakePaymentMethodRepo()
	svc := service.NewPaymentService(sessions, orders, users, payment.NewMock("http://app.test"), "http://app.test")
	svc.SetPaymentMethods(cards)

	// No cards yet.
	if list, err := svc.ListSavedCards(ctx, 1); err != nil || len(list) != 0 {
		t.Fatalf("initial list = (%v,%v), want empty", list, err)
	}

	// Checkout WITHOUT opting to save: nothing is stored.
	order := seedCardOrder(t, orders, 1)
	page, err := svc.StartCheckout(ctx, 1, order.ID, false)
	if err != nil {
		t.Fatalf("start (no save): %v", err)
	}
	token := strings.TrimPrefix(page, "http://app.test/mock-pay?token=")
	if _, paid, err := svc.CompleteCheckout(ctx, token); err != nil || !paid {
		t.Fatalf("complete (no save) = (%v,%v)", paid, err)
	}
	if list, _ := svc.ListSavedCards(ctx, 1); len(list) != 0 {
		t.Fatalf("no-save checkout stored %d cards, want 0", len(list))
	}

	// Checkout opting to save: the card is persisted.
	order2 := seedCardOrder(t, orders, 1)
	page2, err := svc.StartCheckout(ctx, 1, order2.ID, true)
	if err != nil {
		t.Fatalf("start (save): %v", err)
	}
	token2 := strings.TrimPrefix(page2, "http://app.test/mock-pay?token=")
	if _, paid, err := svc.CompleteCheckout(ctx, token2); err != nil || !paid {
		t.Fatalf("complete (save) = (%v,%v)", paid, err)
	}
	list, err := svc.ListSavedCards(ctx, 1)
	if err != nil || len(list) != 1 {
		t.Fatalf("after save list = (%d,%v), want 1 card", len(list), err)
	}
	saved := list[0]
	if saved.MaskedNumber == "" || saved.CardToken == "" {
		t.Errorf("saved card missing metadata: %+v", saved)
	}
	if strings.Contains(saved.MaskedNumber, "0006") == false {
		t.Errorf("masked number = %q, want masked digits", saved.MaskedNumber)
	}

	// The saved card is scoped to its owner — another user sees none.
	if list, _ := svc.ListSavedCards(ctx, 2); len(list) != 0 {
		t.Errorf("foreign user sees %d cards, want 0", len(list))
	}

	// Delete: unknown token → ErrCardNotFound; own token → removed.
	if err := svc.DeleteSavedCard(ctx, 1, "ghost"); !errors.Is(err, domain.ErrCardNotFound) {
		t.Errorf("delete unknown = %v, want ErrCardNotFound", err)
	}
	if err := svc.DeleteSavedCard(ctx, 2, saved.CardToken); !errors.Is(err, domain.ErrCardNotFound) {
		t.Errorf("delete as foreign user = %v, want ErrCardNotFound", err)
	}
	if err := svc.DeleteSavedCard(ctx, 1, saved.CardToken); err != nil {
		t.Fatalf("delete own card: %v", err)
	}
	if list, _ := svc.ListSavedCards(ctx, 1); len(list) != 0 {
		t.Errorf("after delete list = %d, want 0", len(list))
	}
}

// recordingRefunder captures refund calls for the order-service cancel and
// decline paths.
type recordingRefunder struct {
	calls          int
	partialCalls   int
	partialAmounts []float64
	err            error
}

func (r *recordingRefunder) RefundOrderPayment(context.Context, *domain.Order) error {
	r.calls++
	return r.err
}

func (r *recordingRefunder) RefundSubOrderPayment(_ context.Context, _ *domain.Order, amount float64) error {
	r.partialCalls++
	r.partialAmounts = append(r.partialAmounts, amount)
	return r.err
}

func TestOrderService_CancelRefundsPaidCardOrders(t *testing.T) {
	ctx := context.Background()
	chefRepo := newFakeChefRepo()
	_ = chefRepo.Create(ctx, &domain.Chef{UserID: 1, IsActive: true})
	items := newFakeMenuItemRepo()
	item := seedItem(t, items, 1, 5, 10)
	orders := newFakeOrderRepo()
	refunder := &recordingRefunder{}
	svc := service.NewOrderService(orders, items, chefRepo, nil, nil, nil, domain.FeePolicy{}, refunder, nil)

	place := func() *domain.Order {
		o, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
			DeliveryAddress: "x", PaymentMethod: domain.PaymentMethodCard,
			Lines: []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("place: %v", err)
		}
		return o
	}

	// Paid card order -> refund called, payment becomes refunded.
	paidOrder := place()
	stored, _ := orders.FindByID(ctx, paidOrder.ID)
	_ = stored.MarkPaid()
	_ = orders.UpdateStatus(ctx, stored)
	cancelled, err := svc.CancelForCustomer(ctx, 100, paidOrder.ID)
	if err != nil {
		t.Fatalf("cancel paid: %v", err)
	}
	if refunder.calls != 1 || cancelled.PaymentStatus != domain.PaymentStatusRefunded {
		t.Errorf("refunds=%d payment=%q, want 1/refunded", refunder.calls, cancelled.PaymentStatus)
	}

	// Unpaid card order -> no refund call.
	unpaid := place()
	if _, err := svc.CancelForCustomer(ctx, 100, unpaid.ID); err != nil {
		t.Fatalf("cancel unpaid: %v", err)
	}
	if refunder.calls != 1 {
		t.Errorf("refunds = %d, want still 1 (unpaid orders are not refunded)", refunder.calls)
	}

	// Refund failure aborts the cancel — money and state never diverge.
	failing := place()
	storedF, _ := orders.FindByID(ctx, failing.ID)
	_ = storedF.MarkPaid()
	_ = orders.UpdateStatus(ctx, storedF)
	refunder.err = errors.New("gateway down")
	if _, err := svc.CancelForCustomer(ctx, 100, failing.ID); err == nil {
		t.Fatal("cancel should fail when the refund fails")
	}
	after, _ := orders.FindByID(ctx, failing.ID)
	if after.Status == domain.OrderStatusCancelled {
		t.Error("order must stay uncancelled when the refund fails")
	}
}
