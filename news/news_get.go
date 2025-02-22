// SPDX-License-Identifier: ice License 1.0

package news

import (
	"context"

	"github.com/pkg/errors"

	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
	"github.com/ice-blockchain/wintr/time"
)

//nolint:revive,funlen // The alternative worse and requires to create one more struct.
func (r *repository) GetNews(ctx context.Context, newsType Type, language string, limit, offset uint64, createdAfter *time.Time) ([]*PersonalNews, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "get news failed because context failed")
	}
	args := []any{requestingUserID(ctx), language, newsType, int64(limit), int64(offset)}
	sql := `SELECT nvu.created_at IS NOT NULL AS viewed,
							   n.*
						FROM news n
							LEFT JOIN news_viewed_by_users nvu 
								   ON nvu.language = n.language
								  AND nvu.news_id = n.id
								  AND nvu.user_id = $1
						WHERE n.language = $2
							  AND n.type = $3
						ORDER BY 
							(CASE WHEN n.type = 'regular'
								THEN nvu.created_at IS NULL
								ELSE FALSE
					  		END) DESC,
							n.created_at DESC
						LIMIT $4 OFFSET $5`
	result, err := storage.Select[PersonalNews](ctx, r.db, sql, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get news for args:%#v", args...)
	}
	if result == nil {
		return []*PersonalNews{}, nil
	}
	trueVal := true
	for _, elem := range result {
		if elem.Viewed != nil && !*elem.Viewed && elem.CreatedAt.Before(*createdAfter.Time) {
			elem.Viewed = &trueVal
		}
		elem.NotificationChannels = nil
		elem.UpdatedAt = nil
		elem.ImageURL = r.pictureClient.DownloadURL(elem.ImageURL)
	}

	return result, nil
}

func (r *repository) GetUnreadNewsCount(ctx context.Context, language string, createdAfter *time.Time) (*UnreadNewsCount, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "context failed")
	}
	args := []any{requestingUserID(ctx), language, RegularNewsType, FeaturedNewsType, createdAfter.Time}
	sql := `SELECT COALESCE(COUNT(n.id), 0) as count
				FROM news n	
						LEFT JOIN news_viewed_by_users nvu
							ON nvu.language = n.language
							AND nvu.news_id = n.id 
							AND nvu.user_id = $1
				WHERE n.language = $2
					AND (n.type = $3 OR n.type = $4)
					AND n.created_at >= $5
					AND nvu.created_at IS NULL`
	result, err := storage.Get[UnreadNewsCount](ctx, r.db, sql, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get unread news count for params:%#v", args...)
	}

	return &UnreadNewsCount{Count: result.Count}, nil
}

func (r *repository) getNewsByPK(ctx context.Context, newsID, language string) (*TaggedNews, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "context failed")
	}
	sql := `SELECT array_agg(t.news_tag) filter (where t.news_tag is not null) AS tags,
				   n.* 
			FROM news n
			      LEFT JOIN news_tags_per_news t
			      		 ON t.language = n.language
			      		AND t.news_id  = n.id
			WHERE n.language = $1
			 		AND n.id = $2
			GROUP BY n.id, n.language, t.created_at
            ORDER BY t.created_at`
	result, err := storage.Get[TaggedNews](ctx, r.db, sql, language, newsID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select news article by (newsID:%v,language:%v)", newsID, language)
	}

	return result, nil
}
