package repository

import (
	"context"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"
)

type SearchRepo struct {
	db *database.DB
}

func NewSearchRepo(db *database.DB) *SearchRepo {
	return &SearchRepo{db: db}
}

func (r *SearchRepo) Search(ctx context.Context, q string) (*domain.SearchResults, error) {
	pattern := "%" + q + "%"
	results := &domain.SearchResults{}

	tenants, err := r.searchTenants(ctx, pattern)
	if err != nil {
		return nil, err
	}
	results.Tenants = tenants

	rooms, err := r.searchRooms(ctx, pattern)
	if err != nil {
		return nil, err
	}
	results.Rooms = rooms

	bills, err := r.searchBills(ctx, pattern)
	if err != nil {
		return nil, err
	}
	results.Bills = bills

	contracts, err := r.searchContracts(ctx, pattern)
	if err != nil {
		return nil, err
	}
	results.Contracts = contracts

	return results, nil
}

func (r *SearchRepo) searchTenants(ctx context.Context, pattern string) ([]*domain.Tenant, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, first_name, last_name, phone, id_card, email, emergency_contact, note, created_at, updated_at
		FROM tenants
		WHERE (first_name || ' ' || last_name) ILIKE $1
		   OR first_name ILIKE $1
		   OR last_name ILIKE $1
		   OR phone ILIKE $1
		   OR id_card ILIKE $1
		ORDER BY first_name, last_name
		LIMIT 5`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Tenant
	for rows.Next() {
		t := &domain.Tenant{}
		if err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Phone, &t.IDCard,
			&t.Email, &t.EmergencyContact, &t.Note, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	if list == nil {
		list = []*domain.Tenant{}
	}
	return list, rows.Err()
}

func (r *SearchRepo) searchRooms(ctx context.Context, pattern string) ([]*domain.Room, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, room_number, floor, type, status, rent_price, description, created_at, updated_at
		FROM rooms
		WHERE room_number ILIKE $1
		ORDER BY floor, room_number
		LIMIT 5`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Room
	for rows.Next() {
		ro := &domain.Room{}
		if err := rows.Scan(&ro.ID, &ro.RoomNumber, &ro.Floor, &ro.Type, &ro.Status,
			&ro.RentPrice, &ro.Description, &ro.CreatedAt, &ro.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, ro)
	}
	if list == nil {
		list = []*domain.Room{}
	}
	return list, rows.Err()
}

func (r *SearchRepo) searchBills(ctx context.Context, pattern string) ([]*domain.SearchBill, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT b.id, b.billing_month, b.total_amount, b.status,
		       t.first_name, t.last_name, r.room_number
		FROM bills b
		JOIN contracts c ON c.id = b.contract_id
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms r ON r.id = c.room_id
		WHERE (t.first_name || ' ' || t.last_name) ILIKE $1
		   OR t.first_name ILIKE $1
		   OR t.last_name ILIKE $1
		   OR r.room_number ILIKE $1
		ORDER BY b.billing_month DESC
		LIMIT 5`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.SearchBill
	for rows.Next() {
		b := &domain.SearchBill{}
		if err := rows.Scan(&b.ID, &b.BillingMonth, &b.TotalAmount, &b.Status,
			&b.TenantFirstName, &b.TenantLastName, &b.RoomNumber); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	if list == nil {
		list = []*domain.SearchBill{}
	}
	return list, rows.Err()
}

func (r *SearchRepo) searchContracts(ctx context.Context, pattern string) ([]*domain.SearchContract, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT c.id, c.start_date, c.end_date, c.status,
		       t.first_name, t.last_name, r.room_number
		FROM contracts c
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms r ON r.id = c.room_id
		WHERE (t.first_name || ' ' || t.last_name) ILIKE $1
		   OR t.first_name ILIKE $1
		   OR t.last_name ILIKE $1
		   OR r.room_number ILIKE $1
		ORDER BY c.created_at DESC
		LIMIT 5`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.SearchContract
	for rows.Next() {
		con := &domain.SearchContract{}
		if err := rows.Scan(&con.ID, &con.StartDate, &con.EndDate, &con.Status,
			&con.TenantFirstName, &con.TenantLastName, &con.RoomNumber); err != nil {
			return nil, err
		}
		list = append(list, con)
	}
	if list == nil {
		list = []*domain.SearchContract{}
	}
	return list, rows.Err()
}
