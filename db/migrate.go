package db
import("context";"crypto/sha256";"embed";"fmt";"sort";"github.com/jackc/pgx/v5/pgxpool")
//go:embed migrations/*.sql
var migrations embed.FS
func Migrate(ctx context.Context,p *pgxpool.Pool)error{
 tx,e:=p.Begin(ctx);if e!=nil{return e};defer tx.Rollback(ctx)
 if _,e=tx.Exec(ctx,"SELECT pg_advisory_xact_lock(819202301)");e!=nil{return e}
 if _,e=tx.Exec(ctx,"CREATE TABLE IF NOT EXISTS schema_migrations(version text PRIMARY KEY,checksum text NOT NULL,applied_at timestamptz NOT NULL DEFAULT now())");e!=nil{return e}
 fs,e:=migrations.ReadDir("migrations");if e!=nil{return e};sort.Slice(fs,func(i,j int)bool{return fs[i].Name()<fs[j].Name()})
 for _,f:=range fs{b,e:=migrations.ReadFile("migrations/"+f.Name());if e!=nil{return e};sum:=fmt.Sprintf("%x",sha256.Sum256(b));var old string;e=tx.QueryRow(ctx,"SELECT checksum FROM schema_migrations WHERE version=$1",f.Name()).Scan(&old);if e==nil{if old!=sum{return fmt.Errorf("migration checksum mismatch: %s",f.Name())};continue};if e.Error()!="no rows in result set"{return e};if _,e=tx.Exec(ctx,string(b));e!=nil{return fmt.Errorf("migration %s: %w",f.Name(),e)};if _,e=tx.Exec(ctx,"INSERT INTO schema_migrations(version,checksum) VALUES($1,$2)",f.Name(),sum);e!=nil{return e}}
 return tx.Commit(ctx)
}
