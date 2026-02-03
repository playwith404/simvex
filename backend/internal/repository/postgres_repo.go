package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"simvex/internal/models"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &PostgresRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	if err := repo.seedIfEmpty(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) migrate() error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,
		`CREATE TABLE IF NOT EXISTS objects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			thumbnail TEXT,
			model_path TEXT,
			theory TEXT,
			category TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS parts (
			id TEXT PRIMARY KEY,
			object_id TEXT NOT NULL,
			name TEXT NOT NULL,
			material TEXT,
			role TEXT,
			model_path TEXT,
			local_pos_x DOUBLE PRECISION DEFAULT 0,
			local_pos_y DOUBLE PRECISION DEFAULT 0,
			local_pos_z DOUBLE PRECISION DEFAULT 0,
			decompose_dir_x DOUBLE PRECISION DEFAULT 0,
			decompose_dir_y DOUBLE PRECISION DEFAULT 1,
			decompose_dir_z DOUBLE PRECISION DEFAULT 0,
			decompose_distance DOUBLE PRECISION DEFAULT 1,
			FOREIGN KEY (object_id) REFERENCES objects(id)
		);`,
		`CREATE TABLE IF NOT EXISTS assets (
			path TEXT PRIMARY KEY,
			content_type TEXT NOT NULL,
			data BYTEA NOT NULL,
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS notion_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_encrypted TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS workflow_projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			notion_page_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS notes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			part_id TEXT NOT NULL,
			content TEXT NOT NULL,
			notion_page_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS workflow_nodes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES workflow_projects(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT,
			scheduled_date DATE NOT NULL,
			progress INTEGER NOT NULL DEFAULT 0,
			color TEXT,
			position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
			position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
			linked_part_id TEXT,
			linked_note_id UUID REFERENCES notes(id) ON DELETE SET NULL,
			notion_page_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS workflow_edges (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES workflow_projects(id) ON DELETE CASCADE,
			source_node_id UUID NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
			target_node_id UUID NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS node_checklists (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			node_id UUID NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
			text TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT FALSE
		);`,
		`CREATE TABLE IF NOT EXISTS node_attachments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			node_id UUID NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
			type TEXT NOT NULL CHECK (type IN ('link', 'file')),
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_projects_user_id ON workflow_projects(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_nodes_project_id ON workflow_nodes(project_id);`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_edges_project_id ON workflow_edges(project_id);`,
		`CREATE INDEX IF NOT EXISTS idx_notes_user_part ON notes(user_id, part_id);`,
	}

	for _, q := range queries {
		if _, err := r.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) seedIfEmpty() error {
	row := r.db.QueryRow("SELECT COUNT(1) FROM objects")
	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	objects := []models.Object{
		{
			ID:          "engine-v4",
			Name:        "V4 엔진",
			Description: "자동차용 4기통 엔진의 내부 구조를 학습합니다.",
			Thumbnail:   "/assets/models/engine-v4/thumbnail.png",
			ModelPath:   "/assets/models/engine-v4/piston.glb",
			Theory:      "열역학, 4행정 사이클, 피스톤 왕복운동과 크랭크 회전운동 변환",
			Category:    "Powertrain",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "suspension",
			Name:        "서스펜션",
			Description: "차량 현가장치의 스프링-댐퍼 구조를 학습합니다.",
			Thumbnail:   "/assets/models/suspension/thumbnail.png",
			ModelPath:   "/assets/models/suspension/base.glb",
			Theory:      "진동, 감쇠, 하중 분산",
			Category:    "Chassis",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "robot-arm",
			Name:        "로봇 암",
			Description: "산업용 로봇 팔의 링크 구조와 회전축을 학습합니다.",
			Thumbnail:   "/assets/models/robot-arm/thumbnail.png",
			ModelPath:   "/assets/models/robot-arm/base.glb",
			Theory:      "강체 운동학, 관절 회전, 토크 전달",
			Category:    "Robotics",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "machine-vice",
			Name:        "공작 바이스",
			Description: "공작물 고정 장치의 스핀들 구동과 죠 구조를 학습합니다.",
			Thumbnail:   "/assets/models/machine-vice/thumbnail.png",
			ModelPath:   "/assets/models/machine-vice/part1.glb",
			Theory:      "나사 구동, 마찰, 고정력 전달",
			Category:    "Manufacturing",
			CreatedAt:   time.Now(),
		},
	}

	parts := []models.Part{
		// Engine V4
		{id("engine-v4", "piston"), "engine-v4", "피스톤", "알루미늄 합금", "연소 압력을 받아 왕복운동을 수행", "/assets/models/engine-v4/piston.glb", 0, 0, 0, 0, 1, 0, 1.2},
		{id("engine-v4", "piston-ring"), "engine-v4", "피스톤 링", "강재", "실린더 내부 기밀 유지", "/assets/models/engine-v4/piston-ring.glb", 0, 0, 0, 0, 1, 0, 1.4},
		{id("engine-v4", "piston-pin"), "engine-v4", "피스톤 핀", "강재", "피스톤과 커넥팅 로드를 연결", "/assets/models/engine-v4/piston-pin.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("engine-v4", "connecting-rod"), "engine-v4", "커넥팅 로드", "단조강", "피스톤의 운동을 크랭크축에 전달", "/assets/models/engine-v4/connecting-rod.glb", 0, 0, 0, 0, -1, 0, 1.3},
		{id("engine-v4", "connecting-rod-cap"), "engine-v4", "커넥팅 로드 캡", "단조강", "커넥팅 로드와 크랭크축 결합", "/assets/models/engine-v4/connecting-rod-cap.glb", 0, 0, 0, 0, -1, 0, 1.1},
		{id("engine-v4", "conrod-bolt"), "engine-v4", "콘로드 볼트", "강재", "커넥팅 로드 체결", "/assets/models/engine-v4/conrod-bolt.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("engine-v4", "crankshaft"), "engine-v4", "크랭크축", "단조강", "왕복운동을 회전운동으로 변환", "/assets/models/engine-v4/crankshaft.glb", 0, 0, 0, 0, 0, 1, 1.5},

		// Suspension
		{id("suspension", "base"), "suspension", "베이스", "강재", "하중을 지지하는 본체", "/assets/models/suspension/base.glb", 0, 0, 0, 0, -1, 0, 0.9},
		{id("suspension", "rod"), "suspension", "로드", "강재", "스프링과 베이스 연결", "/assets/models/suspension/rod.glb", 0, 0, 0, 0, 1, 0, 1.0},
		{id("suspension", "spring"), "suspension", "스프링", "스프링강", "충격 흡수 및 복원력 제공", "/assets/models/suspension/spring.glb", 0, 0, 0, 0, 1, 0, 1.2},
		{id("suspension", "nut"), "suspension", "너트", "강재", "체결 고정", "/assets/models/suspension/nut.glb", 0, 0, 0, 1, 0, 0, 0.7},
		{id("suspension", "cap"), "suspension", "상단 캡", "강재", "상부 체결 및 보호", "/assets/models/suspension/nit.glb", 0, 0, 0, 0, 1, 0, 0.7},

		// Robot Arm
		{id("robot-arm", "base"), "robot-arm", "베이스", "알루미늄", "하부 지지 및 회전축", "/assets/models/robot-arm/base.glb", 0, 0, 0, 0, -1, 0, 1.0},
		{id("robot-arm", "link-1"), "robot-arm", "링크 1", "알루미늄", "첫 번째 링크", "/assets/models/robot-arm/part2.glb", 0, 0, 0, 1, 0, 0, 1.2},
		{id("robot-arm", "link-2"), "robot-arm", "링크 2", "알루미늄", "두 번째 링크", "/assets/models/robot-arm/part3.glb", 0, 0, 0, -1, 0, 0, 1.2},
		{id("robot-arm", "joint-1"), "robot-arm", "관절 1", "합금강", "회전 관절", "/assets/models/robot-arm/part4.glb", 0, 0, 0, 0, 1, 0, 0.9},
		{id("robot-arm", "joint-2"), "robot-arm", "관절 2", "합금강", "회전 관절", "/assets/models/robot-arm/part5.glb", 0, 0, 0, 0, 1, 0, 0.9},
		{id("robot-arm", "link-3"), "robot-arm", "링크 3", "알루미늄", "세 번째 링크", "/assets/models/robot-arm/part6.glb", 0, 0, 0, 0, 0, 1, 1.1},
		{id("robot-arm", "link-4"), "robot-arm", "링크 4", "알루미늄", "말단 링크", "/assets/models/robot-arm/part7.glb", 0, 0, 0, 0, 0, 1, 1.1},
		{id("robot-arm", "end-effector"), "robot-arm", "엔드 이펙터", "알루미늄", "작업 도구 장착부", "/assets/models/robot-arm/part8.glb", 0, 0, 0, 0, 1, 0, 1.0},

		// Machine Vice
		{id("machine-vice", "body"), "machine-vice", "본체", "주강", "바이스 본체", "/assets/models/machine-vice/part1.glb", 0, 0, 0, 0, -1, 0, 1.1},
		{id("machine-vice", "guide"), "machine-vice", "가이드", "주강", "죠 이동 가이드", "/assets/models/machine-vice/part1-fuhrung.glb", 0, 0, 0, 1, 0, 0, 0.9},
		{id("machine-vice", "fixed-jaw"), "machine-vice", "고정 죠", "합금강", "고정 측 죠", "/assets/models/machine-vice/part2-feste-backe.glb", 0, 0, 0, 0, 1, 0, 1.0},
		{id("machine-vice", "movable-jaw"), "machine-vice", "이동 죠", "합금강", "가동 측 죠", "/assets/models/machine-vice/part3-lose-backe.glb", 0, 0, 0, 0, 1, 0, 1.0},
		{id("machine-vice", "spindle-base"), "machine-vice", "스핀들 베이스", "강재", "스핀들 지지", "/assets/models/machine-vice/part4-spindelsockel.glb", 0, 0, 0, 0, 0, 1, 1.0},
		{id("machine-vice", "clamp-jaw"), "machine-vice", "클램프 죠", "합금강", "클램핑 압력 전달", "/assets/models/machine-vice/part5-spannbacke.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("machine-vice", "guide-rail"), "machine-vice", "가이드 레일", "강재", "가동부 이동 레일", "/assets/models/machine-vice/part6-fuhrungschiene.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("machine-vice", "spindle"), "machine-vice", "트라페조이드 스핀들", "강재", "나사 구동부", "/assets/models/machine-vice/part7-trapezspindel.glb", 0, 0, 0, 0, 0, 1, 1.2},
		{id("machine-vice", "base-plate"), "machine-vice", "베이스 플레이트", "주강", "하부 지지", "/assets/models/machine-vice/part8-grundplatte.glb", 0, 0, 0, 0, -1, 0, 1.1},
		{id("machine-vice", "pressure-sleeve"), "machine-vice", "프레셔 슬리브", "강재", "압력 전달", "/assets/models/machine-vice/part9-druckhulse.glb", 0, 0, 0, 0, 1, 0, 0.9},
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	objStmt, err := tx.Prepare(`INSERT INTO objects (id, name, description, thumbnail, model_path, theory, category, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return rollback(tx, err)
	}
	defer objStmt.Close()

	for _, obj := range objects {
		if _, err := objStmt.Exec(obj.ID, obj.Name, obj.Description, obj.Thumbnail, obj.ModelPath, obj.Theory, obj.Category, obj.CreatedAt); err != nil {
			return rollback(tx, err)
		}
	}

	partStmt, err := tx.Prepare(`INSERT INTO parts (id, object_id, name, material, role, model_path, local_pos_x, local_pos_y, local_pos_z, decompose_dir_x, decompose_dir_y, decompose_dir_z, decompose_distance) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`)
	if err != nil {
		return rollback(tx, err)
	}
	defer partStmt.Close()

	for _, part := range parts {
		if _, err := partStmt.Exec(part.ID, part.ObjectID, part.Name, part.Material, part.Role, part.ModelPath, part.LocalPosX, part.LocalPosY, part.LocalPosZ, part.DecomposeDirX, part.DecomposeDirY, part.DecomposeDirZ, part.DecomposeDistance); err != nil {
			return rollback(tx, err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) SeedAssetsFromDir(basePath string) error {
	if basePath == "" {
		return nil
	}
	if _, err := os.Stat(basePath); err != nil {
		return nil
	}

	row := r.db.QueryRow("SELECT COUNT(1) FROM assets")
	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO assets (path, content_type, data, updated_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT (path) DO NOTHING`)
	if err != nil {
		return rollback(tx, err)
	}
	defer stmt.Close()

	walkErr := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !isAssetExt(ext) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		webPath := "/assets/models/" + rel
		contentType := contentTypeForExt(ext, data)

		if _, err := stmt.Exec(webPath, contentType, data); err != nil {
			return err
		}
		return nil
	})
	if walkErr != nil {
		return rollback(tx, walkErr)
	}

	return tx.Commit()
}

func rollback(tx *sql.Tx, err error) error {
	if rbErr := tx.Rollback(); rbErr != nil {
		return fmt.Errorf("rollback error: %w (original: %v)", rbErr, err)
	}
	return err
}

func id(objectID, part string) string {
	return fmt.Sprintf("%s_%s", objectID, part)
}

func (r *PostgresRepository) GetObjects() ([]models.Object, error) {
	rows, err := r.db.Query(`SELECT id, name, description, thumbnail, model_path, theory, category, created_at FROM objects ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objects []models.Object
	for rows.Next() {
		var obj models.Object
		if err := rows.Scan(&obj.ID, &obj.Name, &obj.Description, &obj.Thumbnail, &obj.ModelPath, &obj.Theory, &obj.Category, &obj.CreatedAt); err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func (r *PostgresRepository) GetObjectByID(id string) (*models.Object, error) {
	row := r.db.QueryRow(`SELECT id, name, description, thumbnail, model_path, theory, category, created_at FROM objects WHERE id = $1`, id)
	var obj models.Object
	if err := row.Scan(&obj.ID, &obj.Name, &obj.Description, &obj.Thumbnail, &obj.ModelPath, &obj.Theory, &obj.Category, &obj.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &obj, nil
}

func (r *PostgresRepository) GetPartsByObjectID(objectID string) ([]models.Part, error) {
	rows, err := r.db.Query(`SELECT id, object_id, name, material, role, model_path, local_pos_x, local_pos_y, local_pos_z, decompose_dir_x, decompose_dir_y, decompose_dir_z, decompose_distance FROM parts WHERE object_id = $1`, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.Part
	for rows.Next() {
		var part models.Part
		if err := rows.Scan(&part.ID, &part.ObjectID, &part.Name, &part.Material, &part.Role, &part.ModelPath, &part.LocalPosX, &part.LocalPosY, &part.LocalPosZ, &part.DecomposeDirX, &part.DecomposeDirY, &part.DecomposeDirZ, &part.DecomposeDistance); err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func (r *PostgresRepository) GetPartByID(partID string) (*models.Part, error) {
	row := r.db.QueryRow(`SELECT id, object_id, name, material, role, model_path, local_pos_x, local_pos_y, local_pos_z, decompose_dir_x, decompose_dir_y, decompose_dir_z, decompose_distance FROM parts WHERE id = $1`, partID)
	var part models.Part
	if err := row.Scan(&part.ID, &part.ObjectID, &part.Name, &part.Material, &part.Role, &part.ModelPath, &part.LocalPosX, &part.LocalPosY, &part.LocalPosZ, &part.DecomposeDirX, &part.DecomposeDirY, &part.DecomposeDirZ, &part.DecomposeDistance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &part, nil
}

func (r *PostgresRepository) GetAssetByPath(path string) (*models.Asset, error) {
	row := r.db.QueryRow(`SELECT path, content_type, data FROM assets WHERE path = $1`, path)
	var asset models.Asset
	if err := row.Scan(&asset.Path, &asset.ContentType, &asset.Data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &asset, nil
}

func isAssetExt(ext string) bool {
	switch ext {
	case ".glb", ".gltf", ".bin", ".png", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}

func contentTypeForExt(ext string, data []byte) string {
	switch ext {
	case ".glb":
		return "model/gltf-binary"
	case ".gltf":
		return "model/gltf+json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		if detected := mime.TypeByExtension(ext); detected != "" {
			return detected
		}
		if len(data) > 0 {
			return http.DetectContentType(data)
		}
		return "application/octet-stream"
	}
}

func (r *PostgresRepository) CreateUser(email, passwordHash string) (*models.User, error) {
	row := r.db.QueryRow(
		`INSERT INTO users (email, password_hash, is_verified, created_at)
		 VALUES ($1, $2, FALSE, NOW())
		 RETURNING id, email, password_hash, is_verified, created_at`,
		email,
		passwordHash,
	)

	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, is_verified, created_at FROM users WHERE email = $1`,
		email,
	)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByID(id string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, is_verified, created_at FROM users WHERE id = $1`,
		id,
	)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) SetUserVerified(id string) error {
	_, err := r.db.Exec(`UPDATE users SET is_verified = TRUE WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) UpdateUserPassword(id, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, passwordHash, id)
	return err
}
