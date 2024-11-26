package repos

import (
	"context"
	"database/sql"
	"song-lib/internal/domain"
	"time"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

type SongRepository struct {
	db *sqlx.DB
}

type song struct {
	ID          ksuid.KSUID    `db:"id"`
	Name        string         `db:"name"`
	MusicGroup  musicGroup     `db:"music_group"`
	Couplets    pq.StringArray `db:"couplets"`
	ReleaseDate time.Time      `db:"release_date"`
	Link        string         `db:"link"`
}

type musicGroup struct {
	ID   ksuid.KSUID `db:"id"`
	Name string      `db:"name"`
}

func (r *SongRepository) SaveSong(
	ctx context.Context, song *domain.Song,
) (*domain.Song, error) {
	query := `
	WITH 
	upsert_music_group AS (
		INSERT INTO music_groups (name)
		VALUES ($1)
		ON CONFLICT (name) DO NOTHING
		RETURNING id
	),
	music_group_id AS (
		SELECT id FROM upsert_music_group
		UNION ALL
		SELECT id FROM music_groups WHERE name = $1
	),
	insert_song AS (
		INSERT INTO songs (id, music_group_id, name, release_date, link)
		VALUES (DEFAULT, (SELECT id FROM music_group_id), $2, $3, $4)
		RETURNING id
	)
	INSERT INTO 
		song_couplets (song_id, couplet_num, text)
	SELECT 
		(SELECT id FROM insert_song) AS song_id,
		ROW_NUMBER() OVER () AS couplet_num,
		text
	FROM 
		UNNEST($5::text[]) AS t(text)
	RETURNING 
		song_id`

	var songID ksuid.KSUID
	err := r.db.QueryRowxContext(
		ctx,
		query, song.MusicGroup.Name, song.Name,
		song.ReleaseDate, song.Link, pq.Array(song.Couplets),
	).Scan(&songID)
	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	resSong, err := r.GetSongByID(ctx, songID)
	if err != nil {
		return nil, err
	}
	return resSong, nil
}

func (r *SongRepository) SongExistsByID(
	ctx context.Context, songID ksuid.KSUID,
) (bool, error) {
	query, args, err := sq.
		Select("1").
		From("songs s").
		Where(sq.Eq{"s.id": songID}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, errors.Wrap(err, "build query")
	}

	var exists bool
	err = r.db.GetContext(ctx, &exists, query, args...)
	if err != nil {
		return false, errors.Wrap(err, "execute query")
	}

	return exists, nil
}

func (r *SongRepository) SongExistsByNameAndMusicGroupName(
	ctx context.Context, songName, musicGroupName string,
) (bool, error) {
	query, args, err := sq.
		Select("1").
		From("songs s").
		Join("music_groups mg ON s.music_group_id = mg.id").
		Where(sq.And{
			sq.Eq{"s.name": songName},
			sq.Eq{"mg.name": musicGroupName},
		}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, errors.Wrap(err, "build query")
	}

	var exists bool
	err = r.db.GetContext(ctx, &exists, query, args...)
	if err != nil {
		return false, errors.Wrap(err, "execute query")
	}

	return exists, nil
}

func (r *SongRepository) GetSongByID(
	ctx context.Context, songID ksuid.KSUID,
) (*domain.Song, error) {
	coupletsSubquery := sq.
		Select("ARRAY_AGG(sc.text ORDER BY sc.couplet_num)").
		From("song_couplets sc").
		Where("sc.song_id = s.id")

	query, args, err := sq.
		Select(
			"s.id",
			"s.name",
			"s.release_date",
			"s.link",
			`mg.id AS "music_group.id"`,
			`mg.name AS "music_group.name"`,
		).
		Column(sq.Alias(coupletsSubquery, "couplets")).
		From("songs s").
		LeftJoin("music_groups mg ON s.music_group_id = mg.id").
		Where(squirrel.Eq{"s.id": songID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build query")
	}

	var songModel song
	err = r.db.GetContext(ctx, &songModel, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	return songModel.toEntity(), nil
}

func (r *SongRepository) GetSongsFilteredPaginated(
	ctx context.Context, f *domain.SongFilters,
	pagination domain.Pagination,
) ([]domain.Song, error) {
	coupletsSubquery := sq.
		Select("ARRAY_AGG(sc.text ORDER BY sc.couplet_num)").
		From("song_couplets sc").
		Where("sc.song_id = s.id")

	builder := sq.
		Select(
			"s.id",
			"s.name",
			"s.release_date",
			"s.link",
			`mg.id AS "music_group.id"`,
			`mg.name AS "music_group.name"`,
		).
		Column(sq.Alias(coupletsSubquery, "couplets")).
		From("songs s").
		LeftJoin("music_groups mg ON s.music_group_id = mg.id").
		OrderBy("s.id").
		Limit(uint64(pagination.PerPage)).
		Offset(uint64(pagination.Page * pagination.PerPage))

	if f.SongName != nil {
		builder = builder.Where(sq.Eq{"s.name": *f.SongName})
	}
	if f.SongLink != nil {
		builder = builder.Where(sq.Eq{"s.link": *f.SongLink})
	}
	if f.MusicGroupName != nil {
		builder = builder.Where(sq.Eq{"mg.name": *f.MusicGroupName})
	}
	if f.SongReleaseDateRange != nil {
		builder = builder.Where(sq.And{
			sq.GtOrEq{"s.release_date": f.SongReleaseDateRange.StartTime},
			sq.LtOrEq{"s.release_date": f.SongReleaseDateRange.EndTime},
		})
	}
	if f.SongCoupletContains != nil {
		songWithTextIDsSubquery := sq.
			Select("sc.song_id").
			From("song_couplets sc").
			Where(sq.Expr(
				"sc.text ILIKE '%' || escape_like_string(?) || '%'",
				*f.SongCoupletContains))

		builder = builder.Where(inConditionWithSubquery(
			"s.id", songWithTextIDsSubquery,
		))
	}

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build query")
	}

	var songModels []song
	err = r.db.SelectContext(ctx, &songModels, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	songs := make([]domain.Song, 0, len(songModels))
	for _, songModel := range songModels {
		songs = append(songs, *songModel.toEntity())
	}

	return songs, nil
}

func (r *SongRepository) GetSongCoupletsPaginated(
	ctx context.Context, songID ksuid.KSUID,
	pagination domain.Pagination,
) ([]string, error) {
	query, args, err := sq.
		Select("sc.text").
		From("song_couplets sc").
		Where(sq.Eq{"sc.song_id": songID}).
		OrderBy("sc.couplet_num").
		Limit(uint64(pagination.PerPage)).
		Offset(uint64(pagination.Page * pagination.PerPage)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build query")
	}

	var couplets []string
	err = r.db.SelectContext(ctx, &couplets, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	return couplets, nil
}

func (r *SongRepository) UpdateSong(
	ctx context.Context, songID ksuid.KSUID,
	songUpdate *domain.SongUpdate,
) (*domain.Song, error) {
	builder := sq.
		Update("songs").
		Where(sq.Eq{"id": songID})
	if songUpdate.Name != nil {
		builder = builder.Set("name", *songUpdate.Name)
	}
	if songUpdate.ReleaseDate != nil {
		builder = builder.Set("release_date", *songUpdate.ReleaseDate)
	}
	if songUpdate.Link != nil {
		builder = builder.Set("link", *songUpdate.Link)
	}

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "update songs table: build query")
	}

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted})
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "update songs table: execute query")
	}

	if songUpdate.Couplets != nil {
		query, args, err := sq.
			Delete("song_couplets").
			Where(sq.Eq{"song_id": songID}).
			PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return nil, errors.Wrap(err,
				"delete old couplets from song_couplets table: build query")
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, errors.Wrap(err,
				"delete old couplets from song_couplets table: execute query")
		}

		query = `
		INSERT INTO 
			song_couplets (song_id, couplet_num, text)
		SELECT
			$1 AS song_id,
			ROW_NUMBER() OVER () AS couplet_num,
			text
		FROM 
			UNNEST($2::text[]) AS t(text)`

		_, err = tx.ExecContext(ctx, query, songID, pq.Array(*(songUpdate.Couplets)))
		if err != nil {
			return nil, errors.Wrap(err,
				"create new couplets in couplets table: execute query")
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "commit")
	}

	song, err := r.GetSongByID(ctx, songID)
	if err != nil {
		return nil, errors.Wrap(err, "get updated song")
	}

	return song, nil
}

func (r *SongRepository) DeleteSong(
	ctx context.Context, songID ksuid.KSUID,
) error {
	query, args, err := sq.
		Delete("songs").
		Where(sq.Eq{"id": songID}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return errors.Wrap(err, "build query")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "execute query")
	}

	return nil
}

func (s *song) toEntity() *domain.Song {
	return &domain.Song{
		ID:   s.ID,
		Name: s.Name,
		MusicGroup: domain.MusicGroup{
			ID:   s.MusicGroup.ID,
			Name: s.MusicGroup.Name,
		},
		Couplets:    s.Couplets,
		ReleaseDate: s.ReleaseDate,
		Link:        s.Link,
	}
}

func NewSongRepository(tx *sqlx.DB) *SongRepository {
	return &SongRepository{db: tx}
}
