package db

import (
	"database/sql"
	"fmt"
	"time"

	"go-browser-inventory/internal/browsers"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite connection
type DB struct {
	conn *sql.DB
}

// NewDB initializes a new SQLite database connection
func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	browsersList := []string{"Chrome", "Edge", "Firefox"}
	for _, browser := range browsersList {
		// Use composite primary key (id, profile, version)
		query := fmt.Sprintf(`
            CREATE TABLE IF NOT EXISTS %s_extensions (
                id TEXT,
                name TEXT NOT NULL,
                browser TEXT NOT NULL,
                version TEXT NOT NULL,
                enabled INTEGER NOT NULL,
                profile TEXT,
                timestamp INTEGER NOT NULL,
                PRIMARY KEY (id, profile, version)
            )`, browser)
		if _, err := conn.Exec(query); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create table %s_extensions: %w", browser, err)
		}
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.conn.Close()
}

// GetExtensions retrieves cached extensions if fresh, or returns nil if stale/empty
func (d *DB) GetExtensions(browser string) ([]browsers.Extension, error) {
	// Check the latest timestamp
	query := fmt.Sprintf("SELECT timestamp FROM %s_extensions ORDER BY timestamp DESC LIMIT 1", browser)
	row := d.conn.QueryRow(query)

	var ts int64
	err := row.Scan(&ts)
	if err == sql.ErrNoRows {
		return nil, nil // No data yet
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query %s_extensions timestamp: %w", browser, err)
	}

	if time.Since(time.Unix(ts, 0)) > 30*time.Minute {
		return nil, nil // Cache is stale
	}

	// Fetch all extensions with the latest timestamp
	query = fmt.Sprintf("SELECT id, name, browser, version, enabled, profile FROM %s_extensions WHERE timestamp = ?", browser)
	rows, err := d.conn.Query(query, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch extensions: %w", err)
	}
	defer rows.Close()

	var extensions []browsers.Extension
	for rows.Next() {
		var e browsers.Extension
		var enabledInt int
		if err := rows.Scan(&e.ID, &e.Name, &e.Browser, &e.Version, &enabledInt, &e.Profile); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		e.Enabled = enabledInt != 0
		extensions = append(extensions, e)
	}

	return extensions, nil
}

// UpdateExtensions updates the extension table for a browser
func (d *DB) UpdateExtensions(browser string, extensions []browsers.Extension) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Clear old data
	query := fmt.Sprintf("DELETE FROM %s_extensions", browser)
	if _, err := tx.Exec(query); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear %s_extensions: %w", browser, err)
	}

	// Insert new data with composite key
	query = fmt.Sprintf("INSERT INTO %s_extensions (id, name, browser, version, enabled, profile, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)", browser)
	now := time.Now().Unix()
	for _, ext := range extensions {
		enabledInt := 0
		if ext.Enabled {
			enabledInt = 1
		}
		if _, err := tx.Exec(query, ext.ID, ext.Name, ext.Browser, ext.Version, enabledInt, ext.Profile, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert extension: %w", err)
		}
	}

	return tx.Commit()
}
