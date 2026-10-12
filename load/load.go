package load

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/marceterrone10/shardlab/coordinator"
	"github.com/marceterrone10/shardlab/seed"
)

const (
	verifiedCustomerID = 4
	flushEvery         = 500

	insertCustomerSQL = `
		INSERT INTO clientes (id, nombre, email, estado, creado_en)
		VALUES ($1, $2, $3, $4, $5)`
	insertContactSQL = `
		INSERT INTO contactos (id, cliente_id, nombre, email, cargo, creado_en)
		VALUES ($1, $2, $3, $4, $5, $6)`
	insertOpportunitySQL = `
		INSERT INTO oportunidades (
			id, cliente_id, nombre, monto, estado, cierre_estimado, creado_en
		) VALUES ($1, $2, $3, CAST($4 AS numeric), $5, $6, $7)`
	insertActivitySQL = `
		INSERT INTO actividades (id, cliente_id, tipo, descripcion, ocurrio_en)
		VALUES ($1, $2, $3, $4, $5)`
)

// Target is one PostgreSQL database. All selects the baseline, which receives
// every customer. A shard receives only the customers routed to its name.
type Target struct {
	Name string
	DSN  string
	All  bool
}

// Result reports what the load wrote and verified.
type Result struct {
	Customers     int64
	ExpectedShard string
	Snapshots     []Snapshot
}

type database struct {
	target Target
	conn   *pgx.Conn
	tx     pgx.Tx
	batch  *pgx.Batch
}

// Load truncates every target, inserts generated customers, and checks that
// customer 4 and the table totals landed on the expected databases.
func Load(ctx context.Context, generator *seed.Generator, customerCount int64, router *coordinator.Router, targets []Target) (result Result, err error) {
	if generator == nil {
		return Result{}, fmt.Errorf("generator is required")
	}
	if customerCount < verifiedCustomerID {
		return Result{}, fmt.Errorf("customer count must be at least %d", verifiedCustomerID)
	}
	if err := validateTargets(router, targets); err != nil {
		return Result{}, err
	}

	databases, err := openDatabases(ctx, targets)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		for _, db := range databases {
			if db.tx != nil {
				_ = db.tx.Rollback(ctx)
			}
			_ = db.conn.Close(ctx)
		}
	}()

	for _, db := range databases {
		tx, beginErr := db.conn.Begin(ctx)
		if beginErr != nil {
			return Result{}, fmt.Errorf("begin %s: %w", db.target.Name, beginErr)
		}
		db.tx = tx
		if _, execErr := tx.Exec(ctx, "TRUNCATE clientes CASCADE"); execErr != nil {
			return Result{}, fmt.Errorf("truncate %s: %w", db.target.Name, execErr)
		}
	}

	for customerID := int64(1); customerID <= customerCount; customerID++ {
		bundle, customerErr := generator.Customer(customerID)
		if customerErr != nil {
			return Result{}, customerErr
		}
		destination, destinationErr := Destination(bundle, router)
		if destinationErr != nil {
			return Result{}, destinationErr
		}
		for _, db := range databases {
			if db.target.All || db.target.Name == destination {
				queueBundle(db.batch, bundle)
			}
		}
		if customerID%flushEvery == 0 {
			if flushErr := flushAll(ctx, databases); flushErr != nil {
				return Result{}, flushErr
			}
		}
	}
	if err := flushAll(ctx, databases); err != nil {
		return Result{}, err
	}
	for _, db := range databases {
		if err := syncSequences(ctx, db.tx); err != nil {
			return Result{}, fmt.Errorf("sync sequences on %s: %w", db.target.Name, err)
		}
		if err := db.tx.Commit(ctx); err != nil {
			return Result{}, fmt.Errorf("commit %s: %w", db.target.Name, err)
		}
		db.tx = nil
	}

	customerBundle, err := generator.Customer(verifiedCustomerID)
	if err != nil {
		return Result{}, err
	}
	expectedShard, err := Destination(customerBundle, router)
	if err != nil {
		return Result{}, err
	}

	snapshots := make([]Snapshot, 0, len(databases))
	var baseline Snapshot
	shards := make([]Snapshot, 0, len(databases)-1)
	for _, db := range databases {
		snapshot, readErr := readSnapshot(ctx, db.conn, db.target.Name, verifiedCustomerID)
		if readErr != nil {
			return Result{}, readErr
		}
		snapshots = append(snapshots, snapshot)
		if db.target.All {
			baseline = snapshot
			continue
		}
		shards = append(shards, snapshot)
	}
	if err := Check(baseline, shards, expectedShard, countsFor(customerBundle)); err != nil {
		return Result{}, err
	}

	return Result{
		Customers:     customerCount,
		ExpectedShard: expectedShard,
		Snapshots:     snapshots,
	}, nil
}

