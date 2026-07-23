package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AdminService implements the admin/moderation use cases. Access is gated to
// the admin role at the router; this layer clears secrets and normalises
// paging.
type AdminService struct {
	admin domain.AdminRepository
}

// NewAdminService builds an AdminService. All admin reads, writes and the audit
// trail go through the one AdminRepository so a mutation and its audit row
// share a transaction.
func NewAdminService(admin domain.AdminRepository) *AdminService {
	return &AdminService{admin: admin}
}

// PromoInput is the data needed to create a promo code.
type PromoInput struct {
	Code          string
	DiscountType  string
	DiscountValue float64
	MinOrder      float64
	ValidFrom     *time.Time
	ValidUntil    *time.Time
	UsageLimit    int
}

// CreatePromo validates and persists a new promo code.
func (s *AdminService) CreatePromo(ctx context.Context, actorID int, in PromoInput) (*domain.PromoCode, error) {
	p := &domain.PromoCode{
		Code:          in.Code,
		DiscountType:  in.DiscountType,
		DiscountValue: in.DiscountValue,
		MinOrder:      in.MinOrder,
		ValidFrom:     in.ValidFrom,
		ValidUntil:    in.ValidUntil,
		UsageLimit:    in.UsageLimit,
		IsActive:      true,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.admin.CreatePromo(ctx, audit(actorID, domain.AuditPromoCreate, domain.AuditTargetPromo, 0, ""), p); err != nil {
		return nil, err
	}
	return p, nil
}

// UpdatePromo edits a promo code's definition (never its usage counter).
func (s *AdminService) UpdatePromo(ctx context.Context, actorID, id int, in PromoInput) (*domain.PromoCode, error) {
	existing, err := s.admin.FindPromo(ctx, id)
	if err != nil {
		return nil, err
	}
	// Edit the definition; preserve identity + usage.
	existing.Code = domain.NormaliseCode(in.Code)
	existing.DiscountType = in.DiscountType
	existing.DiscountValue = in.DiscountValue
	existing.MinOrder = in.MinOrder
	existing.ValidFrom = in.ValidFrom
	existing.ValidUntil = in.ValidUntil
	existing.UsageLimit = in.UsageLimit
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	if err := s.admin.UpdatePromo(ctx, audit(actorID, domain.AuditPromoUpdate, domain.AuditTargetPromo, id, ""), existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// DeletePromo removes a promo code.
func (s *AdminService) DeletePromo(ctx context.Context, actorID, id int) error {
	return s.admin.DeletePromo(ctx, audit(actorID, domain.AuditPromoDelete, domain.AuditTargetPromo, id, ""), id)
}

// ListPromos returns a page of all promo codes and the total.
func (s *AdminService) ListPromos(ctx context.Context, limit, offset int) ([]*domain.PromoCode, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.admin.ListPromos(ctx, limit, offset)
}

// SetPromoActive activates or deactivates a promo code.
func (s *AdminService) SetPromoActive(ctx context.Context, actorID, id int, active bool) error {
	return s.admin.SetPromoActive(ctx, audit(actorID, domain.AuditPromoSetActive, domain.AuditTargetPromo, id, ""), id, active)
}

// ListUsers returns a page of users matching f (password hashes cleared) and
// the total. Unknown role values are rejected rather than silently ignored, so
// a typo'd filter never quietly returns "everything".
func (s *AdminService) ListUsers(ctx context.Context, f domain.AdminUserFilters, limit, offset int) ([]*domain.User, int, error) {
	if f.Role != "" && !domain.ValidRole(f.Role) {
		return nil, 0, ValidationError{Msg: "unknown role filter"}
	}
	limit, offset = normalisePaging(limit, offset)
	users, total, err := s.admin.ListUsers(ctx, f, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range users {
		u.PasswordHash = ""
	}
	return users, total, nil
}

// SetUserActive activates or deactivates a user (deactivation blocks login).
// Deactivation is destructive, so it requires a reason for the audit trail.
func (s *AdminService) SetUserActive(ctx context.Context, actorID, userID int, active bool, reason string) error {
	if err := requireReasonForDeactivation(active, reason); err != nil {
		return err
	}
	return s.admin.SetUserActive(ctx, audit(actorID, domain.AuditUserSetActive, domain.AuditTargetUser, userID, reason), userID, active)
}

// ListChefs returns a page of chefs (including inactive) matching f, plus the
// total matching count.
func (s *AdminService) ListChefs(ctx context.Context, f domain.AdminChefFilters, limit, offset int) ([]*domain.Chef, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.admin.ListChefs(ctx, f, limit, offset)
}

// SetChefActive activates or deactivates a chef (deactivation hides them from
// browse/search and blocks new orders). Deactivation requires a reason.
func (s *AdminService) SetChefActive(ctx context.Context, actorID, chefID int, active bool, reason string) error {
	if err := requireReasonForDeactivation(active, reason); err != nil {
		return err
	}
	return s.admin.SetChefActive(ctx, audit(actorID, domain.AuditChefSetActive, domain.AuditTargetChef, chefID, reason), chefID, active)
}

// SetChefOnline drives a chef's live presence on their behalf (admin support).
func (s *AdminService) SetChefOnline(ctx context.Context, actorID, chefID int, online bool, reason string) error {
	return s.admin.SetChefOnline(ctx, audit(actorID, domain.AuditChefSetOnline, domain.AuditTargetChef, chefID, reason), chefID, online)
}

// SetChefAcceptingOrders drives a chef's availability on their behalf.
func (s *AdminService) SetChefAcceptingOrders(ctx context.Context, actorID, chefID int, accepting bool, reason string) error {
	return s.admin.SetChefAcceptingOrders(ctx, audit(actorID, domain.AuditChefSetAccepting, domain.AuditTargetChef, chefID, reason), chefID, accepting)
}

// ListAudit returns a page of the admin audit log matching f, newest first.
func (s *AdminService) ListAudit(ctx context.Context, f domain.AuditFilters, limit, offset int) ([]*domain.AuditEntry, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.admin.ListAudit(ctx, f, limit, offset)
}

// audit builds an audit entry for a mutation. Before/After are filled by the
// repository inside the mutation's transaction.
func audit(actorID int, action, targetType string, targetID int, reason string) *domain.AuditEntry {
	return &domain.AuditEntry{
		ActorUserID: actorID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Reason:      reason,
	}
}

// requireReasonForDeactivation enforces a reason on destructive toggles: turning
// something off (active == false) must be justified in the audit trail.
func requireReasonForDeactivation(active bool, reason string) error {
	if !active && strings.TrimSpace(reason) == "" {
		return ValidationError{Msg: "a reason is required to deactivate"}
	}
	return nil
}

// ListOrders returns a page of orders matching f (the platform/support
// overview), plus the total matching count. Unknown lifecycle values are
// rejected rather than silently ignored, so a typo'd filter never quietly
// returns "everything".
func (s *AdminService) ListOrders(ctx context.Context, f domain.AdminOrderFilters, limit, offset int) ([]*domain.Order, int, error) {
	if f.Status != "" && !domain.ValidOrderStatus(f.Status) {
		return nil, 0, ValidationError{Msg: "unknown order status filter"}
	}
	if f.PaymentStatus != "" && !domain.ValidPaymentStatus(f.PaymentStatus) {
		return nil, 0, ValidationError{Msg: "unknown payment status filter"}
	}
	limit, offset = normalisePaging(limit, offset)
	return s.admin.ListOrders(ctx, f, limit, offset)
}

// UserDetail returns one account's support context (kitchen, recent orders,
// reviews written). The password hash is always cleared before it leaves the
// service, exactly as in the listings.
func (s *AdminService) UserDetail(ctx context.Context, userID int) (*domain.AdminUserDetail, error) {
	d, err := s.admin.UserDetail(ctx, userID)
	if err != nil {
		return nil, err
	}
	if d.User != nil {
		d.User.PasswordHash = ""
	}
	return d, nil
}

// OrderDetail returns one order's support context (items, sub-orders, customer,
// payment attempts).
func (s *AdminService) OrderDetail(ctx context.Context, orderID int) (*domain.AdminOrderDetail, error) {
	d, err := s.admin.OrderDetail(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if d.Customer != nil {
		d.Customer.PasswordHash = ""
	}
	return d, nil
}

// ChefDetail returns one kitchen's support context (owner, dishes, orders).
func (s *AdminService) ChefDetail(ctx context.Context, chefID int) (*domain.AdminChefDetail, error) {
	d, err := s.admin.ChefDetail(ctx, chefID)
	if err != nil {
		return nil, err
	}
	if d.Owner != nil {
		d.Owner.PasswordHash = ""
	}
	return d, nil
}

// Stats returns the aggregated platform dashboard figures.
func (s *AdminService) Stats(ctx context.Context) (*domain.PlatformStats, error) {
	return s.admin.Stats(ctx)
}
