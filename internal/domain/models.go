package domain

import (
 "errors"
 "net/url"
 "strings"
 "unicode/utf8"
)
var ErrInvalid=errors.New("invalid input")
var ErrForbidden=errors.New("forbidden")
var ErrNotFound=errors.New("not found")
var ErrUnauthorized=errors.New("unauthorized")
type Brand struct { ID string `json:"id"`; Name string `json:"name"`; Description string `json:"description"`; Website string `json:"website"`; Status string `json:"status"` }
func (b Brand) Validate() error {
 if strings.TrimSpace(b.Name)=="" || utf8.RuneCountInString(b.Name)>120 || utf8.RuneCountInString(b.Description)>4000 || len(b.Website)>2048 {return ErrInvalid}
 if b.Status!="" && b.Status!="active" && b.Status!="archived" {return ErrInvalid}
 if b.Website!="" {u,e:=url.Parse(b.Website); if e!=nil || (u.Scheme!="https" && u.Scheme!="http") || u.Hostname()=="" || u.User!=nil {return ErrInvalid}}
 return nil
}
type Actor struct { UserID string; OrgID string; Role string; Revision int64 }
type User struct { ID string `json:"id"`; DisplayName string `json:"display_name"`; Role string `json:"role"` }
type Agent struct { ID string `json:"id"`; Name string `json:"name"`; Role string `json:"role"`; Telegram string `json:"telegram"`; Profile string `json:"profile"`; Status string `json:"status"`; Enabled bool `json:"enabled"`; Reason string `json:"reason"` }
func Agents() []Agent {
 names:=[]string{"Claudia","Jordan","Kobe","Curry","Sabrina","Kyrie"}
 roles:=[]string{"Assistant Brand Manager & Coordinator","Marketing Strategist","Creative Lead","Market Researcher","Social Media & Copywriter","Creative Designer"}
 ids:=[]string{"ec0b8d12-e31e-43a6-9dc3-82a2ee8a1d01","7e66ec1f-afec-4bad-b0eb-b24e05e5de02","07a4226e-5eba-4d2a-95a5-78389ed40603","9b58c0d9-8300-4a10-883d-04bdba2e1b04","cc9f1488-e14c-4203-8f01-9fc9b20ae505","0b1e02ba-94ae-443d-a29c-c84b95019d06"}
 a:=make([]Agent,6); for i,n:=range names {a[i]=Agent{ID:ids[i],Name:n,Role:roles[i],Telegram:"@MB_"+strings.ToLower(n)+"_bot",Profile:strings.ToLower(n),Status:"unknown",Reason:"Execution and native profile routing are unverified; dispatch is disabled."}}
 return a
}
