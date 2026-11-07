package application
import("context";"testing";"github.com/menerjangbadai/dashboard/internal/domain")
type denyRepo struct{ BrandRepository; calls int }
func(r *denyRepo) Membership(context.Context,domain.Actor)(domain.Actor,error){return domain.Actor{},domain.ErrForbidden}
func TestUseCaseRejectsWithoutMembership(t *testing.T){r:=&denyRepo{};s:=Brands{Repo:r};if _,e:=s.List(context.Background(),domain.Actor{}); e!=domain.ErrForbidden {t.Fatal(e)}}
func TestCreatePermissionDefaultDeny(t *testing.T){for _,role:=range []string{"","viewer","analyst","approver","editor","superadmin"}{if canCreate(role){t.Fatal(role)}}; if !canCreate("owner"){t.Fatal("owner denied")}}