func validateTargets(router *coordinator.Router, targets []Target) error {
	if router == nil {
		return fmt.Errorf("router is required")
	}

	shards := router.ScatterPlan()
	seen := make(map[string]Target, len(targets))
	baselines := 0
	for _, target := range targets {
		if target.Name == "" || target.DSN == "" {
			return fmt.Errorf("target name and DSN are required")
		}
		if _, ok := seen[target.Name]; ok {
			return fmt.Errorf("duplicate target %s", target.Name)
		}
		seen[target.Name] = target
		if target.All {
			baselines++
		}
	}
	if baselines != 1 {
		return fmt.Errorf("exactly one baseline target is required")
	}

	for _, shard := range shards {
		target, ok := seen[shard.Name]
		if !ok {
			return fmt.Errorf("missing target for %s", shard.Name)
		}
		if target.All {
			return fmt.Errorf("%s cannot be the baseline and a shard", shard.Name)
		}
	}
	for _, target := range targets {
		if target.All {
			continue
		}
		known := false
		for _, shard := range shards {
			if shard.Name == target.Name {
				known = true
				break
			}
		}
		if !known {
			return fmt.Errorf("target %s is not a router shard", target.Name)
		}
	}
	return nil
}

func openDatabases(ctx context.Context, targets []Target) ([]*database, error) {
	opened := make([]*database, 0, len(targets))
	closeOpened := func() {
		for _, db := range opened {
			_ = db.conn.Close(ctx)
		}
	}

	for _, target := range targets {
		config, err := pgx.ParseConfig(target.DSN)
		if err != nil {
			closeOpened()
			return nil, fmt.Errorf("parse %s DSN: %w", target.Name, err)
		}
		config.RuntimeParams["timezone"] = "UTC"
		config.RuntimeParams["application_name"] = "shardlab-load"

		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			closeOpened()
			return nil, fmt.Errorf("connect %s: %w", target.Name, err)
		}
		opened = append(opened, &database{
			target: target,
			conn:   conn,
			batch:  &pgx.Batch{},
		})
	}
	return opened, nil
}

func queueBundle(batch *pgx.Batch, bundle seed.CustomerBundle) {
	customer := bundle.Customer
	batch.Queue(
		insertCustomerSQL,
		customer.ID,
		customer.Name,
		customer.Email,
		customer.Status,
		customer.CreatedAt,
	)
	for _, contact := range bundle.Contacts {
		batch.Queue(
			insertContactSQL,
			contact.ID,
			contact.CustomerID,
			contact.Name,
			contact.Email,
			contact.Role,
			contact.CreatedAt,
		)
	}
	for _, opportunity := range bundle.Opportunities {
		batch.Queue(
			insertOpportunitySQL,
			opportunity.ID,
			opportunity.CustomerID,
			opportunity.Name,
			opportunity.Amount,
			opportunity.Status,
			opportunity.CloseDate,
			opportunity.CreatedAt,
		)
	}
	for _, activity := range bundle.Activities {
		batch.Queue(
			insertActivitySQL,
			activity.ID,
			activity.CustomerID,
			activity.Type,
			activity.Description,
			activity.OccurredAt,
		)
	}
}

func flushAll(ctx context.Context, databases []*database) error {
	for _, db := range databases {
		if err := sendBatch(ctx, db.tx, db.batch); err != nil {
			return fmt.Errorf("insert into %s: %w", db.target.Name, err)
		}
		db.batch = &pgx.Batch{}
	}
	return nil
}

func sendBatch(ctx context.Context, tx pgx.Tx, batch *pgx.Batch) error {
	queued := batch.Len()
	if queued == 0 {
		return nil
	}

	results := tx.SendBatch(ctx, batch)
	var insertErr error
	for i := 0; i < queued; i++ {
		if _, err := results.Exec(); err != nil && insertErr == nil {
			insertErr = err
		}
	}
	closeErr := results.Close()
	if insertErr != nil {
		return insertErr
	}
	return closeErr
}

func syncSequences(ctx context.Context, tx pgx.Tx) error {
	for _, table := range []string{"clientes", "contactos", "oportunidades", "actividades"} {
		query := fmt.Sprintf(
			`SELECT setval(pg_get_serial_sequence('%s', 'id'), (SELECT MAX(id) FROM %s))`,
			table,
			table,
		)
		if _, err := tx.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func readSnapshot(ctx context.Context, conn *pgx.Conn, name string, customerID int64) (Snapshot, error) {
	snapshot := Snapshot{Name: name}
	if err := scanCounts(ctx, conn, &snapshot.Totals, `
		SELECT
			(SELECT count(*) FROM clientes),
			(SELECT count(*) FROM contactos),
			(SELECT count(*) FROM oportunidades),
			(SELECT count(*) FROM actividades)
	`); err != nil {
		return Snapshot{}, fmt.Errorf("count %s: %w", name, err)
	}
	if err := scanCounts(ctx, conn, &snapshot.Customer, `
		SELECT
			(SELECT count(*) FROM clientes WHERE id = $1),
			(SELECT count(*) FROM contactos WHERE cliente_id = $1),
			(SELECT count(*) FROM oportunidades WHERE cliente_id = $1),
			(SELECT count(*) FROM actividades WHERE cliente_id = $1)
	`, customerID); err != nil {
		return Snapshot{}, fmt.Errorf("count customer %d on %s: %w", customerID, name, err)
	}
	return snapshot, nil
}

func scanCounts(ctx context.Context, conn *pgx.Conn, counts *Counts, query string, args ...any) error {
	return conn.QueryRow(ctx, query, args...).Scan(
		&counts.Customers,
		&counts.Contacts,
		&counts.Opportunities,
		&counts.Activities,
	)
}

func countsFor(bundle seed.CustomerBundle) Counts {
	return Counts{
		Customers:     1,
		Contacts:      int64(len(bundle.Contacts)),
		Opportunities: int64(len(bundle.Opportunities)),
		Activities:    int64(len(bundle.Activities)),
	}
}
