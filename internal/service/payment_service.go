package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// PaymentService implements the card-payment use cases over the
// domain.PaymentGateway port: opening a hosted checkout for an order,
// completing it from the gateway callback, and refunding on cancel. It also
// implements domain.PaymentRefunder for the order service.
type PaymentService struct {
	sessions   domain.PaymentSessionRepository
	orders     domain.OrderRepository
	users      domain.UserRepository
	gateway    domain.PaymentGateway
	cards      domain.PaymentMethodRepository // saved cards (#67); nil disables the feature
	appBaseURL string
}

// NewPaymentService builds a PaymentService. appBaseURL is the public origin
// the gateway calls back to (the SPA origin; /api is proxied to the API).
func NewPaymentService(
	sessions domain.PaymentSessionRepository,
	orders domain.OrderRepository,
	users domain.UserRepository,
	gateway domain.PaymentGateway,
	appBaseURL string,
) *PaymentService {
	return &PaymentService{
		sessions:   sessions,
		orders:     orders,
		users:      users,
		gateway:    gateway,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
	}
}

// SetPaymentMethods enables the saved-card feature (#67) by wiring the
// payment-method repository. Left unset, checkouts store nothing and the
// saved-card endpoints report an empty list.
func (s *PaymentService) SetPaymentMethods(repo domain.PaymentMethodRepository) { s.cards = repo }

// StartCheckout opens a hosted checkout for the caller's pending card order and
// returns the page URL to send the browser to. registerCard opts the customer
// in to saving the card used this time (#67); the customer's existing wallet
// key (if any) is passed so their saved cards appear in the hosted form.
func (s *PaymentService) StartCheckout(ctx context.Context, userID, orderID int, registerCard bool) (string, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order.UserID != userID {
		return "", domain.ErrForbidden
	}
	if order.Status == domain.OrderStatusCancelled ||
		order.PaymentMethod == nil || *order.PaymentMethod != domain.PaymentMethodCard ||
		order.PaymentStatus != domain.PaymentStatusPending {
		return "", domain.ErrOrderNotPayable
	}

	buyer, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	opts := domain.CheckoutOptions{RegisterCard: registerCard}
	if s.cards != nil {
		if key, err := s.cards.CardUserKey(ctx, userID); err == nil {
			opts.CardUserKey = key
		} else {
			return "", err
		}
	}

	callbackURL := s.appBaseURL + "/api/v2/payments/callback"
	cs, err := s.gateway.InitiateCheckout(ctx, order, buyer, callbackURL, opts)
	if err != nil {
		return "", fmt.Errorf("initiate checkout: %w", err)
	}

	session := &domain.PaymentSession{OrderID: order.ID, Token: cs.Token, Status: domain.PaymentSessionInitiated}
	if err := s.sessions.Create(ctx, session); err != nil {
		return "", err
	}
	return cs.PaymentPageURL, nil
}

// CompleteCheckout resolves a gateway callback token: it verifies the outcome
// server-to-server (the browser-supplied token is never trusted alone), marks
// the order paid on success, and records the session state. It is idempotent
// for replayed callbacks.
func (s *PaymentService) CompleteCheckout(ctx context.Context, token string) (orderID int, paid bool, err error) {
	result, err := s.gateway.VerifyCheckout(ctx, token)
	if err != nil {
		return 0, false, fmt.Errorf("verify checkout: %w", err)
	}
	session, err := s.sessions.FindByToken(ctx, result.Token)
	if err != nil {
		return 0, false, err
	}

	// Replayed callback for an already-settled session.
	if session.Status == domain.PaymentSessionPaid {
		return session.OrderID, true, nil
	}

	if !result.Paid {
		if err := s.sessions.UpdateStatus(ctx, session.ID, domain.PaymentSessionFailed, nil); err != nil {
			return 0, false, err
		}
		return session.OrderID, false, nil
	}

	order, err := s.orders.FindByID(ctx, session.OrderID)
	if err != nil {
		return 0, false, err
	}
	if err := order.MarkPaid(); err == nil {
		if err := s.orders.UpdateStatus(ctx, order); err != nil {
			return 0, false, err
		}
	}
	if err := s.sessions.UpdateStatus(ctx, session.ID, domain.PaymentSessionPaid, &result.PaymentID); err != nil {
		return 0, false, err
	}

	// The customer opted to save their card: persist the association. Best
	// effort — a storage hiccup must never fail an already-captured payment.
	if s.cards != nil && result.RegisteredCard != nil {
		rc := result.RegisteredCard
		_ = s.cards.Add(ctx, &domain.SavedCard{
			UserID:       order.UserID,
			CardUserKey:  rc.CardUserKey,
			CardToken:    rc.CardToken,
			MaskedNumber: rc.MaskedNumber,
			Association:  rc.Association,
			Family:       rc.Family,
			BankName:     rc.BankName,
		})
	}
	return session.OrderID, true, nil
}

