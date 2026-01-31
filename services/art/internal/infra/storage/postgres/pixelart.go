package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/1ncrease0/pixly/services/art/internal/domain"
)

type PixelartRepo struct {
	pool *pgxpool.Pool
}

func NewPixelartRepo(pool *pgxpool.Pool) *PixelartRepo {
	return &PixelartRepo{pool: pool}
}

type canvasData struct {
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Palette []string `json:"palette"`
	Pixels  []int    `json:"pixels"`
}

func canvasToData(c domain.Canvas) (canvasData, error) {
	p := c.Palette()
	hexPalette := make([]string, 0, len(p))
	for i := range p {
		clr := &p[i]
		r := clr.RGBA()
		hexPalette = append(hexPalette, rgbaToHex(r))
	}
	return canvasData{
		Width:   c.Width(),
		Height:  c.Height(),
		Palette: hexPalette,
		Pixels:  c.Pixels(),
	}, nil
}

func rgbaToHex(r color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", r.R, r.G, r.B)
}

func dataToCanvas(d canvasData) (domain.Canvas, error) {
	palette := make([]domain.Color, 0, len(d.Palette))
	for _, hex := range d.Palette {
		c, err := domain.NewColor(hex)
		if err != nil {
			return domain.Canvas{}, fmt.Errorf("palette color: %w", err)
		}
		palette = append(palette, *c)
	}
	return domain.NewCanvas(d.Width, d.Height, palette, d.Pixels)
}

func (r *PixelartRepo) CreatePixelart(ctx context.Context, p *domain.Pixelart) error {
	if p == nil {
		return domain.ErrNilPixelart
	}
	data, err := canvasToData(p.Canvas())
	if err != nil {
		return fmt.Errorf("canvas to data: %w", err)
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO pixelarts (id, user_id, title, image_key, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())`,
		p.ID(), p.UserID(), p.Title().String(), p.ImageKey(), dataJSON,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrPixelartConflict
		}
		return fmt.Errorf("insert pixelart: %w", err)
	}
	return nil
}

func (r *PixelartRepo) PixelartByID(ctx context.Context, id uuid.UUID) (*domain.Pixelart, error) {
	var (
		userID    uuid.UUID
		title     string
		imageKey  *string
		dataJSON  []byte
		createdAt time.Time
		updatedAt time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, title, image_key, data, created_at, updated_at
		FROM pixelarts WHERE id = $1`,
		id,
	).Scan(&userID, &title, &imageKey, &dataJSON, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPixelartNotFound
		}
		return nil, fmt.Errorf("select pixelart: %w", err)
	}

	titleVO, err := domain.NewTitle(title)
	if err != nil {
		return nil, err
	}

	var d canvasData
	if err := json.Unmarshal(dataJSON, &d); err != nil {
		return nil, fmt.Errorf("unmarshal data: %w", err)
	}
	canvas, err := dataToCanvas(d)
	if err != nil {
		return nil, err
	}

	key := ""
	if imageKey != nil {
		key = *imageKey
	}
	return domain.NewPixelartFromStorage(id, userID, titleVO, canvas, key, createdAt, updatedAt), nil
}

func (r *PixelartRepo) PixelartsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Pixelart, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, title, image_key, data, created_at, updated_at
		FROM pixelarts WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("select pixelarts: %w", err)
	}
	defer rows.Close()

	var out []*domain.Pixelart
	for rows.Next() {
		var (
			id        uuid.UUID
			uid       uuid.UUID
			title     string
			imageKey  *string
			dataJSON  []byte
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&id, &uid, &title, &imageKey, &dataJSON, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		titleVO, err := domain.NewTitle(title)
		if err != nil {
			return nil, err
		}
		var d canvasData
		if err := json.Unmarshal(dataJSON, &d); err != nil {
			return nil, fmt.Errorf("unmarshal data: %w", err)
		}
		canvas, err := dataToCanvas(d)
		if err != nil {
			return nil, err
		}
		key := ""
		if imageKey != nil {
			key = *imageKey
		}
		out = append(out, domain.NewPixelartFromStorage(id, uid, titleVO, canvas, key, createdAt, updatedAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return out, nil
}

func (r *PixelartRepo) UpdatePixelart(ctx context.Context, p *domain.Pixelart) error {
	data, err := canvasToData(p.Canvas())
	if err != nil {
		return fmt.Errorf("canvas to data: %w", err)
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	result, err := r.pool.Exec(ctx, `
		UPDATE pixelarts SET title = $2, image_key = $3, data = $4, updated_at = now()
		WHERE id = $1`,
		p.ID(), p.Title().String(), p.ImageKey(), dataJSON,
	)
	if err != nil {
		return fmt.Errorf("update pixelart: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPixelartNotFound
	}
	return nil
}

func (r *PixelartRepo) DeletePixelart(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM pixelarts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete pixelart: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPixelartNotFound
	}
	return nil
}
