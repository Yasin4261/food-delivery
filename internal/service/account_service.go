package service

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// maxExportRows caps how many of each collection the data export pulls — a
// generous ceiling for a personal account that keeps the dump bounded.
const maxExportRows = 1000

// AccountService implements the account data-rights use cases (#107): exporting
// everything the platform holds about the caller, and deleting (anonymising)
// their account. It depends only on domain ports.
type AccountService struct {
	users     domain.UserRepository
	chefs     domain.ChefRepository
	addresses domain.AddressRepository
	orders    domain.OrderRepository
	reviews   domain.ReviewRepository
	chats     domain.ChatRepository
	account   domain.AccountRepository
}

// NewAccountService builds an AccountService.
func NewAccountService(
	users domain.UserRepository,
	chefs domain.ChefRepository,
	addresses domain.AddressRepository,
	orders domain.OrderRepository,
	reviews domain.ReviewRepository,
	chats domain.ChatRepository,
	account domain.AccountRepository,
) *AccountService {
	return &AccountService{
		users:     users,
		chefs:     chefs,
		addresses: addresses,
		orders:    orders,
		reviews:   reviews,
		chats:     chats,
		account:   account,
	}
}

// Export assembles the caller's own data into a single portable document. It
// only ever reads records owned by, or that the caller took part in — never
// another user's personal data.
func (s *AccountService) Export(ctx context.Context, userID int) (*domain.AccountExport, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""

	export := &domain.AccountExport{
		ExportedAt:    time.Now().UTC(),
		User:          user,
		Addresses:     []*domain.Address{},
		Orders:        []*domain.Order{},
		Reviews:       []*domain.Review{},
		Conversations: []*domain.ConversationExport{},
	}

	// Chef storefront (optional).
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err == nil {
		export.Chef = chef
	} else if err != domain.ErrChefNotFound {
		return nil, err
	}

	if addrs, err := s.addresses.ListByUser(ctx, userID); err != nil {
		return nil, err
	} else if addrs != nil {
		export.Addresses = addrs
	}

	orders, _, err := s.orders.ListByUser(ctx, userID, maxExportRows, 0)
	if err != nil {
		return nil, err
	}
	if orders != nil {
		export.Orders = orders
	}

	reviews, err := s.reviews.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if reviews != nil {
		export.Reviews = reviews
	}

	// Chat threads the caller took part in — as a customer, and (if they run a
	// kitchen) as a chef. Each thread carries the messages both parties sent,
	// since it is a shared conversation the caller was part of.
	convs, err := s.chats.ListConversationsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if export.Chef != nil {
		chefConvs, err := s.chats.ListConversationsByChef(ctx, export.Chef.ID)
		if err != nil {
			return nil, err
		}
		convs = append(convs, chefConvs...)
	}
	for _, c := range convs {
		msgs, _, err := s.chats.ListMessages(ctx, c.ID, maxExportRows, 0)
		if err != nil {
			return nil, err
		}
		if msgs == nil {
			msgs = []*domain.Message{}
		}
		export.Conversations = append(export.Conversations, &domain.ConversationExport{
			Conversation: c,
			Messages:     msgs,
		})
	}

	return export, nil
}

// Delete anonymises the caller's account after verifying their password. The
// password check means a stolen session alone cannot erase the account. The
// operation is irreversible; login is blocked afterwards.
func (s *AccountService) Delete(ctx context.Context, userID int, password string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return domain.ErrInvalidCredentials
	}
	return s.account.Anonymise(ctx, userID)
}
