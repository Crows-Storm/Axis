package adapters

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/config/logger"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
	commwallet "github.com/Crows-Storm/Axis/common/domain/wallet"
	"github.com/Crows-Storm/Axis/common/server/store"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type currencyModel struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement:false"`
	Class        string `gorm:"column:class;type:varchar(16);not null"`
	Symbol       string `gorm:"column:symbol;type:varchar(32);not null;uniqueIndex"`
	Name         string `gorm:"column:name;type:varchar(64);not null"`
	Address      string `gorm:"column:address;type:varchar(128);not null;default:''"`
	Decimals     uint8  `gorm:"column:decimals;not null"`
	SmallestUnit string `gorm:"column:smallest_unit;type:varchar(16);not null"`
}

func (*currencyModel) TableName() string { return "sys_currency" }

func (m currencyModel) toDomain() commaccount.Currency {
	return commaccount.Currency{
		ID:           m.ID,
		Class:        commaccount.AssetClass(m.Class),
		Symbol:       m.Symbol,
		Name:         m.Name,
		Address:      m.Address,
		Decimals:     m.Decimals,
		SmallestUnit: m.SmallestUnit,
	}
}

type accountModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement:false"`
	UserID    uint64    `gorm:"column:user_id;not null;index"`
	Type      string    `gorm:"column:type;type:varchar(32);not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"`
}

func (*accountModel) TableName() string { return "sys_account" }

type walletModel struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement:false"`
	AccountID  uint64    `gorm:"column:account_id;not null;uniqueIndex:uk_account_currency"`
	CurrencyID uint64    `gorm:"column:currency_id;not null;uniqueIndex:uk_account_currency"`
	Balance    string    `gorm:"column:balance;type:varchar(80);not null;default:0"`
	Address    string    `gorm:"column:address;type:varchar(128);not null;default:''"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime"`
}

func (*walletModel) TableName() string { return "sys_wallet" }

var currencySeed = []currencyModel{
	{1, "fiat", "USD", "US Dollar", "", 2, "cent"},
	{2, "fiat", "CNY", "Chinese Yuan", "", 2, "分"},
	{3, "fiat", "EUR", "Euro", "", 2, "cent"},
	{4, "fiat", "JPY", "Japanese Yen", "", 0, "円"},
	{5, "fiat", "KRW", "Korean Won", "", 0, "원"},
	{6, "fiat", "BHD", "Bahraini Dinar", "", 3, "fils"},
	{7, "fiat", "KWD", "Kuwaiti Dinar", "", 3, "fils"},
	{8, "crypto", "BTC", "Bitcoin", "", 8, "satoshi"},
	{9, "crypto", "ETH", "Ether", "", 18, "wei"},
	{10, "crypto", "USDT", "Tether USD", "", 6, "micro"},
}

type AccountMariaRepository struct {
	store *store.Store
}

func NewAccountMariaRepository(st *store.Store) *AccountMariaRepository {
	repo := &AccountMariaRepository{store: st}
	if err := repo.store.DB().AutoMigrate(&currencyModel{}, &accountModel{}, &walletModel{}); err != nil {
		logger.WithError(err).Error("failed to migrate account tables")
	}
	for _, row := range currencySeed {
		if err := repo.store.DB().Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			logger.WithError(err).Error("failed to seed sys_currency")
		}
	}
	return repo
}

func (r *AccountMariaRepository) Get(ctx context.Context, id uint64) (*domain.Account, error) {
	var row accountModel
	err := r.store.DB().WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, commaccount.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get account %d: %w", id, err)
	}
	return r.hydrate(ctx, row)
}

func (r *AccountMariaRepository) GetByWalletID(ctx context.Context, walletID uint64) (*domain.Account, error) {
	var w walletModel
	err := r.store.DB().WithContext(ctx).Where("id = ?", walletID).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, commaccount.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get wallet %d: %w", walletID, err)
	}
	return r.Get(ctx, w.AccountID)
}

