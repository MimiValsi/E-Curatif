package data

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CSV struct {
        Info struct {
                ID       int
                Priority int
                SourceID int
                Agent    string
                Event    string
                Created  string // Cast to date with PSQL
                Material string
                Pilot    string
                Detail   string
                Target   string
                DayDone  string
                Estimate string
                Oups     string
                Brips    string
                Ameps    string
                Status   string
        }
        Source struct  {
                ID   int
                Name string

                Created time.Time
        }
	DB       *pgxpool.Pool

	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

func (c *CSV) Verify(s string) {
        file := filepath.Ext(s)

        if file != ".csv" {
                c.ErrorLog.Println("Wrong type of file")
        } else {
                c.encoding(s)
        }
}

func (c *CSV) encoding(s string) {
	cmd, err := exec.Command("file", "-i", s).Output()
	if err != nil {
	        c.ErrorLog.Println(err)
	}

	str := []string{}
	tmp := strings.Split(string(cmd), "=")
	str = append(str, tmp...)

	tmp2 := strings.ToUpper(str[1])

	// Vérif si encodage est en UTF-8\n
	// si faux, on lance la commande de changement
	if tmp2 != "UTF-8\n" {
		cmd := exec.Command("iconv", "-f", tmp2,
			"-t", "UTF-8", s, "-o", s)
		err = cmd.Run()
		c.ErrorLog.Println(err)
	}

	c.data(s)
}

func (c *CSV) data(s string) {
	file, err := os.Open(s)
	if err != nil {
		c.ErrorLog.Println(err)
	}
	defer file.Close()

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		c.ErrorLog.Println(err)
	}

        // lines[0][0] == Source name
	source, err := c.sourceNb(lines[0][0])
	if err != nil {
		c.ErrorLog.Println(err)
	}

	for i, j := 2, 0; i < len(lines); i++ {
		line := lines[i]

		c.Info.Agent = line[j]
		c.Info.Event = line[j+1]
		c.Info.Created = line[j+2]
		c.Info.Material = line[j+3]
		c.Info.Detail = line[j+4]
		c.Info.Target = line[j+5]
                c.Info.DayDone = line[j+8]
		c.Info.Priority, _ = strconv.Atoi(line[j+9])
		c.Info.Estimate = line[j+10]
		c.Info.Oups = line[j+11]
		c.Info.Brips = line[j+12]
		c.Info.Ameps = line[j+13]
		c.Info.SourceID = source
                
		if c.Info.Status == "" && c.Info.DayDone == "" &&
			c.Info.Target == "" {
			c.Info.Status = "en attente"
		} else if c.Info.Target != "" && c.Info.DayDone == "" {
			c.Info.Status = "affecté"
		} else if c.Info.Target != "" && c.Info.DayDone != "" {
			c.Info.Status = "résolu"
		}
        }

        c.insert()
}

func (c *CSV) insert() {
        ctx := context.Background()
        query := `
INSERT INTO info
  (source_id, agent, event, material, pilote, detail, target, day_done,
    priority, estimate, oups, brips, ameps, created, status)
  VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
      (to_date($14, 'DD/MM/YYYY')), %15)
`

        ci := c.Info
        args := []any{ci.SourceID, ci.Agent, ci.Event, ci.Material, ci.Pilot,
                ci.Detail, ci.Target, ci.DayDone, ci.Priority, ci.Estimate, 
                ci.Oups, ci.Brips, ci.Ameps, ci.Created, ci.Status}

        _, err := c.DB.Exec(ctx, query, args...)
        if err != nil {
                c.ErrorLog.Println(err)
        }

        c.InfoLog.Println("data successfuly sent")
}

func (c *CSV) sourceNb(s string) (int, error) {
	ctx := context.Background()
	query := `
SELECT id
  FROM source
    WHERE name = $1
`

	var id int
	err := c.DB.QueryRow(ctx, query, s).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, ErrNoRows
		} else {
			return -1, err
		}
	}

	fmt.Printf("@ sourceNumber: id > %v \n\n", id)

	return id, nil
}