// RefundOrderPayment returns a captured card payment (domain.PaymentRefunder,
// called by the order service when a paid order is cancelled).
func (s *PaymentService) RefundOrderPayment(ctx context.Context, order *domain.Order) error {
	session, err := s.sessions.FindPaidByOrder(ctx, order.ID)
	if err != nil {
		return err
	}
	paymentID := ""
	if session.PaymentID != nil {
		paymentID = *session.PaymentID
	}
	if err := s.gateway.Refund(ctx, paymentID); err != nil {
		return fmt.Errorf("refund payment: %w", err)
	}
	return s.sessions.UpdateStatus(ctx, session.ID, domain.PaymentSessionRefunded, nil)
}

// RefundSubOrderPayment returns one declined sub-order's slice of a captured
// card payment (domain.PaymentRefunder). The session stays paid — the other
// sub-orders' money remains captured.
func (s *PaymentService) RefundSubOrderPayment(ctx context.Context, order *domain.Order, amount float64) error {
	session, err := s.sessions.FindPaidByOrder(ctx, order.ID)
	if err != nil {
		return err
	}
	paymentID := ""
	if session.PaymentID != nil {
		paymentID = *session.PaymentID
	}
	if err := s.gateway.RefundPartial(ctx, paymentID, amount); err != nil {
		return fmt.Errorf("refund sub-order payment: %w", err)
	}
	return nil
}

// ListSavedCards returns the caller's saved cards (#67), newest first. With the
// feature disabled (no repository wired) it reports an empty list.
func (s *PaymentService) ListSavedCards(ctx context.Context, userID int) ([]*domain.SavedCard, error) {
	if s.cards == nil {
		return []*domain.SavedCard{}, nil
	}
	return s.cards.ListByUser(ctx, userID)
}

// DeleteSavedCard removes one of the caller's saved cards: it revokes the card
// at the gateway, then deletes the local association. Owner-scoped — deleting a
// token that is not the caller's returns ErrCardNotFound.
func (s *PaymentService) DeleteSavedCard(ctx context.Context, userID int, cardToken string) error {
	if s.cards == nil {
		return domain.ErrCardNotFound
	}
	card, err := s.cards.FindByToken(ctx, userID, cardToken)
	if err != nil {
		return err
	}
	if err := s.gateway.DeleteStoredCard(ctx, card.CardUserKey, card.CardToken); err != nil {
		return fmt.Errorf("revoke stored card: %w", err)
	}
	return s.cards.Delete(ctx, userID, cardToken)
}

// ResultRedirectURL is where the callback handler sends the browser after a
// checkout completes.
func (s *PaymentService) ResultRedirectURL(orderID int, paid bool) string {
	outcome := "failed"
	if paid {
		outcome = "success"
	}
	return fmt.Sprintf("%s/orders?payment=%s&order=%d", s.appBaseURL, outcome, orderID)
}

// ErrorRedirectURL is where the browser goes when a callback cannot be
// resolved at all (unknown token, gateway error).
func (s *PaymentService) ErrorRedirectURL() string {
	return s.appBaseURL + "/orders?payment=error"
}
