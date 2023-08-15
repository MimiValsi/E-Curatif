package data

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Source struct {
        ID int `json:"-"`
        Name string `json:"name"`
        NbCuratifs int `json:"nb_curatifs"`
        CodeGMAO string `json:"code_GMAO"`
        SID int `json:"-"`

        Created time.Time `json:"-"`

        InfoLog *log.Logger `json:"-"`
        ErrorLog *log.Logger `json:"-"`
}

// Needed for Menu graph
func (jsrc *Source) SourceJSON() ([]byte, error) {
        return nil, nil        
}

// GetAllSource() fetch for each active Source, it's id, name, code_GMAO, the total
// number of info !archivé and !résolu.
func (src *Source) GetAllActive(conn *pgxpool.Conn) ([]*Source, error) {
        ctx := context.Background()

        query := `
SELECT s.id
       s.name
       s.code_GMAO
       COUNT(i.status) FILTER (WHERE i.status <> 'archivé' AND i.status <> 'résolu') 
  FROM source AS s
       LEFT JOIN info AS i
       ON i.source_id = s.id
 GROUP BY s.id
 ORDER BY name ASC
`

        rows, err := conn.Query(ctx, query)
        if err != nil {
                src.ErrorLog.Println("Could not fetch data!")
                return nil, err
        }
        defer rows.Close()

        sources := []*Source{}

        for rows.Next() {
                s := &Source{}

                args := []any{s.ID, s.Name, s.CodeGMAO, s.NbCuratifs}

                err := rows.Scan(args...)
                if err != nil {
                        return nil, err
                }
                sources = append(sources, s)
        }

        if err = rows.Err(); err != nil {
                return nil, err
        }

        return sources, nil
}

// CuratifSolved() fetch for each Source it's id, name, code_GMAO and NbCuratifs
// where it's info it's solved. Again this is primarily used for graphs.
func (src *Source) InfoSolved(conn *pgxpool.Conn) ([]*Source, error) {
        ctx := context.Background()

        query := `
SELECT s.id,
       s.name,
       s.code_GMAO,
       COUNT(i.status) FILTER (WHERE i.status = 'résolu')
  FROM source AS s
       LEFT JOIN info AS i 
       ON i.source_id = s.id
 GROUP BY s.id
 ORDER BY name ASC
`
        rows, err := conn.Query(ctx, query)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        sources := []*Source{}

        for rows.Next() {
                s := &Source{}

                args := []any{s.ID, s.Name, s.CodeGMAO, s.NbCuratifs}

                err := rows.Scan(args...)
                if err != nil {
                        return nil, err
                }

                sources = append(sources, s)
        }

        if err = rows.Err(); err != nil {
                return nil, err
        }

        return sources, nil
}

func (src *Source) Data(id int, conn pgxpool.Conn) (*Source, error) {
        ctx := context.Background()

        query := `
SELECT id, name, created
  FROM source
 WHERE id = $1
`

        s := &Source{}

        args := []any{s.ID, s.Name, s.Created}

        err := conn.QueryRow(ctx, query, id).Scan(args...)
        if err != nil {
                if errors.Is(err, pgx.ErrNoRows) {
                        return nil, ErrNoRows
                }

                return nil, err
        }
        
        return s, nil
}

// Make connexion and attempt to insert Source data to DB
// If failed then return 0 as value.
func (src *Source) Insert(name string, conn *pgx.Conn) (int, error) {
        ctx := context.Background()

        query := `
INSERT INTO source (name, created)
VALUES ($1, $2)
  RETURNING id
`

        args := []any{name, time.Now().UTC()}
        err := conn.QueryRow(ctx, query, args...).Scan(&src.ID)
        if err != nil {
                return 0, nil
        }

        return src.ID, nil
}

// Make connexion to PSQL and attempt to delete the source choosed with id.
// It only deletes if source is empty. (No info affiliated)
func (src *Source) Delete(id int, conn *pgxpool.Conn) error {
        ctx := context.Background()
        query := `
DELETE FROM source
 WHERE id = $1
`

        _, err := conn.Exec(ctx, query, id)
        if err != nil {
                return err
        }

        return nil
}

// Make connexion to PSQL and attempt to update choosen data.
func (src *Source) Update(id int, conn *pgxpool.Conn) error {
        ctx := context.Background()

        query := `
UPDATE source
    SET name = $1
 WHERE id = $2
`
        _, err := conn.Exec(ctx, query, src.Name, id)
        if err != nil {
                return err
        }

        return nil
}
