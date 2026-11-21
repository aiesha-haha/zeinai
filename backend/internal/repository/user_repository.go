package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	// Initialize free plan limits
	limits := domain.GetPlanLimits(domain.PlanFree)
	user.SubscriptionPlan = domain.PlanFree
	user.TokensUsed = 0
	user.TokensLimit = limits.TokensLimit
	user.StudentsCount = 0
	user.StudentsLimit = limits.StudentsLimit
	user.SubscriptionStartedAt = time.Now()

	query := `
		INSERT INTO users (
			phone, email, iin, password_hash, first_name, last_name, middle_name,
			birth_year, gender, role, school_id, google_id,
			subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at,
			created_at, updated_at, avatar_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id
	`

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err := r.db.QueryRow(ctx, query,
		user.Phone,
		user.Email,
		user.IIN,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.BirthYear,
		user.Gender,
		user.Role,
		user.SchoolID,
		user.GoogleID,
		user.SubscriptionPlan,
		user.TokensUsed,
		user.TokensLimit,
		user.StudentsCount,
		user.StudentsLimit,
		user.SubscriptionStartedAt,
		user.CreatedAt,
		user.UpdatedAt,
		user.AvatarURL,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE phone = $1 AND deleted_at IS NULL`
	return r.scanUser(ctx, query, phone)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL`
	return r.scanUser(ctx, query, email)
}

func (r *UserRepo) GetByIIN(ctx context.Context, iin string) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE iin = $1 AND deleted_at IS NULL`
	return r.scanUser(ctx, query, iin)
}

func (r *UserRepo) GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE google_id = $1 AND deleted_at IS NULL`
	return r.scanUser(ctx, query, googleID)
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL`
	return r.scanUser(ctx, query, id)
}

func (r *UserRepo) scanUser(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Phone,
		&user.Email,
		&user.IIN,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.BirthYear,
		&user.Gender,
		&user.Role,
		&user.SchoolID,
		&user.GoogleID,
		&user.EmailVerifiedAt,
		&user.AvatarURL,
		&user.SubscriptionPlan,
		&user.TokensUsed,
		&user.TokensLimit,
		&user.StudentsCount,
		&user.StudentsLimit,
		&user.SubscriptionStartedAt,
		&user.SubscriptionExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil if not found, let service handle 404
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateByID(ctx context.Context, id uuid.UUID, user *domain.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, middle_name = $3, birth_year = $4, gender = $5, school_id = $6, phone = $7, email = $8, updated_at = NOW()
		WHERE id = $9 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.BirthYear,
		user.Gender,
		user.SchoolID,
		user.Phone,
		user.Email,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *UserRepo) UpdateEmailVerified(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}
	return nil
}

func (r *UserRepo) IncrementStudentsCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET students_count = students_count + 1, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment students count: %w", err)
	}
	return nil
}

func (r *UserRepo) DecrementStudentsCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET students_count = GREATEST(0, students_count - 1), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to decrement students count: %w", err)
	}
	return nil
}

func (r *UserRepo) DecreaseStudentsCount(ctx context.Context, id uuid.UUID, amount int) error {
	log.Printf("[DEBUG] Decreasing students_count for user %s by %d\n", id, amount)
	query := `
		UPDATE users
		SET students_count = GREATEST(0, students_count - $1), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to decrease students count: %w", err)
	}
	return nil
}

func (r *UserRepo) ListByRole(ctx context.Context, role domain.Role, limit, offset int) ([]domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at FROM users WHERE role = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, role, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by role: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(
			&u.ID, &u.Phone, &u.Email, &u.IIN, &u.PasswordHash, &u.FirstName, &u.LastName, &u.MiddleName,
			&u.BirthYear, &u.Gender, &u.Role, &u.SchoolID, &u.GoogleID, &u.EmailVerifiedAt, &u.AvatarURL,
			&u.SubscriptionPlan, &u.TokensUsed, &u.TokensLimit, &u.StudentsCount, &u.StudentsLimit,
			&u.SubscriptionStartedAt, &u.SubscriptionExpiresAt, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UserRepo) UpdateSubscription(ctx context.Context, id uuid.UUID, plan domain.SubscriptionPlan, studentsLimit int, tokensLimit int, expiresAt *time.Time) error {
	query := `
		UPDATE users
		SET subscription_plan = $1, students_limit = $2, tokens_limit = $3, subscription_expires_at = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, plan, studentsLimit, tokensLimit, expiresAt, id)
	return err
}

func (r *UserRepo) GetByPhoneWithDeleted(ctx context.Context, phone string) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE phone = $1`
	return r.scanUser(ctx, query, phone)
}

func (r *UserRepo) GetByEmailWithDeleted(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE email = $1`
	return r.scanUser(ctx, query, email)
}

func (r *UserRepo) GetByIINWithDeleted(ctx context.Context, iin string) (*domain.User, error) {
	query := `SELECT id, phone, email, iin, password_hash, first_name, last_name, middle_name, birth_year, gender, role, school_id, google_id, email_verified_at, avatar_url,
		subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at,
		created_at, updated_at, deleted_at FROM users WHERE iin = $1`
	return r.scanUser(ctx, query, iin)
}

func (r *UserRepo) Restore(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET phone=$1, email=$2, iin=$3, password_hash=$4, first_name=$5, last_name=$6, middle_name=$7, 
			birth_year=$8, gender=$9, role=$10, school_id=$11, google_id=$12, 
			subscription_plan=$13, tokens_used=$14, tokens_limit=$15, students_count=$16, students_limit=$17, 
			subscription_started_at=$18, created_at=$19, updated_at=NOW(), deleted_at=NULL, avatar_url=$21
		WHERE id=$20
	`
	_, err := r.db.Exec(ctx, query,
		user.Phone, user.Email, user.IIN, user.PasswordHash, user.FirstName, user.LastName, user.MiddleName,
		user.BirthYear, user.Gender, user.Role, user.SchoolID, user.GoogleID,
		user.SubscriptionPlan, user.TokensUsed, user.TokensLimit, user.StudentsCount, user.StudentsLimit,
		user.SubscriptionStartedAt, user.CreatedAt, user.ID, user.AvatarURL,
	)
	if err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}
	return nil
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, avatarURL, id)
	return err
}
