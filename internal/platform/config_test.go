package platform
import("os";"path/filepath";"testing";"strings")
func TestSecretsAreFileOnlyAndPrivate(t *testing.T){p:=filepath.Join(t.TempDir(),"dsn");os.WriteFile(p,[]byte("postgres://example"),0644);if _,e:=ReadSecret(p);e==nil{t.Fatal("world readable secret accepted")};os.Chmod(p,0600);v,e:=ReadSecret(p);if e!=nil||v!="postgres://example"{t.Fatal(e)};t.Setenv("MB_DATABASE_URL","forbidden");if _,e=DatabaseURL();e==nil{t.Fatal("raw secret env accepted")}}
func TestConfigStrictRuntime(t *testing.T){t.Setenv("MB_PUBLIC_URL","http://app.test");t.Setenv("MB_COMMIT_SHA",strings.Repeat("a",40));if _,e:=Load("");e==nil{t.Fatal("insecure origin accepted")}}
