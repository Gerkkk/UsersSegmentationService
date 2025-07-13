package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log/slog"
	"main/internal/domain/models"
	apperrors "main/internal/errors"
	"sync"
)

// SegmentationStorage - структура, которая управляет шардированной бд
type SegmentationStorage struct {
	shardsNum int
	dbShards  map[int]*sql.DB
	log       *slog.Logger
}

func NewSegmentationStorage(numShards int, dsns []string, log *slog.Logger) (*SegmentationStorage, error) {
	dbShards := make(map[int]*sql.DB)

	for i, dsn := range dsns {
		db, err := sql.Open("postgres", dsn)

		if err != nil {
			retError := fmt.Errorf("error connecting to database with DSN %s: %w", dsn, err)
			return nil, retError
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		if err := db.Ping(); err != nil {
			retError := fmt.Errorf("error pinging database with DSN %s: %w", dsn, err)
			return nil, retError
		}

		dbShards[i] = db
	}

	segStorage := &SegmentationStorage{shardsNum: numShards, dbShards: dbShards, log: log}

	return segStorage, nil
}

// CreateSegment - создать сегмент во всех шардах
func (s *SegmentationStorage) CreateSegment(segment models.Segment) (string, error) {
	ctx := context.Background()
	txID := "tx_" + uuid.New().String()
	preparedShards := make(map[int]bool)

	for shardID, db := range s.dbShards {
		conn, err := db.Conn(ctx)
		if err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: failed to get DB connection: %w", shardID, err)
		}

		defer conn.Close()

		if _, err := conn.ExecContext(ctx, "BEGIN"); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: begin failed: %w", shardID, err)
		}

		_, err = conn.ExecContext(ctx,
			"INSERT INTO segments (id, description) VALUES ($1, $2)",
			segment.Id, segment.Description)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				_, _ = conn.ExecContext(ctx, "ROLLBACK")
				s.rollbackAll(txID, preparedShards)
				return "", apperrors.ErrSegmentAlreadyExists
			}

			_, _ = conn.ExecContext(ctx, "ROLLBACK")
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: insert failed: %w", shardID, err)
		}

		prepareQuery := fmt.Sprintf("PREPARE TRANSACTION '%s'", txID)
		if _, err := conn.ExecContext(ctx, prepareQuery); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: prepare failed: %w", shardID, err)
		}

		preparedShards[shardID] = true
	}

	if err := s.commitAll(txID, preparedShards); err != nil {
		return "", fmt.Errorf("commit failed: %w", err)
	}

	return segment.Id, nil
}

/*
	DeleteSegment - удалить из шардов все записи о сегменте с таким id.

Если хотя бы где-то существует сегмент - удаляем, иначе вернем ошибку
*/
func (s *SegmentationStorage) DeleteSegment(id string) (string, error) {
	ctx := context.Background()
	txID := "tx_" + uuid.New().String()
	preparedShards := make(map[int]bool)
	segmentFound := false

	for shardID, db := range s.dbShards {
		conn, err := db.Conn(ctx)
		if err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: failed to get DB connection: %w", shardID, err)
		}
		defer conn.Close()

		if _, err := conn.ExecContext(ctx, "BEGIN"); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: begin failed: %w", shardID, err)
		}

		result, err := conn.ExecContext(ctx, "DELETE FROM segments WHERE id = $1", id)
		if err != nil {
			_, _ = conn.ExecContext(ctx, "ROLLBACK")
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: delete failed: %w", shardID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			_, _ = conn.ExecContext(ctx, "ROLLBACK")
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: failed to get affected rows: %w", shardID, err)
		}

		if rowsAffected > 0 {
			segmentFound = true
		}

		prepareQuery := fmt.Sprintf("PREPARE TRANSACTION '%s'", txID)
		if _, err := conn.ExecContext(ctx, prepareQuery); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: prepare failed: %w", shardID, err)
		}

		preparedShards[shardID] = true
	}

	if !segmentFound {
		s.rollbackAll(txID, preparedShards)
		return "", apperrors.ErrSegmentNotFound
	}

	if err := s.commitAll(txID, preparedShards); err != nil {
		return "", fmt.Errorf("commit failed: %w", err)
	}

	return id, nil
}