func (r *AccountMariaRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Account, error) {
	if filter.UserID == 0 {
		return nil, fmt.Errorf("list accounts: %w", commaccount.ErrInvalid)
	}
	query := r.store.DB().WithContext(ctx).Model(&accountModel{}).Where("user_id = ?", filter.UserID)
	if filter.Type != "" {
		query = query.Where("type = ?", string(filter.Type))
	}
	var rows []accountModel
	if err := query.Order("id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list accounts for user %d: %w", filter.UserID, err)
	}
	out := make([]*domain.Account, 0, len(rows))
	for _, row := range rows {
		acc, err := r.hydrate(ctx, row)
		if err != nil {
			return nil, err
		}
		out = append(out, acc)
	}
	return out, nil
}

func (r *AccountMariaRepository) Save(ctx context.Context, acc *domain.Account) error {
	if acc == nil || acc.ID == 0 {
		return fmt.Errorf("save account: %w", commaccount.ErrInvalid)
	}
	now := time.Now()
	return r.store.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := accountModel{ID: acc.ID, UserID: acc.UserID, Type: string(acc.Type), CreatedAt: now, UpdatedAt: now}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"type", "updated_at"}),
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("upsert account %d: %w", acc.ID, err)
		}
		for _, wallet := range acc.Wallet {
			balance := "0"
			if wallet.Balance != nil {
				balance = wallet.Balance.String()
			}
			wrow := walletModel{
				ID:         wallet.ID,
				AccountID:  acc.ID,
				CurrencyID: wallet.Currency.ID,
				Balance:    balance,
				Address:    wallet.Currency.Address,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"balance", "address", "updated_at"}),
			}).Create(&wrow).Error; err != nil {
				return fmt.Errorf("upsert wallet %d: %w", wallet.ID, err)
			}
		}
		return nil
	})
}

func (r *AccountMariaRepository) ListCurrencies(ctx context.Context) ([]commaccount.Currency, error) {
	var rows []currencyModel
	if err := r.store.DB().WithContext(ctx).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list currencies: %w", err)
	}
	out := make([]commaccount.Currency, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomain())
	}
	return out, nil
}

func (r *AccountMariaRepository) GetCurrency(ctx context.Context, id uint64, symbol string) (commaccount.Currency, error) {
	var row currencyModel
	query := r.store.DB().WithContext(ctx).Model(&currencyModel{})
	switch {
	case id != 0:
		query = query.Where("id = ?", id)
	case symbol != "":
		query = query.Where("symbol = ?", symbol)
	default:
		return commaccount.Currency{}, fmt.Errorf("get currency: %w", commaccount.ErrInvalid)
	}
	err := query.First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return commaccount.Currency{}, commaccount.ErrNotFound
	}
	if err != nil {
		return commaccount.Currency{}, fmt.Errorf("get currency: %w", err)
	}
	return row.toDomain(), nil
}

func (r *AccountMariaRepository) hydrate(ctx context.Context, row accountModel) (*domain.Account, error) {
	var wallets []walletModel
	if err := r.store.DB().WithContext(ctx).Where("account_id = ?", row.ID).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("list wallets for account %d: %w", row.ID, err)
	}
	mapped := make(map[uint64]domain.Wallet, len(wallets))
	for _, w := range wallets {
		currency, err := r.GetCurrency(ctx, w.CurrencyID, "")
		if err != nil {
			return nil, err
		}
		if w.Address != "" {
			currency.Address = w.Address
		}
		bal := new(big.Int)
		if _, ok := bal.SetString(w.Balance, 10); !ok {
			bal = big.NewInt(0)
		}
		mapped[w.ID] = domain.Wallet{ID: w.ID, Currency: currency, Balance: bal}
	}
	return domain.Reconstruct(row.ID, row.UserID, commwallet.Type(row.Type), mapped), nil
}
