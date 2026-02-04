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
			parent_page_id TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (user_id)
		);`,
		`DO $$ BEGIN
			ALTER TABLE notion_tokens ADD COLUMN IF NOT EXISTS parent_page_id TEXT NOT NULL DEFAULT '';
		EXCEPTION WHEN duplicate_column THEN NULL;
		END $$;`,
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
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_notes_user_part_unique ON notes(user_id, part_id);`,
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
			Name:        "V4 Engine",
			Description: "Learn the internal structure of a 4-cylinder engine.",
			Thumbnail:   "/assets/models/engine-v4/thumbnail.png",
			ModelPath:   "/assets/models/engine-v4/piston.glb",
			Theory:      "Thermodynamics, 4-stroke cycle, piston-crank rotation",
			Category:    "Powertrain",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "suspension",
			Name:        "Suspension",
			Description: "Learn suspension structure and motion.",
			Thumbnail:   "/assets/models/suspension/thumbnail.png",
			ModelPath:   "/assets/models/suspension/base.glb",
			Theory:      "Vibration, damping, load distribution",
			Category:    "Chassis",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "robot-arm",
			Name:        "Robot Arm",
			Description: "Learn link structure and rotation axes.",
			Thumbnail:   "/assets/models/robot-arm/thumbnail.png",
			ModelPath:   "/assets/models/robot-arm/base.glb",
			Theory:      "Rigid body motion, rotation, torque",
			Category:    "Robotics",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "machine-vice",
			Name:        "Machine Vice",
			Description: "Learn the clamping mechanism and screw drive.",
			Thumbnail:   "/assets/models/machine-vice/thumbnail.png",
			ModelPath:   "/assets/models/machine-vice/part1.glb",
			Theory:      "Screw drive, friction, clamping force",
			Category:    "Manufacturing",
			CreatedAt:   time.Now(),
		},
	}

	parts := []models.Part{
		// Engine V4
		{id("engine-v4", "piston"), "engine-v4", "Piston", "Aluminum alloy", "Receives combustion force", "/assets/models/engine-v4/piston.glb", 0, 0, 0, 0, 1, 0, 1.2},
		{id("engine-v4", "piston-ring"), "engine-v4", "Piston Ring", "Steel", "Seals cylinder and controls oil", "/assets/models/engine-v4/piston-ring.glb", 0, 0, 0, 0, 1, 0, 1.4},
		{id("engine-v4", "piston-pin"), "engine-v4", "Piston Pin", "Steel", "Connects piston and rod", "/assets/models/engine-v4/piston-pin.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("engine-v4", "connecting-rod"), "engine-v4", "Connecting Rod", "Steel", "Transmits motion to crankshaft", "/assets/models/engine-v4/connecting-rod.glb", 0, 0, 0, 0, -1, 0, 1.3},
		{id("engine-v4", "connecting-rod-cap"), "engine-v4", "Rod Cap", "Steel", "Secures rod to crankshaft", "/assets/models/engine-v4/connecting-rod-cap.glb", 0, 0, 0, 0, -1, 0, 1.1},
		{id("engine-v4", "conrod-bolt"), "engine-v4", "Rod Bolt", "Steel", "Fastens the rod cap", "/assets/models/engine-v4/conrod-bolt.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("engine-v4", "crankshaft"), "engine-v4", "Crankshaft", "Steel", "Converts linear to rotary motion", "/assets/models/engine-v4/crankshaft.glb", 0, 0, 0, 0, 0, 1, 1.5},

		// Suspension
		{id("suspension", "base"), "suspension", "Base", "Steel", "Structural base frame", "/assets/models/suspension/base.glb", 0, 0, 0, 0, -1, 0, 0.9},
		{id("suspension", "rod"), "suspension", "Rod", "Steel", "Connects components", "/assets/models/suspension/rod.glb", 0, 0, 0, 0, 1, 0, 1.0},
		{id("suspension", "spring"), "suspension", "Spring", "Spring steel", "Stores and releases energy", "/assets/models/suspension/spring.glb", 0, 0, 0, 0, 1, 0, 1.2},
		{id("suspension", "nut"), "suspension", "Nut", "Steel", "Fastening element", "/assets/models/suspension/nut.glb", 0, 0, 0, 1, 0, 0, 0.7},
		{id("suspension", "cap"), "suspension", "Cap", "Steel", "Top cap and guard", "/assets/models/suspension/nit.glb", 0, 0, 0, 0, 1, 0, 0.7},

		// Robot Arm
		{id("robot-arm", "base"), "robot-arm", "Base", "Aluminum", "Supports the arm", "/assets/models/robot-arm/base.glb", 0, 0, 0, 0, -1, 0, 1.0},
		{id("robot-arm", "link-1"), "robot-arm", "Link 1", "Aluminum", "First link", "/assets/models/robot-arm/part2.glb", 0, 0, 0, 1, 0, 0, 1.2},
		{id("robot-arm", "link-2"), "robot-arm", "Link 2", "Aluminum", "Second link", "/assets/models/robot-arm/part3.glb", 0, 0, 0, -1, 0, 0, 1.2},
		{id("robot-arm", "joint-1"), "robot-arm", "Joint 1", "Steel", "Rotation joint", "/assets/models/robot-arm/part4.glb", 0, 0, 0, 0, 1, 0, 0.9},
		{id("robot-arm", "joint-2"), "robot-arm", "Joint 2", "Steel", "Rotation joint", "/assets/models/robot-arm/part5.glb", 0, 0, 0, 0, 1, 0, 0.9},
		{id("robot-arm", "link-3"), "robot-arm", "Link 3", "Aluminum", "Third link", "/assets/models/robot-arm/part6.glb", 0, 0, 0, 0, 0, 1, 1.1},
		{id("robot-arm", "link-4"), "robot-arm", "Link 4", "Aluminum", "End link", "/assets/models/robot-arm/part7.glb", 0, 0, 0, 0, 0, 1, 1.1},
		{id("robot-arm", "end-effector"), "robot-arm", "End Effector", "Aluminum", "Tool mounting part", "/assets/models/robot-arm/part8.glb", 0, 0, 0, 0, 1, 0, 1.0},

		// Machine Vice
		{id("machine-vice", "body"), "machine-vice", "Body", "Cast iron", "Main body", "/assets/models/machine-vice/part1.glb", 0, 0, 0, 0, -1, 0, 1.1},
		{id("machine-vice", "guide"), "machine-vice", "Guide", "Cast iron", "Sliding guide", "/assets/models/machine-vice/part1-fuhrung.glb", 0, 0, 0, 1, 0, 0, 0.9},
		{id("machine-vice", "fixed-jaw"), "machine-vice", "Fixed Jaw", "Steel", "Fixed clamping jaw", "/assets/models/machine-vice/part2-feste-backe.glb", 0, 0, 0, 0, 1, 0, 1.0},
		{id("machine-vice", "movable-jaw"), "machine-vice", "Movable Jaw", "Steel", "Movable clamping jaw", "/assets/models/machine-vice/part3-lose-backe.glb", 0, 0, 0, 0, 1, 0, 1.0},
		{id("machine-vice", "spindle-base"), "machine-vice", "Spindle Base", "Steel", "Spindle support", "/assets/models/machine-vice/part4-spindelsockel.glb", 0, 0, 0, 0, 0, 1, 1.0},
		{id("machine-vice", "clamp-jaw"), "machine-vice", "Clamp Jaw", "Steel", "Clamping surface", "/assets/models/machine-vice/part5-spannbacke.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("machine-vice", "guide-rail"), "machine-vice", "Guide Rail", "Steel", "Linear guide rail", "/assets/models/machine-vice/part6-fuhrungschiene.glb", 0, 0, 0, 1, 0, 0, 1.0},
		{id("machine-vice", "spindle"), "machine-vice", "Spindle", "Steel", "Screw drive spindle", "/assets/models/machine-vice/part7-trapezspindel.glb", 0, 0, 0, 0, 0, 1, 1.2},
		{id("machine-vice", "base-plate"), "machine-vice", "Base Plate", "Cast iron", "Mounting base", "/assets/models/machine-vice/part8-grundplatte.glb", 0, 0, 0, 0, -1, 0, 1.1},
		{id("machine-vice", "pressure-sleeve"), "machine-vice", "Pressure Sleeve", "Steel", "Load distribution", "/assets/models/machine-vice/part9-druckhulse.glb", 0, 0, 0, 0, 1, 0, 0.9},
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

func (r *PostgresRepository) CreateProject(userID, title string) (*models.WorkflowProject, error) {
	row := r.db.QueryRow(
		`INSERT INTO workflow_projects (user_id, title, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id, user_id, title, COALESCE(notion_page_id, ''), created_at, updated_at`,
		userID, title,
	)
	var project models.WorkflowProject
	if err := row.Scan(&project.ID, &project.UserID, &project.Title, &project.NotionPageID, &project.CreatedAt, &project.UpdatedAt); err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *PostgresRepository) ListProjects(userID string) ([]models.WorkflowProject, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, title, COALESCE(notion_page_id, ''), created_at, updated_at
		 FROM workflow_projects WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.WorkflowProject
	for rows.Next() {
		var p models.WorkflowProject
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.NotionPageID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *PostgresRepository) GetProject(userID, projectID string) (*models.WorkflowProject, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, title, COALESCE(notion_page_id, ''), created_at, updated_at
		 FROM workflow_projects WHERE user_id = $1 AND id = $2`,
		userID, projectID,
	)
	var p models.WorkflowProject
	if err := row.Scan(&p.ID, &p.UserID, &p.Title, &p.NotionPageID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) UpdateProject(userID, projectID, title string) (*models.WorkflowProject, error) {
	row := r.db.QueryRow(
		`UPDATE workflow_projects
		 SET title = $1, updated_at = NOW()
		 WHERE user_id = $2 AND id = $3
		 RETURNING id, user_id, title, COALESCE(notion_page_id, ''), created_at, updated_at`,
		title, userID, projectID,
	)
	var p models.WorkflowProject
	if err := row.Scan(&p.ID, &p.UserID, &p.Title, &p.NotionPageID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) DeleteProject(userID, projectID string) error {
	_, err := r.db.Exec(`DELETE FROM workflow_projects WHERE user_id = $1 AND id = $2`, userID, projectID)
	return err
}

func (r *PostgresRepository) SaveWorkflowFull(userID, projectID string, nodes []models.WorkflowNode, edges []models.WorkflowEdge, checklists []models.WorkflowChecklist, attachments []models.WorkflowAttachment) error {
	project, err := r.GetProject(userID, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return sql.ErrNoRows
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	existingNotion := make(map[string]string)
	rows, err := tx.Query(`SELECT id, COALESCE(notion_page_id, '') FROM workflow_nodes WHERE project_id = $1`, projectID)
	if err != nil {
		return rollback(tx, err)
	}
	for rows.Next() {
		var id string
		var notionID string
		if err := rows.Scan(&id, &notionID); err != nil {
			rows.Close()
			return rollback(tx, err)
		}
		if notionID != "" {
			existingNotion[id] = notionID
		}
	}
	rows.Close()

	if _, err := tx.Exec(`DELETE FROM workflow_edges WHERE project_id = $1`, projectID); err != nil {
		return rollback(tx, err)
	}
	if _, err := tx.Exec(`DELETE FROM workflow_nodes WHERE project_id = $1`, projectID); err != nil {
		return rollback(tx, err)
	}

	nodeStmt, err := tx.Prepare(`INSERT INTO workflow_nodes
		(id, project_id, title, description, scheduled_date, progress, color, position_x, position_y, linked_part_id, linked_note_id, notion_page_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW(),NOW())`)
	if err != nil {
		return rollback(tx, err)
	}
	defer nodeStmt.Close()

	for _, n := range nodes {
		if n.ID == "" {
			return rollback(tx, fmt.Errorf("node id is required"))
		}
		if n.ScheduledDate == "" {
			return rollback(tx, fmt.Errorf("scheduledDate is required"))
		}
		dateVal, err := time.Parse("2006-01-02", n.ScheduledDate)
		if err != nil {
			return rollback(tx, err)
		}
		notionID := n.NotionPageID
		if strings.TrimSpace(notionID) == "" {
			if existing, ok := existingNotion[n.ID]; ok {
				notionID = existing
			}
		}
		if _, err := nodeStmt.Exec(
			n.ID, projectID, n.Title, n.Description, dateVal, n.Progress, n.Color,
			n.PositionX, n.PositionY, nullableText(n.LinkedPartID), nullableText(n.LinkedNoteID), nullableText(notionID),
		); err != nil {
			return rollback(tx, err)
		}
	}

	edgeStmt, err := tx.Prepare(`INSERT INTO workflow_edges (id, project_id, source_node_id, target_node_id) VALUES ($1,$2,$3,$4)`)
	if err != nil {
		return rollback(tx, err)
	}
	defer edgeStmt.Close()
	for _, e := range edges {
		if e.ID == "" {
			return rollback(tx, fmt.Errorf("edge id is required"))
		}
		if _, err := edgeStmt.Exec(e.ID, projectID, e.SourceID, e.TargetID); err != nil {
			return rollback(tx, err)
		}
	}

	checkStmt, err := tx.Prepare(`INSERT INTO node_checklists (id, node_id, text, done) VALUES ($1,$2,$3,$4)`)
	if err != nil {
		return rollback(tx, err)
	}
	defer checkStmt.Close()
	for _, c := range checklists {
		if c.ID == "" {
			return rollback(tx, fmt.Errorf("checklist id is required"))
		}
		if _, err := checkStmt.Exec(c.ID, c.NodeID, c.Text, c.Done); err != nil {
			return rollback(tx, err)
		}
	}

	attStmt, err := tx.Prepare(`INSERT INTO node_attachments (id, node_id, type, name, url, created_at) VALUES ($1,$2,$3,$4,$5,NOW())`)
	if err != nil {
		return rollback(tx, err)
	}
	defer attStmt.Close()
	for _, a := range attachments {
		if a.ID == "" {
			return rollback(tx, fmt.Errorf("attachment id is required"))
		}
		if _, err := attStmt.Exec(a.ID, a.NodeID, a.Type, a.Name, a.URL); err != nil {
			return rollback(tx, err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) LoadWorkflowFull(userID, projectID string) ([]models.WorkflowNode, []models.WorkflowEdge, []models.WorkflowChecklist, []models.WorkflowAttachment, error) {
	project, err := r.GetProject(userID, projectID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if project == nil {
		return nil, nil, nil, nil, sql.ErrNoRows
	}

	nodes := []models.WorkflowNode{}
	nodeRows, err := r.db.Query(
		`SELECT id, project_id, title, description, scheduled_date, progress, color, position_x, position_y,
		        COALESCE(linked_part_id, ''), COALESCE(linked_note_id::text, ''), COALESCE(notion_page_id, ''), created_at, updated_at
		 FROM workflow_nodes WHERE project_id = $1`,
		projectID,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer nodeRows.Close()
	for nodeRows.Next() {
		var n models.WorkflowNode
		var dateVal time.Time
		if err := nodeRows.Scan(&n.ID, &n.ProjectID, &n.Title, &n.Description, &dateVal, &n.Progress, &n.Color, &n.PositionX, &n.PositionY, &n.LinkedPartID, &n.LinkedNoteID, &n.NotionPageID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, nil, nil, nil, err
		}
		n.ScheduledDate = dateVal.Format("2006-01-02")
		nodes = append(nodes, n)
	}

	edges := []models.WorkflowEdge{}
	edgeRows, err := r.db.Query(`SELECT id, project_id, source_node_id, target_node_id FROM workflow_edges WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer edgeRows.Close()
	for edgeRows.Next() {
		var e models.WorkflowEdge
		if err := edgeRows.Scan(&e.ID, &e.ProjectID, &e.SourceID, &e.TargetID); err != nil {
			return nil, nil, nil, nil, err
		}
		edges = append(edges, e)
	}

	checklists := []models.WorkflowChecklist{}
	checkRows, err := r.db.Query(
		`SELECT id, node_id, text, done FROM node_checklists WHERE node_id IN (SELECT id FROM workflow_nodes WHERE project_id = $1)`,
		projectID,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer checkRows.Close()
	for checkRows.Next() {
		var c models.WorkflowChecklist
		if err := checkRows.Scan(&c.ID, &c.NodeID, &c.Text, &c.Done); err != nil {
			return nil, nil, nil, nil, err
		}
		checklists = append(checklists, c)
	}

	attachments := []models.WorkflowAttachment{}
	attRows, err := r.db.Query(
		`SELECT id, node_id, type, name, url FROM node_attachments WHERE node_id IN (SELECT id FROM workflow_nodes WHERE project_id = $1)`,
		projectID,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer attRows.Close()
	for attRows.Next() {
		var a models.WorkflowAttachment
		if err := attRows.Scan(&a.ID, &a.NodeID, &a.Type, &a.Name, &a.URL); err != nil {
			return nil, nil, nil, nil, err
		}
		attachments = append(attachments, a)
	}

	return nodes, edges, checklists, attachments, nil
}

func (r *PostgresRepository) GetNoteByPart(userID, partID string) (*models.Note, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, part_id, content, COALESCE(notion_page_id, ''), created_at, updated_at
		 FROM notes WHERE user_id = $1 AND part_id = $2`,
		userID, partID,
	)
	var n models.Note
	if err := row.Scan(&n.ID, &n.UserID, &n.PartID, &n.Content, &n.NotionPageID, &n.CreatedAt, &n.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

func (r *PostgresRepository) UpsertNote(userID, partID, content string) (*models.Note, error) {
	row := r.db.QueryRow(
		`INSERT INTO notes (user_id, part_id, content, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 ON CONFLICT (user_id, part_id)
		 DO UPDATE SET content = EXCLUDED.content, updated_at = NOW()
		 RETURNING id, user_id, part_id, content, COALESCE(notion_page_id, ''), created_at, updated_at`,
		userID, partID, content,
	)
	var n models.Note
	if err := row.Scan(&n.ID, &n.UserID, &n.PartID, &n.Content, &n.NotionPageID, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *PostgresRepository) SetNotionToken(userID, tokenEncrypted, parentPageID string) error {
	_, err := r.db.Exec(
		`INSERT INTO notion_tokens (user_id, token_encrypted, parent_page_id, created_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET token_encrypted = EXCLUDED.token_encrypted, parent_page_id = EXCLUDED.parent_page_id, created_at = NOW()`,
		userID, tokenEncrypted, parentPageID,
	)
	return err
}

func (r *PostgresRepository) GetNotionToken(userID string) (string, string, error) {
	row := r.db.QueryRow(`SELECT token_encrypted, COALESCE(parent_page_id, '') FROM notion_tokens WHERE user_id = $1`, userID)
	var token, parentPageID string
	if err := row.Scan(&token, &parentPageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}
	return token, parentPageID, nil
}

func (r *PostgresRepository) DeleteNotionToken(userID string) error {
	_, err := r.db.Exec(`DELETE FROM notion_tokens WHERE user_id = $1`, userID)
	return err
}

func nullableText(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (r *PostgresRepository) UpdateProjectNotionPageID(userID, projectID, notionPageID string) error {
	_, err := r.db.Exec(
		`UPDATE workflow_projects SET notion_page_id = $1, updated_at = NOW() WHERE user_id = $2 AND id = $3`,
		notionPageID, userID, projectID,
	)
	return err
}

func (r *PostgresRepository) UpdateNodeNotionPageID(userID, projectID, nodeID, notionPageID string) error {
	_, err := r.db.Exec(
		`UPDATE workflow_nodes AS n
		 SET notion_page_id = $1, updated_at = NOW()
		 FROM workflow_projects AS p
		 WHERE n.project_id = p.id
		   AND p.user_id = $2
		   AND n.project_id = $3
		   AND n.id = $4`,
		notionPageID, userID, projectID, nodeID,
	)
	return err
}