/*
	UpdateSegment - обновить записи о сегменте с таким id во всех шардах.

Если хотя бы где-то существует сегмент - обновляем, иначе вернем ошибку
*/
func (s *SegmentationStorage) UpdateSegment(id string, newSegment models.Segment) (string, error) {
	ctx := context.Background()
	txID := "tx_" + uuid.New().String()
	preparedShards := make(map[int]bool)
	segmentFound := false

	for shardID, db := range s.dbShards {
		conn, err := db.Conn(ctx)
		if err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: failed to get DB connection: %w", shardID, err)
		}
		defer conn.Close()

		if _, err := conn.ExecContext(ctx, "BEGIN"); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: begin failed: %w", shardID, err)
		}

		result, err := conn.ExecContext(ctx,
			"UPDATE segments SET description = $1 WHERE id = $2",
			newSegment.Description, id)
		if err != nil {
			_, _ = conn.ExecContext(ctx, "ROLLBACK")
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: update failed: %w", shardID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			_, _ = conn.ExecContext(ctx, "ROLLBACK")
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: failed to get affected rows: %w", shardID, err)
		}

		if rowsAffected > 0 {
			segmentFound = true
		}

		prepareQuery := fmt.Sprintf("PREPARE TRANSACTION '%s'", txID)
		if _, err := conn.ExecContext(ctx, prepareQuery); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: prepare failed: %w", shardID, err)
		}

		preparedShards[shardID] = true
	}

	if !segmentFound {
		s.rollbackAll(txID, preparedShards)
		return "", apperrors.ErrSegmentNotFound
	}

	if err := s.commitAll(txID, preparedShards); err != nil {
		return "", fmt.Errorf("commit failed: %w", err)
	}

	return id, nil
}

/* GetUserSegments - Получить данные о сегментах, в которых есть заданный пользователь
 */
func (s *SegmentationStorage) GetUserSegments(id int) ([]models.Segment, error) {
	ctx := context.Background()
	shardNum := id % s.shardsNum
	db := s.dbShards[shardNum]

	var exists bool
	err := db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if !exists {
		return nil, apperrors.ErrUserNotFound
	}

	rows, err := db.QueryContext(ctx, `
        SELECT seg.id, seg.description
        FROM users_segments us
        JOIN segments seg ON us.segment_id = seg.id
        WHERE us.user_id = $1
    `, id)

	if err != nil {
		return nil, fmt.Errorf("failed to query user segments: %w", err)
	}
	defer rows.Close()

	segments := []models.Segment{}
	for rows.Next() {
		var seg models.Segment
		if err := rows.Scan(&seg.Id, &seg.Description); err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}
		segments = append(segments, seg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return segments, nil
}

/*
	GetUserSegments - Получить статистику сегмента по id.

Ошибка только если нигде не нашли сегмент. Иначе информацию выведем
*/
func (s *SegmentationStorage) GetSegmentInfo(id string) (models.SegmentInfo, error) {
	type result struct {
		info models.SegmentInfo
		err  error
	}

	ctx := context.Background()
	resultCh := make(chan result)
	wg := sync.WaitGroup{}

	for shardID, db := range s.dbShards {
		wg.Add(1)

		go func(shardID int, db *sql.DB) {
			defer wg.Done()

			query := `
				WITH
				cnt AS (
					SELECT COUNT(*) as users_count FROM users_segments WHERE segment_id = $1
				),
				info AS (
					SELECT id, description FROM segments WHERE id = $1
				)
				SELECT info.id, info.description, cnt.users_count
				FROM cnt JOIN info ON TRUE;
			`

			row := db.QueryRowContext(ctx, query, id)

			var si models.SegmentInfo
			err := row.Scan(&si.Id, &si.Description, &si.UsersNum)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					resultCh <- result{err: apperrors.ErrSegmentNotFound}
					return
				}
				resultCh <- result{err: fmt.Errorf("shard %d: %w", shardID, err)}
				return
			}

			resultCh <- result{info: si}
		}(shardID, db)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	cumResult := models.SegmentInfo{}
	found := false
	errorsFound := []error{}

	for res := range resultCh {
		if res.err != nil {
			errorsFound = append(errorsFound, res.err)
			continue
		}
		if res.info.Id != "" {
			if !found {
				cumResult.Id = res.info.Id
				cumResult.Description = res.info.Description
				found = true
			}
			cumResult.UsersNum += res.info.UsersNum
		}
	}

	if !found {
		s.log.Error(fmt.Sprintf("failed to read user segment info. Errors: %v", errorsFound), slog.String("id", id))
		return models.SegmentInfo{}, apperrors.ErrSegmentNotFound
	}

	return cumResult, nil
}

/* DistributeSegment - распространить сегмент на пользователей, если только среди новых пользователей нет уже добавленных записей
 */
