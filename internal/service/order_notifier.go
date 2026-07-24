package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// OrderNotifier composes and sends transactional order emails through the
// domain.Mailer port: "new order" to each participating chef on placement,
// "status changed" to the customer on meaningful sub-order transitions.
//
// Sends are fire-and-forget: dispatched on a goroutine with a cancel-free
// context so mail latency never blocks the order flow; failures are logged
// and never surfaced to the buyer.
type OrderNotifier struct {
	mailer domain.Mailer
	users  domain.UserRepository
	chefs  domain.ChefRepository
	// currency is the ISO-4217 code amounts are quoted in (config). Emails are
	// the record people keep, so they must never state the wrong currency.
	currency string
}

// NewOrderNotifier builds an OrderNotifier. currency is the ISO-4217 code from
// config; an empty value falls back to TRY.
func NewOrderNotifier(mailer domain.Mailer, users domain.UserRepository, chefs domain.ChefRepository, currency string) *OrderNotifier {
	if currency == "" {
		currency = "TRY"
	}
	return &OrderNotifier{mailer: mailer, users: users, chefs: chefs, currency: currency}
}

// money renders an amount with its currency code — deliberately unambiguous
// ("45.00 TRY") rather than a locale-guessed symbol (#125).
func (n *OrderNotifier) money(amount float64) string {
	return fmt.Sprintf("%.2f %s", amount, n.currency)
}

// OrderPlaced emails every chef with a slice of the new order (their items
// only). Called by the order service after the order is persisted.
func (n *OrderNotifier) OrderPlaced(ctx context.Context, order *domain.Order) {
	ctx = context.WithoutCancel(ctx)
	for _, sub := range order.SubOrders {
		sub := sub
		go func() {
			chef, err := n.chefs.FindByID(ctx, sub.ChefID)
			if err != nil {
				n.logFailure("new-order", order, err)
				return
			}
			user, err := n.users.FindByID(ctx, chef.UserID)
			if err != nil {
				n.logFailure("new-order", order, err)
				return
			}
			if !user.EmailNotifications {
				return // the chef opted out of order emails (#71)
			}
			msg := domain.Email{
				To:      user.Email,
				Subject: fmt.Sprintf("New order %s", order.OrderCode),
				Body:    n.newOrderBody(order, sub, chef.BusinessName),
			}
			if err := n.mailer.Send(ctx, msg); err != nil {
				n.logFailure("new-order", order, err)
			}
		}()
	}
}

// Sub-order statuses worth a customer email. Preparing/ready are internal
// kitchen steps — mailing every one of them is spam.
var notifiableStatuses = map[string]string{
	domain.OrderStatusConfirmed:  "accepted your order",
	domain.OrderStatusDelivering: "is delivering your order",
	domain.OrderStatusDelivered:  "delivered your order",
	domain.OrderStatusCancelled:  "declined your order",
}

// SubOrderAdvanced emails the customer when a chef moves their slice through a
// meaningful transition (confirmed, delivering, delivered, declined). Called
// by the order service after the transition is persisted.
func (n *OrderNotifier) SubOrderAdvanced(ctx context.Context, order *domain.Order, sub *domain.SubOrder) {
	verb, ok := notifiableStatuses[sub.Status]
	if !ok {
		return
	}
	ctx = context.WithoutCancel(ctx)
	go func() {
		user, err := n.users.FindByID(ctx, order.UserID)
		if err != nil {
			n.logFailure("status-change", order, err)
			return
		}
		if !user.EmailNotifications {
			return // the customer opted out of order emails (#71)
		}
		chefName := sub.ChefName
		if chefName == "" {
			chefName = "The chef"
		}
		msg := domain.Email{
			To:      user.Email,
			Subject: fmt.Sprintf("Order %s: %s %s", order.OrderCode, chefName, verb),
			Body:    n.statusChangeBody(order, sub, chefName, verb),
		}
		if err := n.mailer.Send(ctx, msg); err != nil {
			n.logFailure("status-change", order, err)
		}
	}()
}

func (n *OrderNotifier) logFailure(kind string, order *domain.Order, err error) {
	slog.Error("order email failed", "kind", kind, "order_id", order.ID, "error", err)
}

// newOrderBody renders the chef-facing "new order" email: the chef's own
// items and slice subtotal, never other chefs' lines.
func (n *OrderNotifier) newOrderBody(order *domain.Order, sub *domain.SubOrder, businessName string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Hi %s,\n\nYou have a new order (%s):\n\n", businessName, order.OrderCode)
	for _, it := range order.Items {
		if it.ChefID != sub.ChefID {
			continue
		}
		fmt.Fprintf(&b, "  %d x %s — %s\n", it.Quantity, it.ItemName, n.money(it.Subtotal))
	}
	fmt.Fprintf(&b, "\nYour subtotal: %s\n", n.money(sub.Subtotal))
	fmt.Fprintf(&b, "Delivery address: %s\n", order.DeliveryAddress)
	if order.CustomerNotes != nil && *order.CustomerNotes != "" {
		fmt.Fprintf(&b, "Customer notes: %s\n", *order.CustomerNotes)
	}
	b.WriteString("\nOpen your dashboard to accept or decline it.\n")
	return b.String()
}

// statusChangeBody renders the customer-facing "status changed" email.
func (n *OrderNotifier) statusChangeBody(order *domain.Order, sub *domain.SubOrder, chefName, verb string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Hi,\n\n%s %s (%s).\n", chefName, verb, order.OrderCode)
	// A declined slice of a card payment is refunded (the order shows paid
	// while other slices stay alive, refunded once every slice is declined).
	cardCharged := order.PaymentMethod != nil && *order.PaymentMethod == domain.PaymentMethodCard &&
		(order.PaymentStatus == domain.PaymentStatusPaid || order.PaymentStatus == domain.PaymentStatusRefunded)
	if sub.Status == domain.OrderStatusCancelled && cardCharged {
		fmt.Fprintf(&b, "The %s you paid for this part of the order is being refunded.\n", n.money(sub.Subtotal))
	}
	b.WriteString("\nYou can follow your order on the My orders page.\n")
	return b.String()
}
