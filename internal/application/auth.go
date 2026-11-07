package application
import("context";"crypto/rand";"crypto/sha256";"encoding/base64";"encoding/hex";"time";"github.com/menerjangbadai/dashboard/internal/domain")
type Identity struct {Issuer,Subject string}
type LoginRequest struct{StateHash,BindingHash,Nonce,Verifier string;ExpiresAt time.Time}
type Session struct{Actor domain.Actor;User domain.User;AbsoluteExpiresAt time.Time}
type AuthStore interface{
 SaveLogin(context.Context,LoginRequest)error
 ConsumeLogin(context.Context,string,string)(LoginRequest,error)
 CreateSession(context.Context,Identity,string,string)error
 Session(context.Context,string)(Session,error)
 RevokeSession(context.Context,string)error
}
type IdentityProvider interface{
 AuthorizationURL(state,nonce,verifier string)string
 Exchange(context.Context,string,string,string)(Identity,error)
 Health(context.Context)error
}
func RandomToken()string{b:=make([]byte,32);if _,e:=rand.Read(b);e!=nil{panic("secure random unavailable")};return base64.RawURLEncoding.EncodeToString(b)}
func Hash(v string)string{s:=sha256.Sum256([]byte(v));return hex.EncodeToString(s[:])}
func CSRF(session string)string{return Hash("mb-csrf-v1:"+session)}