func (s *SegmentationStorage) DistributeSegment(id string, usersPercentage int) (string, error) {
	ctx := context.Background()
	txID := "tx_" + uuid.New().String()
	preparedShards := make(map[int]bool)

	percentage := float64(usersPercentage)
	// Добавим 5 процентов, чтобы уменьшить вероятность выборки меньшего числа пользователей
	upperPercentage := percentage + 5
	if upperPercentage > 100 {
		upperPercentage = 100
	}

	for shardID, db := range s.dbShards {
		conn, err := db.Conn(ctx)
		if err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: failed to get DB connection: %w", shardID, err)
		}

		defer conn.Close()

		if _, err := conn.ExecContext(ctx, "BEGIN"); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: begin failed: %w", shardID, err)
		}

		query := `
			WITH
			target_limit AS (
				SELECT COUNT(*) * $2 / 100 AS max_count FROM users
			),
			users_to_add AS (
				SELECT id FROM users TABLESAMPLE BERNOULLI($3)
			),
			users_to_add_limited AS (
				SELECT id FROM users_to_add LIMIT (SELECT max_count FROM target_limit)
			),
			seg AS (
				SELECT id FROM segments WHERE id = $1
			),
			new_vals AS (
				SELECT users_to_add_limited.id as user_id, seg.id as segment_id FROM users_to_add_limited JOIN seg ON TRUE
			)
			INSERT INTO users_segments (user_id, segment_id)
			SELECT user_id, segment_id FROM new_vals;
		`

		_, err = conn.ExecContext(ctx, query, id, percentage, upperPercentage)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				_, _ = conn.ExecContext(ctx, "ROLLBACK")
				s.rollbackAll(txID, preparedShards)
				return "", apperrors.ErrSegmentDistributed
			}

			_, _ = conn.ExecContext(ctx, "ROLLBACK")
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: insert failed: %w", shardID, err)
		}

		prepareQuery := fmt.Sprintf("PREPARE TRANSACTION '%s'", txID)
		if _, err := conn.ExecContext(ctx, prepareQuery); err != nil {
			s.rollbackAll(txID, preparedShards)
			return "", fmt.Errorf("shard %d: prepare failed: %w", shardID, err)
		}

		preparedShards[shardID] = true
	}

	if err := s.commitAll(txID, preparedShards); err != nil {
		return "", fmt.Errorf("commit failed: %w", err)
	}

	return id, nil
}

/*
CreateUser - создать пользователя в нужном шарде
*/
func (s *SegmentationStorage) CreateUser(user models.User) (int, error) {
	ctx := context.Background()
	shardNum := user.Id % s.shardsNum
	db := s.dbShards[shardNum]

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf("failed to begin tx: %w", err)
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO users (id) VALUES ($1)", user.Id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return -1, apperrors.ErrUserExists
		}
		return -1, fmt.Errorf("insert failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return -1, fmt.Errorf("commit failed: %w", err)
	}

	return user.Id, nil
}

/*
DeleteUser - удалить пользователя в нужном шарде
*/
func (s *SegmentationStorage) DeleteUser(id int) (int, error) {
	ctx := context.Background()
	shardNum := id % s.shardsNum
	db := s.dbShards[shardNum]

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf("failed to begin tx: %w", err)
	}

	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return -1, fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return -1, fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return -1, apperrors.ErrUserNotFound
	}

	if err = tx.Commit(); err != nil {
		return -1, fmt.Errorf("commit failed: %w", err)
	}

	return id, nil
}

// rollbackAll - пробуем роллбэкнуть транзакции во всех шардах
func (s *SegmentationStorage) rollbackAll(txID string, preparedShards map[int]bool) {
	for shardID, prepared := range preparedShards {
		if !prepared {
			continue
		}

		db := s.dbShards[shardID]
		query := fmt.Sprintf("ROLLBACK PREPARED '%s'", txID)
		if _, err := db.Exec(query); err != nil {
			s.log.Error("rollback failed", "shard", shardID, "txID", txID, "error", err)
		}
	}
}

// commitAll - пробуем закоммитить подготовленные транзакции во всех шардах. Может и не получиться. Но это маловероятно.
func (s *SegmentationStorage) commitAll(txID string, preparedShards map[int]bool) error {
	var firstErr error

	for shardID, prepared := range preparedShards {
		if !prepared {
			continue
		}

		db := s.dbShards[shardID]
		query := fmt.Sprintf("COMMIT PREPARED '%s'", txID)
		if _, err := db.Exec(query); err != nil {
			s.log.Error("commit failed", "shard", shardID, "txID", txID, "error", err)
			if firstErr == nil {
				firstErr = fmt.Errorf("commit failed on shard %d: %w", shardID, err)
			}
		}
	}

	if firstErr != nil {
		s.rollbackAll(txID, preparedShards)
		return firstErr
	}

	return nil
}
