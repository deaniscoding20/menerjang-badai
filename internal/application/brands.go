package application
import("context";"strings";"github.com/menerjangbadai/dashboard/internal/domain")
type BrandRepository interface {
 Membership(context.Context,domain.Actor)(domain.Actor,error)
 ListBrands(context.Context,domain.Actor)([]domain.Brand,error)
 GetBrand(context.Context,domain.Actor,string)(domain.Brand,error)
 CreateBrand(context.Context,domain.Actor,domain.Brand)(domain.Brand,error)
 UpdateBrand(context.Context,domain.Actor,domain.Brand)(domain.Brand,error)
}
type Brands struct{Repo BrandRepository}
func canCreate(role string)bool{return role=="owner" || role=="administrator" || role=="marketing_manager"}
func canWrite(role string)bool{return canCreate(role)||role=="editor"}
func(s Brands) List(ctx context.Context,a domain.Actor)([]domain.Brand,error){a,e:=s.Repo.Membership(ctx,a);if e!=nil{return nil,e};return s.Repo.ListBrands(ctx,a)}
func(s Brands) Get(ctx context.Context,a domain.Actor,id string)(domain.Brand,error){a,e:=s.Repo.Membership(ctx,a);if e!=nil{return domain.Brand{},e};return s.Repo.GetBrand(ctx,a,id)}
func(s Brands) Create(ctx context.Context,a domain.Actor,b domain.Brand)(domain.Brand,error){a,e:=s.Repo.Membership(ctx,a);if e!=nil{return domain.Brand{},e};if !canCreate(a.Role){return domain.Brand{},domain.ErrForbidden};b.Name=strings.TrimSpace(b.Name);b.Status="active";if e=b.Validate();e!=nil{return domain.Brand{},e};return s.Repo.CreateBrand(ctx,a,b)}
// Update is a full validated profile update; HTTP merges PATCH with the authorized current profile.
func(s Brands) Update(ctx context.Context,a domain.Actor,b domain.Brand)(domain.Brand,error){a,e:=s.Repo.Membership(ctx,a);if e!=nil{return domain.Brand{},e};if !canWrite(a.Role){return domain.Brand{},domain.ErrForbidden};if _,e=s.Repo.GetBrand(ctx,a,b.ID);e!=nil{return domain.Brand{},e};b.Name=strings.TrimSpace(b.Name);if e=b.Validate();e!=nil{return domain.Brand{},e};return s.Repo.UpdateBrand(ctx,a,b)}
