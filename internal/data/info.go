package data

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Info struct {
        ID int `json:"-"`
        Priority int `json:"priority,omitempty"`
	SourceID   int    `json:"-"` // foreign key en référence au PK de source
	Counter    int    `json:"counter,omitempty"`
	Agent      string `json:"agent,omitempty"`
	Material   string `json:"material,omitempty"`
	Target     string `json:"target,omitempty"`
	Rte        string `json:"rte,omitempty"`
	Detail     string `json:"detail,omitempty"`
	Estimate   string `json:"estimate,omitempty"`
	Brips      string `json:"brips,omitempty"`
	Oups       string `json:"oups,omitempty"`
	Ameps      string `json:"ameps,omitempty"`
	Ais        string `json:"ais,omitempty"`
	Status     string `json:"status,omitempty"`
	Event      string `json:"event,omitempty"`
	Doneby     string `json:"doneby,omitempty"`
	DayDone    string `json:"dayDone,omitempty"`

	ZeroTime time.Time `json:"-"`
	Created  time.Time `json:"-"`
	Updated  time.Time `json:"-"`

        InfoLog *log.Logger
        ErrorLog *log.Logger
}

func (i *Info) ActiveInfo(conn *pgxpool.Conn) ([]*Info, error) {
        ctx := context.Background()

        query := `
SELECT i.material, 
       i.detail
  FROM info AS i
 WHERE status <> 'résolu' AND
       status <> 'archivé' 
`

        rows, err := conn.Query(ctx, query)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        infos := []*Info{}

        for rows.Next() {
                // i := &Info{}

                err := rows.Scan(&i.Material, &i.Detail)
                if err != nil {
                        return nil, err
                }

                infos = append(infos, i)
        }

        if err = rows.Err(); err != nil {
                return nil, err
        }

        return infos, nil
} 

// Send data to DB 
func (i *Info) Insert(id int, conn *pgxpool.Conn) (int, error) {
        ctx := context.Background()

        query := `
INSERT INTO info (source_id, agent, material, detail, 
	   	  event, priority, oups, ameps,
       		  brips, rte, ais, estimate, 
		  target, status, doneby, created)
VALUES ($1,  $2,  $3,  $4, 
	$5,  $6,  $7,  $8, 
	$9,  $10, $11, $12, 
	$13, $14, $15, $16)
  RETURNING id;
        `

        args := []any{id, i.Agent, i.Material, i.Detail, i.Event, i.Priority,
                      i.Oups, i.Ameps, i.Brips, i.Rte, i.Ais, i.Estimate,
                      i.Target, i.Status, i.Doneby, time.Now().UTC()}

        err := conn.QueryRow(ctx, query, args...).Scan(&i.ID)
        if err != nil {
                i.InfoLog.Println("Could not insert info data!")
                return 0, nil
        }

        return i.ID, nil
}

// Fetch Info data so it can be displayed @ infoView page.
func (i *Info) Data(id int, conn *pgxpool.Conn) (*Info, error) {
        ctx := context.Background()

        query := `
SELECT id, agent, material, priority, 
       rte, detail, estimate, brips,
       oups, ameps, ais, source_id, 
       created, updated, status, event, target, doneby
  FROM info
 WHERE id = $1 AND 
 status <> 'résolu'
        `

        var rte, ameps, ais, brips, oups, estimate, target, doneby *string
        var updated *time.Time

        scan := []any{&i.ID, &i.Agent, &i.Material, &i.Priority, &i.Rte,
                      &i.Detail, &i.Estimate, &i.Brips, &i.Oups, &i.Ameps,
                      &i.Ais, &i.SourceID, &i.Created, &i.Updated, &i.Status,
                      &i.Event, &i.Target, &i.Doneby}

        err := conn.QueryRow(ctx, query).Scan(scan...)
        if err != nil {
                if errors.Is(err, pgx.ErrNoRows) {
                        return nil, ErrNoRows
                }

                return nil, err
        }

        // Used inside a template so it can be rendered as "empty time or 0 time"
        i.ZeroTime = time.Date(0001, time.January, 1, 0, 0, 0, 0, time.UTC)

        // PSQL returns NULL if empty row, Golang doesn't supporty NULL value
        // but nil value. Then we cast to a pointer.

        if target != nil {
                i.Target = *target
        }

        if rte != nil {
                i.Rte = *rte
        }

        if ameps != nil {
                i.Ameps = *ameps
        }

        if ais != nil {
                i.Ais = *ais
        }

        if brips != nil {
                i.Brips = *brips
        }

        if oups != nil {
                i.Oups = *oups
        }

        if updated != nil {
                i.Updated = *updated
        }

        if estimate != nil {
                i.Estimate = *estimate
        }

        if doneby != nil {
                i.Doneby = *doneby
        }

        return i, nil
}

// List() fetch every info associated to it's Source so it can be displayed by
// infoView handler
func (i *Info) List(id int, conn *pgxpool.Conn) ([]*Info, error) {
        ctx := context.Background()

        query := `
SELECT id, material, created, status, source_id, priority
  FROM info
 WHERE source_id = $1 AND status <> 'archivé'
 ORDER BY priority ASC
`

        rows, err := conn.Query(ctx, query, id)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        infos := []*Info{}

        for rows.Next() {
                args := []any{&i.ID, &i.Material, &i.Created, &i.Status,
                              &i.SourceID, &i.Priority}

                err = rows.Scan(args...)
                if err != nil {
                        return nil, err
                }

                infos = append(infos, i)
        }

        if err = rows.Err(); err != nil {
                return nil, err
        }

        return infos, nil
}

func (i *Info) Delete(id int, conn *pgxpool.Conn) error {
        ctx := context.Background()

        query := `
DELETE FROM info 
 WHERE id = $1
`
        _, err := conn.Exec(ctx, query, id)
        if err != nil {
                return err
        }

        return nil

}

func (i *Info) Update(id int, conn *pgxpool.Conn) error {
        ctx := context.Background()

        query := `
UPDATE info
   SET agent = $1, material = $2, priority = $3, target = $4, rte = $5,
       detail = $6, estimate = $7, brips = $8, oups = $9, ameps = $10,
       ais = $11, updated = $12, status = $13, event = $14, doneby = $15
 WHERE id = $16
`
        args := []any{i.Agent, i.Material, i.Priority,i.Target, i.Rte,
                      i.Detail, i.Estimate, i.Brips, i.Oups, i.Ameps,
                      i.Ais, time.Now().UTC(), i.Status, i.Event, i.Doneby, id}

        _, err := conn.Exec(ctx, query, args...) 
        if err != nil {
                return err
        }

        return nil
}

func (i *Info) Test(id int, conn *pgxpool.Conn) error {
        ctx := context.Background()

        query := `
UPDATE info
   SET material = $1, updated = $2
 WHERE id = $3
`
        args := []any{i.Material, time.Now().UTC(), id}

        _, err := conn.Exec(ctx, query, args...)
        if err != nil {
                return err
        }

        return nil
}
