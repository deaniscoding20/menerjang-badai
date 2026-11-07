package application
import("testing";"strings")
func TestOpaqueCredentials(t *testing.T){a:=RandomToken();b:=RandomToken();if len(a)<40||a==b{t.Fatal("unsafe entropy")};if len(Hash(a))!=64||strings.Contains(Hash(a),a){t.Fatal("hash")};if CSRF(a)==CSRF(b)||CSRF(a)==a{t.Fatal("csrf binding")}}
