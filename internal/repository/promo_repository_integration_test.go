//go:build integration

package repository_test

import (
	"sync"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestPromoRepository_CRUDAndFind(t *testing.T) {
	resetDB(t)
	repo := repository.NewPromoRepository(testDB)

	p := &domain.PromoCode{Code: "SAVE10", DiscountType: domain.PromoPercent, DiscountValue: 10, MinOrder: 50, UsageLimit: 3, IsActive: true}
	if err := repo.Create(ctx(), p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.ID == 0 {
		t.Fatal("id not back-filled")
	}
	// Duplicate code -> ErrPromoExists.
	if err := repo.Create(ctx(), &domain.PromoCode{Code: "SAVE10", DiscountType: domain.PromoFixed, DiscountValue: 5}); err != domain.ErrPromoExists {
		t.Errorf("duplicate = %v, want ErrPromoExists", err)
	}

	// Case-insensitive lookup.
	got, err := repo.FindByCode(ctx(), "save10")
	if err != nil || got.DiscountValue != 10 || got.UsageLimit != 3 {
		t.Fatalf("find = %+v (%v)", got, err)
	}
	if _, err := repo.FindByCode(ctx(), "ghost"); err != domain.ErrPromoNotFound {
		t.Errorf("unknown = %v, want ErrPromoNotFound", err)
	}

	// List + deactivate.
	list, total, _ := repo.List(ctx(), 20, 0)
	if total != 1 || len(list) != 1 {
		t.Errorf("list = %d/%d, want 1", len(list), total)
	}
	if err := repo.SetActive(ctx(), p.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	got, _ = repo.FindByCode(ctx(), "SAVE10")
	if got.IsActive {
		t.Error("promo still active after deactivate")
	}
}

// The guarded UPDATE must enforce the usage cap even under concurrency: with a
// limit of 5, exactly 5 of 20 racing redeems succeed.
func TestPromoRepository_RedeemAtomicCap(t *testing.T) {
	resetDB(t)
	repo := repository.NewPromoRepository(testDB)
	p := &domain.PromoCode{Code: "LIMITED", DiscountType: domain.PromoFixed, DiscountValue: 5, UsageLimit: 5, IsActive: true}
	if err := repo.Create(ctx(), p); err != nil {
		t.Fatalf("create: %v", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, used := 0, 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.Redeem(ctx(), p.ID)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				ok++
			} else if err == domain.ErrPromoUsedUp {
				used++
			} else {
				t.Errorf("unexpected redeem error: %v", err)
			}
		}()
	}
	wg.Wait()

	if ok != 5 || used != 15 {
		t.Errorf("redeems = %d ok / %d used-up, want 5/15", ok, used)
	}
	got, _ := repo.FindByCode(ctx(), "LIMITED")
	if got.UsedCount != 5 {
		t.Errorf("used_count = %d, want 5", got.UsedCount)
	}
}

// An unlimited code (usage_limit 0) always redeems.
func TestPromoRepository_RedeemUnlimited(t *testing.T) {
	resetDB(t)
	repo := repository.NewPromoRepository(testDB)
	p := &domain.PromoCode{Code: "OPEN", DiscountType: domain.PromoPercent, DiscountValue: 10, UsageLimit: 0, IsActive: true}
	_ = repo.Create(ctx(), p)
	for i := 0; i < 10; i++ {
		if err := repo.Redeem(ctx(), p.ID); err != nil {
			t.Fatalf("redeem %d: %v", i, err)
		}
	}
}
