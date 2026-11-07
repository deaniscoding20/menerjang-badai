package domain

import "testing"

func TestBrandValidation(t *testing.T) {
 for _, b := range []Brand{{Name:""},{Name:"ok", Website:"javascript:alert(1)"},{Name:"ok", Website:"https://u:p@example.com"},{Name:"ok",Status:"deleted"}} { if b.Validate()==nil { t.Fatalf("accepted invalid brand: %+v",b) } }
 if err:= (Brand{Name:"A",Website:"https://example.com",Status:"active"}).Validate(); err!=nil { t.Fatal(err) }
}
func TestRegistryIsStableAndDisabled(t *testing.T) {
 a:=Agents(); b:=Agents(); if len(a)!=6 {t.Fatalf("want 6 got %d",len(a))}; seen:=map[string]bool{}
 for i,v:=range a {if v.ID!=b[i].ID || seen[v.ID] || v.Enabled || v.Status!="unknown" || v.Reason=="" { t.Fatalf("invalid registry %+v",v) }; seen[v.ID]=true}
 if a[0].Name!="Claudia" || a[4].Telegram!="@MB_sabrina_bot" {t.Fatal(a)}
}
