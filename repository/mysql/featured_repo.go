package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/repository"
)

type featuredRepo struct {
	db *sql.DB
}

func NewFeaturedRepository(db *sql.DB) repository.FeaturedRepository {
	return &featuredRepo{db: db}
}

func (r *featuredRepo) List(ctx context.Context) ([]map[string]interface{}, error) {
	query := `SELECT id, content_type, content_id, title, description, banner_url, click_url, order_index FROM featured_content ORDER BY order_index ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []map[string]interface{}
	for rows.Next() {
		var id, contentId, order int
		var cType, title, desc, banner, click sql.NullString
		
		if err := rows.Scan(&id, &cType, &contentId, &title, &desc, &banner, &click, &order); err == nil {
			bannerMap := map[string]interface{}{
				"id":           id,
				"content_type": cType.String,
				"content_id":   contentId,
				"title":        title.String,
				"description":  desc.String,
				"banner_url":   banner.String,
				"click_url":    click.String,
				"order_index":  order,
			}
			banners = append(banners, bannerMap)
		}
	}
	if banners == nil {
		banners = []map[string]interface{}{}
	}
	return banners, nil
}
